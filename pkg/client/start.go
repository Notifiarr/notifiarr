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
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/securecookie"
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

	switch {
	case client.Flags.VerReq: // print version and exit.
		fmt.Println(version.Print(client.Flags.Name())) //nolint:forbidigo
		return nil
	case client.Flags.PSlist: // print process list and exit.
		return printProcessList()
	case client.Flags.Curl != "": // curl a URL and exit.
		return curlURL(client.Flags.Curl, client.Flags.Headers)
	default:
		return client.start()
	}
}

func (c *Client) start() error { //nolint:cyclop
	msg, newCon, err := c.loadConfiguration()

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
	c.Printf("%s v%s-%s Starting! [PID: %v] %v",
		c.Flags.Name(), version.Version, version.Revision, os.Getpid(), time.Now())
	c.Printf("==> %s", msg)

	c.printUpdateMessage()

	if err := c.loadAssetsTemplates(); err != nil {
		return err
	}

	clientInfo, err := c.configureServices()
	if err != nil {
		return err
	}

	if newCon {
		_, _ = c.Config.Write(c.Flags.ConfigFile)
		_ = ui.OpenFile(c.Flags.ConfigFile)
		_, _ = ui.Warning(mnd.Title, "A new configuration file was created @ "+
			c.Flags.ConfigFile+" - it should open in a text editor. "+
			"Please edit the file and reload this application using the tray menu.")
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
func (c *Client) loadConfiguration() (msg string, newCon bool, err error) {
	// Find or write a config file. This does not parse it.
	// A config file is only written when none is found on Windows, macOS (GUI App only), or Docker.
	// And in the case of Docker, only if `/config` is a mounted volume.
	write := (!c.Flags.Restart && ui.HasGUI()) || mnd.IsDocker
	c.Flags.ConfigFile, newCon, msg = c.Config.FindAndReturn(c.Flags.ConfigFile, write)

	if c.Flags.Restart {
		return msg, newCon, update.Restart(&update.Command{ //nolint:wrapcheck
			Path: os.Args[0],
			Args: []string{"--updated", "--config", c.Flags.ConfigFile},
		})
	}

	// Parse the config file and environment variables.
	c.website, c.triggers, err = c.Config.Get(c.Flags)
	if err != nil {
		return msg, newCon, fmt.Errorf("getting config: %w", err)
	}

	return msg, newCon, nil
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
		c.Config.Snapshot.ZFSPools = clientInfo.Actions.Snapshot.ZFSPools
		c.Config.Snapshot.UseSudo = clientInfo.Actions.Snapshot.UseSudo
		c.Config.Snapshot.Raid = clientInfo.Actions.Snapshot.Raid
		c.Config.Snapshot.DriveData = clientInfo.Actions.Snapshot.DriveData
		c.Config.Snapshot.DiskUsage = clientInfo.Actions.Snapshot.DiskUsage
		c.Config.Snapshot.AllDrives = clientInfo.Actions.Snapshot.AllDrives
		c.Config.Snapshot.IOTop = clientInfo.Actions.Snapshot.IOTop
		c.Config.Snapshot.PSTop = clientInfo.Actions.Snapshot.PSTop
		c.Config.Snapshot.MyTop = clientInfo.Actions.Snapshot.MyTop
	}

	if clientInfo.Actions.Plex != nil && c.Config.Plex != nil {
		c.Config.Plex.Interval = clientInfo.Actions.Plex.Interval
		c.Config.Plex.Cooldown = clientInfo.Actions.Plex.Cooldown
		c.Config.Plex.MoviesPC = clientInfo.Actions.Plex.MoviesPC
		c.Config.Plex.SeriesPC = clientInfo.Actions.Plex.SeriesPC
		c.Config.Plex.NoActivity = clientInfo.Actions.Plex.NoActivity
		c.Config.Plex.Delay = clientInfo.Actions.Plex.Delay
	}

	c.loadSiteAppsConfig(clientInfo)

	return clientInfo
}

func (c *Client) loadSiteAppsConfig(clientInfo *website.ClientInfo) { //nolint:cyclop
	for _, app := range clientInfo.Actions.Apps.Lidarr {
		if app.Instance < 1 || app.Instance > len(c.Config.Apps.Lidarr) {
			c.ErrorfNoShare("Website provided configuration for missing Lidarr app: %d:%s", app.Instance, app.Name)
			continue
		}

		c.Config.Apps.Lidarr[app.Instance-1].StuckItem = app.Stuck
		c.Config.Apps.Lidarr[app.Instance-1].Corrupt = app.Corrupt
		c.Config.Apps.Lidarr[app.Instance-1].Backup = app.Backup
	}

	for _, app := range clientInfo.Actions.Apps.Prowlarr {
		if app.Instance < 1 || app.Instance > len(c.Config.Apps.Prowlarr) {
			c.ErrorfNoShare("Website provided configuration for missing Prowlarr app: %d:%s", app.Instance, app.Name)
			continue
		}

		c.Config.Apps.Prowlarr[app.Instance-1].Corrupt = app.Corrupt
		c.Config.Apps.Prowlarr[app.Instance-1].Backup = app.Backup
	}

	for _, app := range clientInfo.Actions.Apps.Radarr {
		if app.Instance < 1 || app.Instance > len(c.Config.Apps.Radarr) {
			c.ErrorfNoShare("Website provided configuration for missing Radarr app: %d:%s", app.Instance, app.Name)
			continue
		}

		c.Config.Apps.Radarr[app.Instance-1].StuckItem = app.Stuck
		c.Config.Apps.Radarr[app.Instance-1].Corrupt = app.Corrupt
		c.Config.Apps.Radarr[app.Instance-1].Backup = app.Backup
	}

	for _, app := range clientInfo.Actions.Apps.Readarr {
		if app.Instance < 1 || app.Instance > len(c.Config.Apps.Readarr) {
			c.ErrorfNoShare("Website provided configuration for missing Readarr app: %d:%s", app.Instance, app.Name)
			continue
		}

		c.Config.Apps.Readarr[app.Instance-1].StuckItem = app.Stuck
		c.Config.Apps.Readarr[app.Instance-1].Corrupt = app.Corrupt
		c.Config.Apps.Readarr[app.Instance-1].Backup = app.Backup
	}

	for _, app := range clientInfo.Actions.Apps.Sonarr {
		if app.Instance < 1 || app.Instance > len(c.Config.Apps.Sonarr) {
			c.ErrorfNoShare("Website provided configuration for missing Sonarr app: %d:%s", app.Instance, app.Name)
			continue
		}

		c.Config.Apps.Sonarr[app.Instance-1].StuckItem = app.Stuck
		c.Config.Apps.Sonarr[app.Instance-1].Corrupt = app.Corrupt
		c.Config.Apps.Sonarr[app.Instance-1].Backup = app.Backup
	}
}

// configureServices is called on startup and on reload, so be careful what goes in here.
func (c *Client) configureServices() (*website.ClientInfo, error) {
	clientInfo := c.loadSiteConfig()
	c.configureServicesPlex()
	c.website.ReloadCh(c.sighup)
	c.Config.Snapshot.Validate()
	c.PrintStartupInfo(clientInfo)
	c.website.Start()
	c.triggers.Start()
	/* // debug stuff.
	snap, err, _ := c.Config.Snapshot.GetSnapshot()
	b, _ := json.MarshalIndent(snap, "", "   ")
	c.Print(string(b), err)
	os.Exit(1)
	/**/
	c.Config.Services.Start()

	// Make sure each app has a sane timeout.
	if err := c.Config.Apps.Setup(c.Config.Timeout.Duration); err != nil {
		return clientInfo, fmt.Errorf("setting up app: %w", err)
	}

	return clientInfo, nil
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
			return c.exit()
		case sigc := <-c.sighup:
			if err := c.checkReloadSignal(sigc); err != nil {
				return err // reloadConfiguration()
			}
		}
	}
}

