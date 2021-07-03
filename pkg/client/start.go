// Package client provides the low level assembly of the Notifiarr client application.
// This package orchestrates reading in of configuration, parsing cli flags, actioning
// those cli flags, setting up logging, and finally the starting of internal service
// routines for the webserver, plex sessions, snapshots, service checks and others.
// This package sets everything up for the client application.
package client

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/curl"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	flag "github.com/spf13/pflag"
	"golift.io/cnfg"
	"golift.io/version"
)

// Flags are our CLI input flags.
type Flags struct {
	*flag.FlagSet
	verReq     bool
	testSnaps  bool
	restart    bool
	updated    bool
	cfsync     bool
	pslist     bool
	write      string
	curl       string
	ConfigFile string
	EnvPrefix  string
}

// Client stores all the running data.
type Client struct {
	*logs.Logger
	Flags  *Flags
	Config *configfile.Config
	server *http.Server
	sigkil chan os.Signal
	sighup chan os.Signal
	menu   map[string]ui.MenuItem
	info   string
	notify *notifiarr.Config
	alert  *logs.Cooler
	plex   *logs.Timer
	newCon bool
}

// Errors returned by this package.
var (
	ErrNilAPIKey = fmt.Errorf("API key may not be empty: set a key in config file, OR with environment variable")
)

// NewDefaults returns a new Client pointer with default settings.
func NewDefaults() *Client {
	return &Client{
		sigkil: make(chan os.Signal, 1),
		sighup: make(chan os.Signal, 1),
		menu:   make(map[string]ui.MenuItem),
		plex:   &logs.Timer{},
		alert:  &logs.Cooler{},
		Logger: logs.New(),
		Config: &configfile.Config{
			Apps: &apps.Apps{
				URLBase: "/",
			},
			Services: &services.Config{
				Interval: cnfg.Duration{Duration: services.DefaultSendInterval},
				Parallel: 1,
			},
			BindAddr: mnd.DefaultBindAddr,
			Snapshot: &snapshot.Config{
				Timeout: cnfg.Duration{Duration: snapshot.DefaultTimeout},
			},
			LogConfig: &logs.LogConfig{
				LogFiles:  mnd.DefaultLogFiles,
				LogFileMb: mnd.DefaultLogFileMb,
			},
			Timeout: cnfg.Duration{Duration: mnd.DefaultTimeout},
		}, Flags: &Flags{
			FlagSet:    flag.NewFlagSet(mnd.DefaultName, flag.ExitOnError),
			ConfigFile: os.Getenv(mnd.DefaultEnvPrefix + "_CONFIG_FILE"),
			EnvPrefix:  mnd.DefaultEnvPrefix,
		},
	}
}

// ParseArgs stores the cli flag data into the Flags pointer.
func (f *Flags) ParseArgs(args []string) {
	f.StringVarP(&f.ConfigFile, "config", "c", os.Getenv(mnd.DefaultEnvPrefix+"_CONFIG_FILE"), f.Name()+" Config File")
	f.BoolVar(&f.testSnaps, "snaps", false, f.Name()+"Test Snapshots")
	f.StringVarP(&f.EnvPrefix, "prefix", "p", mnd.DefaultEnvPrefix, "Environment Variable Prefix")
	f.BoolVarP(&f.verReq, "version", "v", false, "Print the version and exit.")
	f.BoolVar(&f.cfsync, "cfsync", false, "Trigger Custom Format sync and exit.")
	f.StringVar(&f.curl, "curl", "", "GET a URL and display headers and payload.")
	f.BoolVar(&f.pslist, "ps", false, "Print the system process list; useful for 'process' service checks")
	f.StringVar(&f.write, "write", "", "Write new config file to provided path. Use - to overwrite '--config' file.")

	if runtime.GOOS == mnd.Windows {
		f.BoolVar(&f.restart, "restart", false, "This is used by auto-update, do not call it")
		f.BoolVar(&f.updated, "updated", false, "This flags causes the app to print an updated message")
	}

	f.Parse(args) // nolint: errcheck
}

// Start runs the app.
func Start() error {
	c := NewDefaults()
	c.Flags.ParseArgs(os.Args[1:])

	switch {
	case c.Flags.verReq:
		fmt.Println(version.Print(c.Flags.Name()))
		return nil // print version and exit.
	case c.Flags.pslist:
		return printProcessList()
	case c.Flags.curl != "":
		resp, body, err := curl.Get(c.Flags.curl) //nolint:bodyclose // it's already closed.
		if err != nil {
			return fmt.Errorf("getting URL '%s': %w", c.Flags.curl, err)
		}

		curl.Print(resp, body)

		return nil
	}

	if err := c.config(); err != nil {
		_, _ = ui.Error(mnd.Title, err.Error())
		return err
	} else if c.Flags.restart || c.Flags.write != "" {
		return nil
	}

	if err := c.start(); err != nil {
		_, _ = ui.Error(mnd.Title, err.Error())
		return err
	}

	return nil
}

func (c *Client) config() error {
	var msg string

	// Find or write a config file. This does not parse it.
	// A config file is only written when none is found on Windows, macOS (GUI App only), or Docker.
	// And in the case of Docker, only if `/config` is a mounted volume.
	write := (!c.Flags.restart && ui.HasGUI()) || os.Getenv("NOTIFIARR_IN_DOCKER") == "true"
	c.Flags.ConfigFile, c.newCon, msg = c.Config.FindAndReturn(c.Flags.ConfigFile, write)

	if c.Flags.restart {
		return update.Restart(&update.Command{ //nolint:wrapcheck
			Path: os.Args[0],
			Args: []string{"--updated", "--config", c.Flags.ConfigFile},
		})
	}

	// Parse the config file and environment variables.
	if err := c.Config.Get(c.Flags.ConfigFile, c.Flags.EnvPrefix); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	// If c.Flags.write is set it will force-write the read-config to the provided file path.
	if c.Flags.write != "" {
		return c.forceWriteWithExit(c.Flags.write, msg)
	}

	c.startupMessage([]string{msg})

	return nil
}

