// Package client provides the low level assembly of the Notifiarr client application.
// This package orchestrates reading in of configuration, parsing cli flags, actioning
// those cli flags, setting up logging, and finally the starting of internal service
// routines for the webserver, plex sessions, snapshots, service checks and others.
// This package sets everything up for the client application.
package client

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/cooldown"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/logs/share"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/gorilla/securecookie"
	"github.com/hako/durafmt"
	flag "github.com/spf13/pflag"
	mulery "golift.io/mulery/client"
	"golift.io/version"
)

// Client stores all the running data.
type Client struct {
	*logs.Logger
	plexTimer  *cooldown.Timer
	Flags      *configfile.Flags
	Config     *configfile.Config
	server     *http.Server
	sigkil     chan os.Signal
	sighup     chan os.Signal
	reload     chan customReload
	website    *website.Server
	clientinfo *clientinfo.Config
	triggers   *triggers.Actions
	cookies    *securecookie.SecureCookie
	template   *template.Template
	tunnel     *mulery.Client
	webauth    bool
	noauth     bool
	authHeader string
	reloading  bool
	// this locks anything that may be updated while running.
	// at least "UIPassword" and "reloading" as of its creation.
	sync.RWMutex
}

type customReload struct {
	event website.EventType
	msg   string
}

// Errors returned by this package.
var (
	ErrNilAPIKey = fmt.Errorf("API key may not be empty: set a key in config file, OR with environment variable")
)

// newDefaults returns a new Client pointer with default settings.
func newDefaults() *Client {
	logger := logs.New() // This persists throughout the app.

	return &Client{
		sigkil:    make(chan os.Signal, 1),
		sighup:    make(chan os.Signal, 1),
		reload:    make(chan customReload, 1),
		Logger:    logger,
		plexTimer: cooldown.NewTimer(false, time.Hour),
		Config:    configfile.NewConfig(logger),
		Flags: &configfile.Flags{
			FlagSet:    flag.NewFlagSet(mnd.DefaultName, flag.ExitOnError),
			ConfigFile: os.Getenv(mnd.DefaultEnvPrefix + "_CONFIG_FILE"),
			EnvPrefix:  mnd.DefaultEnvPrefix,
		},
		cookies: securecookie.New(securecookie.GenerateRandomKey(mnd.Bits64), securecookie.GenerateRandomKey(mnd.Bits32)),
	}
}

// Start runs the app.
func Start() error {
	client := newDefaults()
	client.Flags.ParseArgs(os.Args[1:])

	ctx := context.Background()

	//nolint:forbidigo,wrapcheck
	switch {
	case client.Flags.LongVerReq: // print version and exit.
		_, err := fmt.Println(version.Print(client.Flags.Name()))
		return err
	case client.Flags.VerReq: // print version and exit.
		_, err := fmt.Println(client.Flags.Name() + " " + version.Version + "-" + version.Revision)
		return err
	case client.Flags.PSlist: // print process list and exit.
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		return printProcessList(ctx)
	case client.Flags.Fortune: // print fortune and exit.
		_, err := fmt.Println(Fortune())
		return err
	case client.Flags.Curl != "": // curl a URL and exit.
		return curlURL(client.Flags.Curl, client.Flags.Headers)
	default:
		return client.start(ctx)
	}
}

