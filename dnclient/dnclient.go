package dnclient

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/gorilla/mux"
	flag "github.com/spf13/pflag"
	"golift.io/cnfg"
	"golift.io/version"
)

// Application Defaults.
const (
	Title            = "DiscordNotifier Client"
	DefaultName      = "discordnotifier-client"
	DefaultLogFileMb = 100
	DefaultLogFiles  = 0 // delete none
	DefaultTimeout   = time.Minute
	DefaultBindAddr  = "0.0.0.0:5454"
	DefaultEnvPrefix = "DN"
)

// Flags are our CLI input flags.
type Flags struct {
	*flag.FlagSet
	verReq     bool
	ConfigFile string
	EnvPrefix  string
}

// Config represents the data in our config file.
type Config struct {
	APIKey     string           `json:"api_key" toml:"api_key" xml:"api_key" yaml:"api_key"`
	BindAddr   string           `json:"bind_addr" toml:"bind_addr" xml:"bind_addr" yaml:"bind_addr"`
	SSLCrtFile string           `json:"ssl_cert_file" toml:"ssl_cert_file" xml:"ssl_cert_file" yaml:"ssl_cert_file"`
	SSLKeyFile string           `json:"ssl_key_file" toml:"ssl_key_file" xml:"ssl_key_file" yaml:"ssl_key_file"`
	Debug      bool             `json:"debug" toml:"debug" xml:"debug" yaml:"debug"`
	Quiet      bool             `json:"quiet" toml:"quiet" xml:"quiet" yaml:"quiet"`
	LogFile    string           `json:"log_file" toml:"log_file" xml:"log_file" yaml:"log_file"`
	LogFiles   int              `json:"log_files" toml:"log_files" xml:"log_files" yaml:"log_files"`
	LogFileMb  int              `json:"log_file_mb" toml:"log_file_mb" xml:"log_file_mb" yaml:"log_file_mb"`
	URLBase    string           `json:"urlbase" toml:"urlbase" xml:"urlbase" yaml:"urlbase"`
	Upstreams  []string         `json:"upstreams" toml:"upstreams" xml:"upstreams" yaml:"upstreams"`
	Timeout    cnfg.Duration    `json:"timeout" toml:"timeout" xml:"timeout" yaml:"timeout"`
	Sonarr     []*SonarrConfig  `json:"sonarr,omitempty" toml:"sonarr" xml:"sonarr" yaml:"sonarr,omitempty"`
	Radarr     []*RadarrConfig  `json:"radarr,omitempty" toml:"radarr" xml:"radarr" yaml:"radarr,omitempty"`
	Lidarr     []*LidarrConfig  `json:"lidarr,omitempty" toml:"lidarr" xml:"lidarr" yaml:"lidarr,omitempty"`
	Readarr    []*ReadarrConfig `json:"readarr,omitempty" toml:"readarr" xml:"readarr" yaml:"readarr,omitempty"`
}

// Logger provides a struct we can pass into other packages.
type Logger struct {
	debug  bool
	Logger *log.Logger
}

// Client stores all the running data.
type Client struct {
	*Logger
	Flags  *Flags
	Config *Config
	server *http.Server
	router *mux.Router
	signal chan os.Signal
	allow  allowedIPs
}

// Errors returned by this package.
var (
	ErrNilAPIKey = fmt.Errorf("API key may not be empty")
	ErrNoApps    = fmt.Errorf("0 applications configured")
)

// NewDefaults returns a new Client pointer with default settings.
func NewDefaults() *Client {
	return &Client{
		signal: make(chan os.Signal, 1),
		Logger: &Logger{Logger: log.New(os.Stdout, "", log.LstdFlags)},
		Config: &Config{
			URLBase:   "/",
			LogFiles:  DefaultLogFiles,
			LogFileMb: DefaultLogFileMb,
			BindAddr:  DefaultBindAddr,
			Timeout:   cnfg.Duration{Duration: DefaultTimeout},
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
	f.StringVarP(&f.EnvPrefix, "prefix", "p", DefaultEnvPrefix, "Environment Variable Prefix")
	f.BoolVarP(&f.verReq, "version", "v", false, "Print the version and exit.")
	f.Parse(args) // nolint: errcheck
}

// Start runs the app.
func Start() error {
	err := start()
	if err != nil && hasGUI() {
		_, _ = dlgs.Error(Title, err.Error())
	}

	return err
}

func start() error {
	log.SetFlags(log.LstdFlags) // in case we throw an error for main.go before logging is setup.

	c := NewDefaults()
	c.Flags.ParseArgs(os.Args[1:])

	if c.Flags.verReq {
		fmt.Println(version.Print(c.Flags.Name()))
		return nil // nolint: nlreturn // print version and exit.
	}

	msg, err := c.getConfig()
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	if c.Config.APIKey == "" {
		return ErrNilAPIKey
	} else if len(c.Config.Radarr) < 1 && len(c.Config.Readarr) < 1 &&
		len(c.Config.Sonarr) < 1 && len(c.Config.Lidarr) < 1 {
		return fmt.Errorf("at least 1 application must be configured: %w", ErrNoApps)
	}

	if strings.HasPrefix(msg, msgConfigCreate) && hasGUI() {
		_ = openFile(c.Flags.ConfigFile)
		_, _ = dlgs.Warning(Title, "A new configuration file was created @ "+
			c.Flags.ConfigFile+" - it should open in a text editor. "+
			"Please edit the file and reload this application using the tray menu.")
	}

	c.SetupLogging()
	c.Printf("%s v%s Starting! [PID: %v] %v", c.Flags.Name(), version.Version, os.Getpid(), version.Started)
	c.Printf("==> %s", msg)
	c.InitStartup()
	c.StartWebServer()
	signal.Notify(c.signal, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	c.startTray() // processing stops here on windows and in mac app.

	return c.Exit()
}
