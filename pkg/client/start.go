package client

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	flag "github.com/spf13/pflag"
	"golift.io/cnfg"
	"golift.io/version"
)

// Application Defaults.
const (
	Title            = "Notifiarr"
	DefaultName      = "notifiarr"
	DefaultLogFileMb = 100
	DefaultLogFiles  = 0 // delete none
	DefaultTimeout   = time.Minute
	DefaultBindAddr  = "0.0.0.0:5454"
	DefaultEnvPrefix = "DN"
)

const (
	windows = "windows"
)

// Flags are our CLI input flags.
type Flags struct {
	*flag.FlagSet
	verReq     bool
	testSnaps  bool
	restart    bool
	updated    bool
	ConfigFile string
	EnvPrefix  string
	Mode       string
}

// Config represents the data in our config file.
type Config struct {
	BindAddr   string              `json:"bind_addr" toml:"bind_addr" xml:"bind_addr" yaml:"bind_addr"`
	SSLCrtFile string              `json:"ssl_cert_file" toml:"ssl_cert_file" xml:"ssl_cert_file" yaml:"ssl_cert_file"`
	SSLKeyFile string              `json:"ssl_key_file" toml:"ssl_key_file" xml:"ssl_key_file" yaml:"ssl_key_file"`
	Upstreams  []string            `json:"upstreams" toml:"upstreams" xml:"upstreams" yaml:"upstreams"`
	Timeout    cnfg.Duration       `json:"timeout" toml:"timeout" xml:"timeout" yaml:"timeout"`
	Plex       *plex.Server        `json:"plex" toml:"plex" xml:"plex" yaml:"plex"`
	Snapshot   *snapshot.Config    `json:"snapshot" toml:"snapshot" xml:"snapshot" yaml:"snapshot"`
	Services   *services.Config    `json:"services" toml:"services" xml:"services" yaml:"services"`
	Service    []*services.Service `json:"service" toml:"service" xml:"service" yaml:"service"`
	AutoUpdate string              `json:"auto_update" toml:"auto_update" xml:"auto_update" yaml:"auto_update"`
	*logs.Logs
	*apps.Apps
}

// Client stores all the running data.
type Client struct {
	*logs.Logger
	Flags  *Flags
	Config *Config
	server *http.Server
	sigkil chan os.Signal
	sighup chan os.Signal
	allow  allowedIPs
	menu   map[string]ui.MenuItem
	info   string
	notify *notifiarr.Config
	alert  *logs.Cooler
	plex   *logs.Timer
}

// Errors returned by this package.
var (
	ErrNilAPIKey = fmt.Errorf("API key may not be empty: set a key in config file or with environment variable")
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
		Config: &Config{
			Apps: &apps.Apps{
				URLBase: "/",
			},
			Services: &services.Config{},
			BindAddr: DefaultBindAddr,
			Snapshot: &snapshot.Config{},
			Logs: &logs.Logs{
				LogFiles:  DefaultLogFiles,
				LogFileMb: DefaultLogFileMb,
			},
			Timeout: cnfg.Duration{Duration: DefaultTimeout},
		}, Flags: &Flags{
			FlagSet:    flag.NewFlagSet(DefaultName, flag.ExitOnError),
			ConfigFile: os.Getenv(DefaultEnvPrefix + "_CONFIG_FILE"),
			EnvPrefix:  DefaultEnvPrefix,
		},
	}
}

// ParseArgs stores the cli flag data into the Flags pointer.
func (f *Flags) ParseArgs(args []string) {
	f.StringVarP(&f.ConfigFile, "config", "c", os.Getenv(DefaultEnvPrefix+"_CONFIG_FILE"), f.Name()+" Config File")
	f.BoolVar(&f.testSnaps, "snaps", false, f.Name()+"Test Snapshots")
	f.StringVarP(&f.Mode, "mode", "m", "prod", "Selects Notifiarr URL: test, dev, prod")
	f.StringVarP(&f.EnvPrefix, "prefix", "p", DefaultEnvPrefix, "Environment Variable Prefix")
	f.BoolVarP(&f.verReq, "version", "v", false, "Print the version and exit.")

	if runtime.GOOS == windows {
		f.BoolVar(&f.restart, "restart", false, "This is used by auto-update, do not call it")
		f.BoolVar(&f.updated, "updated", false, "This flags causes the app to print an updated message")
	}

	f.Parse(args) // nolint: errcheck
}

// Start runs the app.
func Start() error {
	err := start()
	if err != nil {
		_, _ = ui.Error(Title, err.Error())
	}

	return err
}

func start() error {
	c := NewDefaults()
	c.Flags.ParseArgs(os.Args[1:])

	if c.Flags.verReq {
		fmt.Println(version.Print(c.Flags.Name()))
		return nil // print version and exit.
	}

	msg := c.findAndSetConfigFile()

	if c.Flags.restart {
		return update.Restart(&update.Command{ //nolint:wrapcheck
			Path: os.Args[0],
			Args: []string{"--updated", "--config", c.Flags.ConfigFile},
		})
	}

	if err := c.getConfig(); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	if ui.HasGUI() {
		// Setting AppName forces log files (even if not configured).
		// Used for GUI apps that have no console output.
		c.Config.Logs.AppName = c.Flags.Name()
	}

	c.Logger.SetupLogging(c.Config.Logs)
	c.Printf("%s v%s-%s Starting! [PID: %v]", c.Flags.Name(), version.Version, version.Revision, os.Getpid())
	c.Printf("==> %s", msg)

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
	} else if err := c.notify.CheckAPIKey(); err != nil {
		c.Print("[WARNING] API Key may be invalid:", err)
	}

	return c.run(strings.HasPrefix(msg, msgConfigCreate))
}

func (c *Client) run(newConfig bool) error {
	c.PrintStartupInfo()
	c.checkPlex()
	c.Config.Snapshot.Validate()
	c.notify.Start(c.Flags.Mode)

	c.Config.Services.Logger = c.Logger
	c.Config.Services.Apps = c.Config.Apps
	c.Config.Services.Notify = c.notify

	if err := c.Config.Services.Start(c.Config.Service); err != nil {
		return fmt.Errorf("service checks: %w", err)
	}

	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	signal.Notify(c.sighup, syscall.SIGHUP)

	if newConfig {
		_ = ui.OpenFile(c.Flags.ConfigFile)
		_, _ = ui.Warning(Title, "A new configuration file was created @ "+
			c.Flags.ConfigFile+" - it should open in a text editor. "+
			"Please edit the file and reload this application using the tray menu.")
	}

	if c.Config.AutoUpdate != "" && runtime.GOOS == windows {
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

// starts plex if it's configured. logs any error.
func (c *Client) checkPlex() bool {
	var err error

	if c.Config.Plex != nil && c.Config.Plex.URL != "" && c.Config.Plex.Token != "" {
		if err = c.Config.Plex.Validate(); err != nil {
			c.Errorf("plex config: %v (plex DISABLED)", err)
			c.Config.Plex = nil
		}
	}

	return err != nil
}