// This is called from at least two different exit points.
func (c *Client) exit() error {
	if c.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout.Duration)
	defer cancel()

	if err := c.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	c.server = nil

	return nil
}

// reloadConfiguration is called from a menu tray item or when a HUP signal is received.
// Re-reads the configuration file and stops/starts all the internal routines.
// Also closes and re-opens all log files. Any errors cause the application to exit.
func (c *Client) reloadConfiguration(event website.EventType, source string) error {
	c.Printf("==> Reloading Configuration (%s): %s", event, source)
	c.closeDynamicTimerMenus()
	c.triggers.Stop(event)
	c.Config.Services.Stop()
	c.website.Stop()

	err := c.StopWebServer()
	if err != nil && !errors.Is(err, ErrNoServer) {
		return fmt.Errorf("stoping web server: %w", err)
	} else if err == nil {
		defer c.StartWebServer()
	}

	// start over.
	c.Config = configfile.NewConfig(c.Logger)
	if c.website, c.triggers, err = c.Config.Get(c.Flags); err != nil {
		return fmt.Errorf("getting configuration: %w", err)
	}

	if errs := c.Logger.Close(); len(errs) > 0 {
		return fmt.Errorf("closing logger(s): %w", errs[0])
	}

	c.Logger.SetupLogging(c.Config.LogConfig)

	clientInfo, err := c.configureServices()
	if err != nil {
		return err
	}

	c.setupMenus(clientInfo)
	c.Print("==> Configuration Reloaded! Config File:", c.Flags.ConfigFile)

	if err = ui.Notify("Configuration Reloaded! Config File: %s", c.Flags.ConfigFile); err != nil {
		c.Errorf("Creating Toast Notification: %v", err)
	}

	return nil
}