func (c *Client) forceWriteWithExit(f, msg string) error {
	if f == "-" {
		f = c.Flags.ConfigFile
	} else if f == "example" || f == "---" {
		// Bubilding a default template.
		f = c.Flags.ConfigFile
		c.Config.LogFile = ""
		c.Config.DebugLog = ""
		c.Config.HTTPLog = ""
		c.Config.Debug = false
		c.Config.Snapshot.Interval.Duration = mnd.HalfHour
		configfile.ForceAllTmpl = true
	}

	c.Printf("%s", msg)

	f, err := c.Config.Write(f)
	if err != nil { // f purposely shadowed.
		return fmt.Errorf("writing config: %s: %w", f, err)
	}

	c.Print("Wrote Config File:", f)

	return nil
}

func (c *Client) startupMessage(msg []string) {
	if ui.HasGUI() {
		// Setting AppName forces log files (even if not configured).
		// Used for GUI apps that have no console output.
		c.Config.LogConfig.AppName = c.Flags.Name()
	}

	c.Logger.SetupLogging(c.Config.LogConfig)
	c.Printf("%s v%s-%s Starting! [PID: %v]", c.Flags.Name(), version.Version, version.Revision, os.Getpid())

	for _, m := range msg {
		c.Printf("==> %s", m)
	}
}

func (c *Client) start() error {
	if c.Flags.updated {
		c.printUpdateMessage()
	}

	if c.Flags.testSnaps {
		c.checkPlex()
		c.Config.Snapshot.Validate()
		c.logSnaps()

		return nil
	}

	if c.Config.APIKey == "" {
		return fmt.Errorf("%w %s_API_KEY", ErrNilAPIKey, c.Flags.EnvPrefix)
	}

	c.configureServices(!c.Flags.cfsync) // do not collect plex info if cfsync is active.

	if c.Flags.cfsync {
		c.Printf("==> Flag Requested: Syncing Custom Formats and Release Profiles (then exiting)")
		// c.notify.SendFinishedQueueItems(c.notify.BaseURL)
		c.notify.SyncRadarrCF()
		c.notify.SyncSonarrRP()

		return nil
	}

	if err := c.Config.Services.Start(c.Config.Service); err != nil {
		return fmt.Errorf("service checks: %w", err)
	}

	if ci, err := c.notify.GetClientInfo(); err != nil {
		c.Printf("==> [WARNING] API Key may be invalid: %v: %s", err, ci)
	} else if ci != nil {
		c.Printf("==> %s", ci)
	}

	return c.run()
}

func printProcessList() error {
	pslist, err := services.GetAllProcesses()
	if err != nil {
		return fmt.Errorf("unable to get processes: %w", err)
	}

	for _, p := range pslist {
		if runtime.GOOS == "freebsd" {
			fmt.Printf("[%-5d] %s\n", p.PID, p.CmdLine)
			continue
		}

		t := "unknown"
		if !p.Created.IsZero() {
			t = time.Since(p.Created).Round(time.Second).String()
		}

		fmt.Printf("[%-5d] %-11s: %s\n", p.PID, t, p.CmdLine)
	}

	return nil
}

// configureServices is called on startup and on reload.
func (c *Client) configureServices(getPlexInfo bool) {
	c.checkPlex() // This runs plex.Validate().

	if getPlexInfo && c.Config.Plex.Configured() {
		if info, err := c.Config.Plex.GetInfo(); err != nil {
			c.Config.Plex.Name = ""
			c.Errorf("=> Getting Plex Media Server info (check url and token): %v", err)
		} else {
			c.Config.Plex.Name = info.FriendlyName
		}
	}

	c.notify = &notifiarr.Config{
		Apps:    c.Config.Apps,
		Plex:    c.Config.Plex,
		Snap:    c.Config.Snapshot,
		Logger:  c.Logger,
		URL:     notifiarr.ProdURL,
		Timeout: c.Config.Timeout.Duration,
	}

	c.Config.Snapshot.Validate()
	c.PrintStartupInfo()
	c.notify.Start(c.Config.Mode)

	c.Config.Services.Logger = c.Logger
	c.Config.Services.Apps = c.Config.Apps
	c.Config.Services.Notify = c.notify
}

// run turns on the auto updater if enabled, and starts the web server, and system tray icon.
func (c *Client) run() error {
	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	c.setReloadSignals()

	if c.newCon {
		_, _ = c.Config.Write(c.Flags.ConfigFile)
		_ = ui.OpenFile(c.Flags.ConfigFile)
		_, _ = ui.Warning(mnd.Title, "A new configuration file was created @ "+
			c.Flags.ConfigFile+" - it should open in a text editor. "+
			"Please edit the file and reload this application using the tray menu.")
	}

	if c.Config.AutoUpdate != "" {
		go c.AutoWatchUpdate()
	}

	switch ui.HasGUI() {
	case true:
		c.startTray() // This starts the web server.
		return nil    // startTray() calls os.Exit()
	default:
		c.StartWebServer()
		return c.Exit()
	}
}

// starts plex if it's configured.
func (c *Client) checkPlex() {
	if c.Config.Plex.Configured() {
		// Validate only returns an error if Configured == false.
		_ = c.Config.Plex.Validate()
	}
}
