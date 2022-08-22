// Package client provides the low level assembly of the Notifiarr client application.
// This package orchestrates reading in of configuration, parsing cli flags, actioning
// those cli flags, setting up logging, and finally the starting of internal service
// routines for the webserver, plex sessions, snapshots, service checks and others.
// This package sets everything up for the client application.
package client

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/securecookie"
	"github.com/hako/durafmt"
	flag "github.com/spf13/pflag"
	"golift.io/version"
)

// Client stores all the running data.
type Client struct {
	*logs.Logger
	plexTimer Timer
	Flags     *configfile.Flags
	Config    *configfile.Config
	server    *http.Server
	sigkil    chan os.Signal
	sighup    chan os.Signal
	reload    chan customReload
	website   *website.Server
	triggers  *triggers.Actions
	cookies   *securecookie.SecureCookie
	templat   *template.Template
	webauth   bool
	// this locks anything that may be updated while running.
	// at least "UIPassword" as of its creation.
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
		sigkil: make(chan os.Signal, 1),
		sighup: make(chan os.Signal, 1),
		reload: make(chan customReload, 1),
		Logger: logger,
		Config: configfile.NewConfig(logger),
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

	//nolint:forbidigo,wrapcheck
	switch {
	case client.Flags.LongVerReq: // print version and exit.
		_, err := fmt.Println(version.Print(client.Flags.Name()))
		return err
	case client.Flags.VerReq: // print version and exit.
		_, err := fmt.Println(client.Flags.Name() + " " + version.Version + "-" + version.Revision)
		return err
	case client.Flags.PSlist: // print process list and exit.
		return printProcessList()
	case client.Flags.Fortune: // print fortune and exit.
		_, err := fmt.Println(Fortune())
		return err
	case client.Flags.Curl != "": // curl a URL and exit.
		return curlURL(client.Flags.Curl, client.Flags.Headers)
	default:
		return client.start()
	}
}

func (c *Client) start() error { //nolint:cyclop
	msg, newPassword, err := c.loadConfiguration()

	switch {
	case c.Flags.AptHook:
		return c.handleAptHook() // err?
	case c.Flags.Write != "" && (err == nil || strings.Contains(err.Error(), "ip:port")):
		c.Printf("==> %s", msg)
		return c.forceWriteWithExit(c.Flags.Write)
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

	if err := c.loadAssetsTemplates(); err != nil {
		return err
	}

	clientInfo := c.configureServices()

	if newPassword != "" {
		_, _ = c.Config.Write(c.Flags.ConfigFile)
		_ = ui.OpenFile(c.Flags.ConfigFile)
		_, _ = ui.Warning(mnd.Title, "A new configuration file was created @ "+
			c.Flags.ConfigFile+" - it should open in a text editor. "+
			"Please edit the file and reload this application using the tray menu. "+
			"Your Web UI password was set to "+newPassword+
			" and was also printed in the log file '"+c.Config.LogFile+"' and/or app ouptput.")
	} else if c.Config.AutoUpdate != "" {
		go c.AutoWatchUpdate() // do not run updater if there's a brand new config file.
	}

	if ui.HasGUI() {
		// This starts the web server and calls os.Exit() when done.
		c.startTray(clientInfo)
		return nil
	}

	return c.Exit()
}

// loadConfiguration brings in, and sometimes creates, the initial running confguration.
func (c *Client) loadConfiguration() (msg string, newPassword string, err error) {
	// Find or write a config file. This does not parse it.
	// A config file is only written when none is found on Windows, macOS (GUI App only), or Docker.
	// And in the case of Docker, only if `/config` is a mounted volume.
	write := (!c.Flags.Restart && ui.HasGUI()) || mnd.IsDocker
	c.Flags.ConfigFile, newPassword, msg = c.Config.FindAndReturn(c.Flags.ConfigFile, write)

	if c.Flags.Restart {
		return msg, newPassword, update.Restart(&update.Command{ //nolint:wrapcheck
			Path: os.Args[0],
			Args: []string{"--updated", "--config", c.Flags.ConfigFile},
		})
	}

	// Parse the config file and environment variables.
	c.website, c.triggers, err = c.Config.Get(c.Flags)
	if err != nil {
		return msg, newPassword, fmt.Errorf("getting config: %w", err)
	}

	return msg, newPassword, nil
}

// Load configuration from the website.
func (c *Client) loadSiteConfig() *website.ClientInfo {
	clientInfo, err := c.website.GetClientInfo()
	if err != nil || clientInfo == nil {
		c.Printf("==> [WARNING] Problem validating API key: %v, info: %s", err, clientInfo)
		return nil
	}

	if clientInfo.Actions.Snapshot != nil {
		c.Config.Snapshot.Interval.Duration = clientInfo.Actions.Snapshot.Interval.Duration
		c.Config.Snapshot.Timeout.Duration = clientInfo.Actions.Snapshot.Timeout.Duration
		c.Config.Snapshot.ZFSPools = clientInfo.Actions.Snapshot.ZFSPools
		c.Config.Snapshot.UseSudo = clientInfo.Actions.Snapshot.UseSudo
		c.Config.Snapshot.Raid = clientInfo.Actions.Snapshot.Raid
		c.Config.Snapshot.DriveData = clientInfo.Actions.Snapshot.DriveData
		c.Config.Snapshot.DiskUsage = clientInfo.Actions.Snapshot.DiskUsage
		c.Config.Snapshot.AllDrives = clientInfo.Actions.Snapshot.AllDrives
		c.Config.Snapshot.Quotas = true // TODO: this is for testing.
		// c.Config.Snapshot.Quotas = clientInfo.Actions.Snapshot.Quotas
		c.Config.Snapshot.IOTop = clientInfo.Actions.Snapshot.IOTop
		c.Config.Snapshot.PSTop = clientInfo.Actions.Snapshot.PSTop
		c.Config.Snapshot.MyTop = clientInfo.Actions.Snapshot.MyTop
	}

	return clientInfo
}

// configureServices is called on startup and on reload, so be careful what goes in here.
func (c *Client) configureServices() *website.ClientInfo {
	c.website.Start()

	clientInfo := c.loadSiteConfig()
	c.configureServicesPlex()
	c.website.ReloadCh(c.sighup)
	c.Config.Snapshot.Validate()
	c.PrintStartupInfo(clientInfo)
	c.triggers.Start()
	/* // debug stuff.
	snap, err, _ := c.Config.Snapshot.GetSnapshot()
	b, _ := json.MarshalIndent(snap, "", "   ")
	c.Print(string(b), err)
	os.Exit(1)
	/**/
	c.Config.Services.Start()

	return clientInfo
}

func (c *Client) configureServicesPlex() {
	if !c.Config.Plex.Configured() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Config.Plex.Timeout.Duration)
	defer cancel()

	if info, err := c.Config.Plex.GetInfo(ctx); err != nil {
		c.Config.Plex.Name = ""
		c.Errorf("=> Getting Plex Media Server info (check url and token): %v", err)
	} else {
		c.Config.Plex.Name = info.FriendlyName
	}
}

