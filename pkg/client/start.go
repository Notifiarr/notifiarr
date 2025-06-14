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
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/cooldown"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/logs/share"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/triggers"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/gorilla/securecookie"
	flag "github.com/spf13/pflag"
	mulery "golift.io/mulery/client"
	"golift.io/version"
)

// Client stores all the running data.
type Client struct {
	apps       *apps.Apps
	plexTimer  *cooldown.Timer
	Flags      *configfile.Flags
	Config     *configfile.Config
	Input      *configfile.Config
	server     *http.Server
	sigkil     chan os.Signal
	sighup     chan os.Signal
	reload     chan customReload
	triggers   *triggers.Actions
	Services   *services.Services
	cookies    *securecookie.SecureCookie
	template   *template.Template
	tunnel     *mulery.Client
	webauth    bool
	noauth     bool
	authHeader string
	reloading  bool
	newUI      bool
	allow      configfile.AllowedIPs `json:"-" toml:"-" xml:"-" yaml:"-"`

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
	ErrUnsupport = errors.New("this feature is not supported on this platform")
	ErrNilAPIKey = errors.New("API key may not be empty: set a key in config file, OR with environment variable")
)

// newDefaults returns a new Client pointer with default settings.
func newDefaults() *Client {
	mnd.Log = logs.Log

	return &Client{
		sigkil:    make(chan os.Signal, 1),
		sighup:    make(chan os.Signal, 1),
		reload:    make(chan customReload, 1),
		plexTimer: cooldown.NewTimer(false, time.Hour),
		Config:    configfile.NewConfig(),
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
		_, err := fmt.Println(mnd.PrintVersionInfo(client.Flags.Name()))
		return err
	case client.Flags.VerReq: // print version and exit.
		_, err := fmt.Println(client.Flags.Name() + " " + version.Version + "-" + version.Revision)
		return err
	case client.Flags.PSlist: // print process list and exit.
		ctx, cancel := context.WithTimeout(ctx, mnd.DefaultTimeout)
		defer cancel()

		return printProcessList(ctx)
	case client.Flags.Fortune: // print fortune and exit.
		_, err := fmt.Println(Fortune())
		return err
	case client.Flags.Curl != "": // curl a URL and exit.
		return curlURL(client.Flags.Curl, client.Flags.Headers)
	default:
		return client.checkFlags(ctx)
	}
}