func (c *Client) start(ctx context.Context) error { //nolint:cyclop
	msg, newPassword, err := c.loadConfiguration(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	switch {
	case c.Flags.AptHook:
		return c.handleAptHook(ctx) // ignore config errors. *cringe*
	case c.Flags.Reset:
		if err != nil && !strings.Contains(err.Error(), "ip:port") {
			return fmt.Errorf("cannot reset admin password, got error reading configuration file: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		return c.resetAdminPassword(ctx)
	case c.Flags.Write != "" && (err == nil || strings.Contains(err.Error(), "ip:port")):
		c.Printf("==> %s", msg)

		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		return c.forceWriteWithExit(ctx, c.Flags.Write)
	case err != nil:
		return fmt.Errorf("%s: %w", msg, err)
	case c.Flags.Restart:
		return nil
	case c.Config.APIKey == "":
		return fmt.Errorf("%s: %w %s_API_KEY", msg, ErrNilAPIKey, c.Flags.EnvPrefix)
	}

	c.Logger.SetupLogging(c.Config.LogConfig)
	c.Printf(" %s %s v%s-%s Starting! [PID: %v] %s",
		mnd.TodaysEmoji(), c.Flags.Name(), version.Version, version.Revision, os.Getpid(),
		version.Started.Format("Monday, January 2, 2006 @ 3:04:05 PM MST -0700"))
	c.Printf("==> %s", msg)
	c.printUpdateMessage()

	if err := c.loadAssetsTemplates(ctx); err != nil {
		return err
	}

	clientInfo := c.configureServices(ctx)

	if newPassword != "" {
		// If newPassword is set it means we need to write out a new config file for a new installation. Do that now.
		c.makeNewConfigFile(ctx, newPassword)
	} else if c.Config.AutoUpdate != "" {
		// do not run updater if there's a brand new config file.
		go c.AutoWatchUpdate(ctx)
	}

	if ui.HasGUI() {
		// This starts the web server and calls os.Exit() when done.
		c.startTray(ctx, cancel, clientInfo)
		return nil
	}

	return c.Exit(ctx, cancel)
}

func (c *Client) makeNewConfigFile(ctx context.Context, newPassword string) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	_, _ = c.Config.Write(ctx, c.Flags.ConfigFile, false)
	_ = ui.OpenFile(c.Flags.ConfigFile)
	_, _ = ui.Warning(mnd.Title, "A new configuration file was created @ "+
		c.Flags.ConfigFile+" - it should open in a text editor. "+
		"Please edit the file and reload this application using the tray menu. "+
		"Your Web UI password was set to "+newPassword+
		" and was also printed in the log file '"+c.Config.LogFile+"' and/or app ouptput.")
}

// loadConfiguration brings in, and sometimes creates, the initial running configuration.
func (c *Client) loadConfiguration(ctx context.Context) (msg string, newPassword string, err error) {
	// Find or write a config file. This does not parse it.
	// A config file is only written when none is found on Windows, macOS (GUI App only), or Docker.
	// And in the case of Docker, only if `/config` is a mounted volume.
	write := (!c.Flags.Restart && ui.HasGUI()) || mnd.IsDocker
	c.Flags.ConfigFile, newPassword, msg = c.Config.FindAndReturn(ctx, c.Flags.ConfigFile, write)

	if c.Flags.Restart {
		return msg, newPassword, update.Restart(&update.Command{ //nolint:wrapcheck
			Path: os.Args[0],
			Args: []string{"--updated", "--config", c.Flags.ConfigFile},
		})
	}

	// Parse the config file and environment variables.
	c.website, c.triggers, err = c.Config.Get(c.Flags, c.Logger)
	if err != nil {
		return msg, newPassword, fmt.Errorf("getting config: %w", err)
	}

	c.clientinfo = c.triggers.Timers.CIC

	return msg, newPassword, nil
}

// Load configuration from the website.
func (c *Client) loadSiteConfig(ctx context.Context) *clientinfo.ClientInfo {
	clientInfo, err := c.clientinfo.SaveClientInfo(ctx, true)
	if err != nil || clientInfo == nil {
		if errors.Is(err, website.ErrInvalidAPIKey) {
			c.ErrorfNoShare("==> Problem validating API key: %v", err)
			c.ErrorfNoShare("==> NOTICE! No Further requests will be sent to the website until you reload with a valid API Key!")
		} else {
			c.Printf("==> [WARNING] Problem validating API key: %v, info: %s", err, clientInfo)
		}

		return nil
	}

	// Snapshot is a bit complicated because config-file data (plugins) merges with site-data (snapshot config).
	clientInfo.Actions.Snapshot.Plugins = c.Config.Snapshot.Plugins
	c.Config.Snapshot = &clientInfo.Actions.Snapshot
	c.triggers.Timers.Snapshot = c.Config.Snapshot
	c.Config.Services.Plugins = c.Config.Snapshot.Plugins

	return clientInfo
}

// configureServices is called on startup and on reload, so be careful what goes in here.
func (c *Client) configureServices(ctx context.Context) *clientinfo.ClientInfo {
	c.website.Start(ctx)

	clientInfo := c.loadSiteConfig(ctx)
	if clientInfo != nil && !clientInfo.User.StopLogs {
		share.Setup(c.website)
	}

	c.configureServicesPlex(ctx)
	c.Config.Snapshot.Validate()
	c.PrintStartupInfo(ctx, clientInfo)
	c.triggers.Start(ctx, c.sighup)
	c.Config.Services.Start(ctx)

	return clientInfo
}

func (c *Client) configureServicesPlex(ctx context.Context) {
	if !c.Config.Plex.Enabled() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, c.Config.Plex.Timeout.Duration)
	defer cancel()

	if _, err := c.Config.Plex.GetInfo(ctx); err != nil {
		c.Errorf("=> Getting Plex Media Server info (check url and token): %v", err)
	}
}

func (c *Client) triggerConfigReload(event website.EventType, source string) {
	c.reload <- customReload{event: event, msg: source}
}

// Exit stops the web server and logs our exit messages. Start() calls this.
func (c *Client) Exit(ctx context.Context, cancel context.CancelFunc) error {
	defer func() {
		defer c.CapturePanic()
		cancel()
		//nolint:gomnd
		c.Print(" ❌ Good bye! Uptime:", durafmt.Parse(time.Since(version.Started).Round(time.Second)).LimitFirstN(3))
	}()

	c.StartWebServer(ctx)
	c.setSignals()

	// For non-GUI systems, this is where the main go routine stops (and waits).
	for {
		select {
		case data := <-c.reload:
			if err := c.reloadConfiguration(ctx, data.event, data.msg); err != nil {
				return err
			}
		case sigc := <-c.sigkil:
			c.Printf("[%s] Need help? %s\n=====> Exiting! Caught Signal: %v", c.Flags.Name(), mnd.HelpLink, sigc)
			return c.stop(ctx, website.EventSignal)
		case sigc := <-c.sighup:
			if err := c.checkReloadSignal(ctx, sigc); err != nil {
				return err // reloadConfiguration()
			}
		}
	}
}

// reloadConfiguration is called from a menu tray item or when a HUP signal is received.
// Re-reads the configuration file and stops/starts all the internal routines.
// Also closes and re-opens all log files. Any errors cause the application to exit.
func (c *Client) reloadConfiguration(ctx context.Context, event website.EventType, source string) error {
	c.Printf("==> Reloading Configuration (%s): %s", event, source)

	err := c.stop(ctx, event)
	if err != nil {
		return fmt.Errorf("stopping web server: %w", err)
	}

	// start over.
	c.Config = configfile.NewConfig(c.Logger)
	if c.website, c.triggers, err = c.Config.Get(c.Flags, c.Logger); err != nil {
		return fmt.Errorf("getting configuration: %w", err)
	}

	c.clientinfo = c.triggers.Timers.CIC

	if errs := c.Logger.Close(); len(errs) > 0 {
		return fmt.Errorf("closing logger(s): %w", errs[0])
	}

	defer c.StartWebServer(ctx)

	c.Logger.SetupLogging(c.Config.LogConfig)
	clientInfo := c.configureServices(ctx)
	c.setupMenus(clientInfo)
	c.Printf(" 🌀 %s v%s-%s Configuration Reloaded! Config File: %s",
		c.Flags.Name(), version.Version, version.Revision, c.Flags.ConfigFile)

	if err = ui.Notify("Configuration Reloaded! Config File: %s", c.Flags.ConfigFile); err != nil {
		c.Errorf("Creating Toast Notification: %v", err)
	}

	// This doesn't need to lock because web server is not running.
	c.reloading = false // We're done.

	return nil
}

// stop is called from at least two different exit points and on reload.
func (c *Client) stop(ctx context.Context, event website.EventType) error {
	defer func() {
		defer c.CapturePanic()
		c.triggers.Stop(event)
		c.Config.Services.Stop()
		c.website.Stop()
		c.Print("==> All systems powered down!")
	}()

	return c.StopWebServer(ctx)
}