func (c *Client) triggerConfigReload(event website.EventType, source string) {
	c.reload <- customReload{event: event, msg: source}
}

// Exit stops the web server and logs our exit messages. Start() calls this.
func (c *Client) Exit() error {
	defer func() {
		defer c.CapturePanic()
		//nolint:gomnd
		c.Print(" ❌ Good bye! Uptime:", durafmt.Parse(time.Since(version.Started).Round(time.Second)).LimitFirstN(3))
	}()

	c.StartWebServer()
	c.setSignals()

	// For non-GUI systems, this is where the main go routine stops (and waits).
	for {
		select {
		case data := <-c.reload:
			if err := c.reloadConfiguration(data.event, data.msg); err != nil {
				return err
			}
		case sigc := <-c.sigkil:
			c.Printf("[%s] Need help? %s\n=====> Exiting! Caught Signal: %v", c.Flags.Name(), mnd.HelpLink, sigc)
			return c.stop(website.EventSignal)
		case sigc := <-c.sighup:
			if err := c.checkReloadSignal(sigc); err != nil {
				return err // reloadConfiguration()
			}
		}
	}
}

// reloadConfiguration is called from a menu tray item or when a HUP signal is received.
// Re-reads the configuration file and stops/starts all the internal routines.
// Also closes and re-opens all log files. Any errors cause the application to exit.
func (c *Client) reloadConfiguration(event website.EventType, source string) error {
	c.Printf("==> Reloading Configuration (%s): %s", event, source)

	err := c.stop(event)
	if err != nil {
		return fmt.Errorf("stoping web server: %w", err)
	}

	defer c.StartWebServer()

	// start over.
	c.Config = configfile.NewConfig(c.Logger)
	if c.website, c.triggers, err = c.Config.Get(c.Flags); err != nil {
		return fmt.Errorf("getting configuration: %w", err)
	}

	if errs := c.Logger.Close(); len(errs) > 0 {
		return fmt.Errorf("closing logger(s): %w", errs[0])
	}

	c.Logger.SetupLogging(c.Config.LogConfig)
	clientInfo := c.configureServices()
	c.setupMenus(clientInfo)
	c.Print(" 🌀 Configuration Reloaded! Config File:", c.Flags.ConfigFile)

	if err = ui.Notify("Configuration Reloaded! Config File: %s", c.Flags.ConfigFile); err != nil {
		c.Errorf("Creating Toast Notification: %v", err)
	}

	return nil
}

// stop is called from at least two different exit points and on reload.
func (c *Client) stop(event website.EventType) error {
	defer func() {
		defer c.CapturePanic()
		c.closeDynamicTimerMenus()
		c.triggers.Stop(event)
		c.Config.Services.Stop()
		c.website.Stop()
		c.Print("==> All systems powered down!")
	}()

	return c.StopWebServer()
}