func (c *Client) checkFlags(ctx context.Context) error { //nolint:cyclop
	c.newUI = os.Getenv(c.Flags.EnvPrefix+"_NEW_UI") == mnd.True
	msgs, newPassword, err := c.loadConfiguration(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	switch {
	case c.Flags.AptHook:
		return c.handleAptHook(ctx) // ignore config errors. *cringe*
	case c.Flags.Reset:
		if err != nil && !strings.Contains(err.Error(), "ip:port") {
			return fmt.Errorf("cannot reset admin password, got error reading configuration file: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, mnd.DefaultTimeout)
		defer cancel()

		return c.resetAdminPassword(ctx)
	case c.Flags.Write != "" && (err == nil || strings.Contains(err.Error(), "ip:port")):
		for _, msg := range msgs {
			logs.Log.Printf("==> %s", msg)
		}

		ctx, cancel := context.WithTimeout(ctx, mnd.DefaultTimeout)
		defer cancel()

		return c.forceWriteWithExit(ctx, c.Flags.Write)
	case err != nil:
		return fmt.Errorf("messages: %q, error: %w", msgs, err)
	case c.Flags.Restart:
		return nil
	case c.Config.APIKey == "":
		return fmt.Errorf("messages: %q, %w %s_API_KEY", msgs, ErrNilAPIKey, c.Flags.EnvPrefix)
	default:
		return c.start(ctx, msgs, newPassword)
	}
}

func (c *Client) start(ctx context.Context, msgs []string, newPassword string) error {
	logs.Log.SetupLogging(c.Config.LogConfig)
	logs.Log.Printf(" %s %s v%s-%s Starting! [PID: %v, UID: %d, GID: %d] %s",
		mnd.TodaysEmoji(), mnd.Title, version.Version, version.Revision,
		os.Getpid(), os.Getuid(), os.Getgid(),
		version.Started.Format("Mon, Jan 2, 2006 @ 3:04:05 PM MST -0700"))

	for _, msg := range msgs {
		logs.Log.Printf("==> %s", msg)
	}

	if c.Flags.Updated {
		go ui.Toast("%s updated to v%s-%s", mnd.Title, version.Version, version.Revision) //nolint:errcheck
	}

	if err := c.loadAssetsTemplates(ctx); err != nil {
		return err
	}

	if newPassword != "" {
		// If newPassword is set it means we need to write out a new config file for a new installation. Do that now.
		c.makeNewConfigFile(ctx, newPassword)
	}

	clientInfo, reload := c.configureServices(ctx)

	if ui.HasGUI() {
		// This starts the web server and calls os.Exit() when done.
		c.startTray(ctx, clientInfo, reload)
		return nil
	}

	return c.Exit(ctx, reload)
}

func (c *Client) makeNewConfigFile(ctx context.Context, newPassword string) {
	ctx, cancel := context.WithTimeout(ctx, mnd.DefaultTimeout)
	defer cancel()

	c.Config.APIKey, _, _ = ui.Entry("Enter 'All' API Key from notifiarr.com", "api-key-from-notifiarr.com")
	if website.Site.ValidAPIKey() != nil {
		c.Config.APIKey = "api-key-from-notifiarr.com"
	} else {
		c.Input.APIKey = c.Config.APIKey
	}

	// write new config file to temporary path.
	destFile := filepath.Join(filepath.Dir(c.Flags.ConfigFile), "_tmpConfig")
	if _, err := c.Config.Write(ctx, destFile, true); err != nil { // write our config file template.
		logs.Log.Errorf("writing new (temporary) config file: %v", err)
	}

	// move new config file to existing config file.
	if err := os.Rename(destFile, c.Flags.ConfigFile); err != nil {
		logs.Log.Errorf("renaming temporary config file: %v", err)
	}

	go func() {
		open, _ := ui.Question("http://127.0.0/1:5454 - Your Web UI password was set to "+newPassword+
			" and was also printed in the log file:"+c.Config.LogFile+"\n\nOpen Web UI?\n", false)
		if open {
			_ = ui.OpenURL("http://127.0.0.1:5454")
		}
	}()
}

// loadConfiguration brings in, and sometimes creates, the initial running configuration.
func (c *Client) loadConfiguration(ctx context.Context) ([]string, string, error) {
	var (
		msg, newPassword string
		err              error
	)
	// Find or write a config file. This does not parse it.
	// A config file is only written when none is found on Windows, macOS (GUI App only), or Docker.
	// And in the case of Docker, only if `/config` is a mounted volume.
	write := (!c.Flags.Restart && ui.HasGUI()) || mnd.IsDocker
	c.Flags.ConfigFile, newPassword, msg = c.Config.FindAndReturn(ctx, c.Flags.ConfigFile, write)
	output := []string{msg}

	if c.Flags.Restart {
		return output, newPassword, update.Restart(&update.Command{ //nolint:wrapcheck
			Path: os.Args[0],
			Args: []string{"--updated", "--delay", "5s", "--config", c.Flags.ConfigFile},
		})
	}

	// Parse the config file and environment variables.
	result, err := c.getConfig()
	if err != nil {
		return output, newPassword, err
	}

	for file, path := range result.Output {
		output = append(output, fmt.Sprintf("Extra Config File: %s => %s", file, path))
	}

	return output, newPassword, nil
}

// Load configuration from the website.
func (c *Client) loadSiteConfig(ctx context.Context) *clientinfo.ClientInfo {
	clientInfo, err := c.triggers.CI.SaveClientInfo(ctx, true)
	if err != nil || clientInfo == nil {
		if errors.Is(err, website.ErrInvalidAPIKey) {
			logs.Log.ErrorfNoShare("==> Problem validating API key: %v", err)
			logs.Log.ErrorfNoShare(
				"==> NOTICE! No Further requests will be sent to the website until you reload with a valid API Key!")
		} else {
			logs.Log.Printf("==> [WARNING] Problem validating API key: %v, info: %s", err, clientInfo)
		}

		return nil
	}

	// Snapshot is a bit complicated because config-file data (plugins) merges with site-data (snapshot config).
	clientInfo.Actions.Snapshot.Plugins = c.Config.Snapshot.Plugins
	c.Config.Snapshot = clientInfo.Actions.Snapshot
	c.triggers.Snapshot = &c.Config.Snapshot
	c.Config.Services.Plugins = &c.Config.Snapshot.Plugins

	return clientInfo
}

// configureServices is called on startup and on reload, so be careful what goes in here.
func (c *Client) configureServices(ctx context.Context) (*clientinfo.ClientInfo, func()) {
	// Cancelling this context should stop most of the things.
	// It's just a backup, because they all have Stop methods.
	ctx, reload := context.WithCancel(ctx)
	// Website starts a routine that's require to send requests to notifiarr.com. Start it early.
	website.Site.Start(ctx)
	// Load the site config (this connects to Tautulli and notifiarr.com)
	clientInfo := c.loadSiteConfig(ctx)
	if clientInfo != nil && !clientInfo.User.StopLogs {
		share.Enable()
	}
	// Get the Plex server name.
	c.configureServicesPlex(ctx)
	// Start the service checks, which needs the Plex server name.
	c.Services.Start(ctx, c.apps.Plex.Name())
	// Validate the snapshot configuration settings (data from website clientinfo).
	c.Config.Snapshot.Validate()
	// Print the startup configuration info.
	c.PrintStartupInfo(ctx, clientInfo)
	// Start the triggers/actions routines.
	c.triggers.Start(ctx, c.sighup, c.sigkil)

	return clientInfo, reload
}

// configureServicesPlex is called on startup to set the Plex server name.
func (c *Client) configureServicesPlex(ctx context.Context) {
	if !c.Config.Plex.Enabled() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, c.Config.Plex.Timeout.Duration)
	defer cancel()

	if _, err := c.apps.Plex.GetInfo(ctx); err != nil {
		logs.Log.Errorf("=> Getting Plex Media Server info (check url and token): %v", err)
	}
}

func (c *Client) triggerConfigReload(event website.EventType, source string) {
	c.reload <- customReload{event: event, msg: source}
}

// Exit stops the web server and logs our exit messages. Start() calls this.
func (c *Client) Exit(ctx context.Context, reload func()) error {
	defer func() {
		defer logs.Log.CapturePanic()
		logs.Log.Print(" ❌ Good bye! Exiting" + mnd.DurationAge(version.Started))
	}()

	// Start external webserver.
	c.SetupWebServer()
	// Start the Notifiarr.com origin websocket tunnel (internal webserver).
	// This uses the Routes created in the StartWebServer function.
	c.startTunnel(ctx)
	go c.RunWebServer()

	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c.sighup, syscall.SIGHUP)

	var err error
	// For non-GUI systems, this is where the main go routine stops (and waits).
	for {
		select {
		case data := <-c.reload:
			reload()
			if err, reload = c.reloadConfiguration(ctx, data.event, data.msg); err != nil {
				return err
			}
		case sigc := <-c.sigkil:
			logs.Log.Printf("[%s] Need help? %s\n=====> Exiting! Caught Signal: %v", c.Flags.Name(), mnd.HelpLink, sigc)
			return c.stop(ctx, website.EventSignal)
		case sigc := <-c.sighup:
			reload()
			err, reload = c.reloadConfiguration(ctx, website.EventSignal, "Caught Signal: "+sigc.String())
			if err != nil {
				return err
			}
		}
	}
}

// getConfig is the piece shared between loadConfiguration and reloadConfiguration.
func (c *Client) getConfig() (*configfile.SetupResult, error) {
	var err error
	if c.Input, err = c.Config.Get(c.Flags); err != nil {
		return nil, fmt.Errorf("getting config: %w", err)
	}

	result, err := c.Config.Setup(c.Flags)
	if err != nil {
		return nil, fmt.Errorf("setting config: %w", err)
	}

	c.triggers = result.Triggers
	c.Services = result.Services
	c.apps = result.Apps
	c.allow = configfile.MakeIPs(c.Config.Upstreams)

	return result, nil
}

// reloadConfiguration is called from a menu tray item or when a HUP signal is received.
// Re-reads the configuration file and stops/starts all the internal routines.
// Also closes and re-opens all log files. Any errors cause the application to exit.
func (c *Client) reloadConfiguration(ctx context.Context, event website.EventType, source string) (error, func()) {
	logs.Log.Printf("==> Reloading Configuration (%s): %s", event, source)

	err := c.stop(ctx, event)
	if err != nil {
		return fmt.Errorf("stopping web server: %w", err), nil
	}

	// start over.
	c.Config = configfile.NewConfig()

	result, err := c.getConfig()
	if err != nil {
		return err, nil
	}

	if errs := logs.Log.Close(); len(errs) > 0 {
		return fmt.Errorf("closing logger: %w", errs[0]), nil
	}

	defer func() {
		c.SetupWebServer()
		c.startTunnel(ctx)
		go c.RunWebServer()
	}()

	logs.Log.SetupLogging(c.Config.LogConfig)
	clientInfo, reload := c.configureServices(ctx)
	c.setupMenus(clientInfo)
	uptime := mnd.DurationAge(version.Started)

	if c.Flags.ConfigFile == "" {
		logs.Log.Printf(" 🌀 %s v%s-%s Configuration Reloaded! No config file, Uptime: %s",
			c.Flags.Name(), version.Version, version.Revision, uptime)

		if err = ui.Toast("Configuration Reloaded! No config file."); err != nil {
			logs.Log.Errorf("Creating Toast Notification: %v", err)
		}
	} else {
		logs.Log.Printf(" 🌀 %s v%s-%s Configuration Reloaded! Config File: %s, Uptime: %s",
			c.Flags.Name(), version.Version, version.Revision, c.Flags.ConfigFile, uptime)

		if err = ui.Toast("Configuration Reloaded! Config File: %s", c.Flags.ConfigFile); err != nil {
			logs.Log.Errorf("Creating Toast Notification: %v", err)
		}
	}

	for path, file := range result.Output {
		logs.Log.Printf(" => Extra Config File: %s => %s", file, path)
	}

	// This doesn't need to lock because web server is not running.
	c.reloading = false // We're done.

	return nil, reload
}

// stop is called from at least two different exit points and on reload.
func (c *Client) stop(ctx context.Context, event website.EventType) error {
	defer func() {
		defer logs.Log.CapturePanic()
		c.triggers.Stop(event)
		c.Services.Stop()
		website.Site.Stop()
		logs.Log.Printf("==> All systems powered down!")
	}()

	if c.tunnel != nil {
		c.tunnel.Shutdown()
	}

	return c.StopWebServer(ctx)
}
