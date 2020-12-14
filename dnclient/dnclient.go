package dnclient

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"
	"golift.io/cnfg"
	"golift.io/cnfg/cnfgfile"
	"golift.io/version"
)

const (
	defaultLogFileMb = 10
	defaultLogFiles  = 10
	defaultTimeout   = 10 * time.Second
	defaultBindAddr  = ":5454"
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
	APIKey    string           `json:"api_key" toml:"api_key" xml:"api_key" yaml:"api_key"`
	BindAddr  string           `json:"bind_addr" toml:"bind_addr" xml:"bind_addr" yaml:"bind_addr"`
	Debug     bool             `json:"debug" toml:"debug" xml:"debug" yaml:"debug"`
	Quiet     bool             `json:"quiet" toml:"quiet" xml:"quiet" yaml:"quiet"`
	LogFile   string           `json:"log_file" toml:"log_file" xml:"log_file" yaml:"log_file"`
	LogFiles  int              `json:"log_files" toml:"log_files" xml:"log_files" yaml:"log_files"`
	LogFileMb int              `json:"log_file_mb" toml:"log_file_mb" xml:"log_file_mb" yaml:"log_file_mb"`
	Timeout   cnfg.Duration    `json:"timeout" toml:"timeout" xml:"timeout" yaml:"timeout"`
	Sonarr    []*SonarrConfig  `json:"sonarr,omitempty" toml:"sonarr" xml:"sonarr" yaml:"sonarr,omitempty"`
	Radarr    []*RadarrConfig  `json:"radarr,omitempty" toml:"radarr" xml:"radarr" yaml:"radarr,omitempty"`
	Lidarr    []*LidarrConfig  `json:"lidarr,omitempty" toml:"lidarr" xml:"lidarr" yaml:"lidarr,omitempty"`
	Readarr   []*ReadarrConfig `json:"readarr,omitempty" toml:"readarr" xml:"readarr" yaml:"readarr,omitempty"`
}

// Logger provides a struct we can pass into other packages.
type Logger struct {
	debug  bool
	Logger *log.Logger
}

// Client stores all the running data.
type Client struct {
	Flags *Flags
	*Config
	*Logger
}

// IncomingPayload is the data we expect to get from discord notifier.
type IncomingPayload struct {
	Root  string `json:"root_folder"` // optional
	Key   string `json:"api_key"`     // required
	App   string `json:"application"` // required
	Title string `json:"title"`       // required
	Year  int    `json:"year"`        // required
	ID    int    `json:"id"`          // default: 0 (configured instance)
	TMDB  int    `json:"tmdb"`        // required if App = radarr
	TVDB  int    `json:"tvdb"`        // required if App = sonarr
	GRID  int    `json:"grid"`        // required if App = readarr
}

// OutgoingPayload is the response structure for API requests.
type OutgoingPayload struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

func new() *Client {
	return &Client{
		Config: &Config{
			LogFiles:  defaultLogFiles,
			LogFileMb: defaultLogFileMb,
			BindAddr:  defaultBindAddr,
			Timeout:   cnfg.Duration{Duration: defaultTimeout},
		}, Flags: &Flags{FlagSet: flag.NewFlagSet("dnclient", flag.ContinueOnError)},
	}
}

func (f *Flags) parse(args []string) {
	f.StringVarP(&f.ConfigFile, "config", "c", defaultConfFile, "App Config File (TOML Format)")
	f.StringVarP(&f.EnvPrefix, "prefix", "p", "UN", "Environment Variable Prefix")
	f.BoolVarP(&f.verReq, "version", "v", false, "Print the version and exit.")

	_ = f.Parse(args)
}

// Start runs the app.
func Start() error {
	log.SetFlags(log.LstdFlags) // in case we throw an error for main.go before logging is setup.

	c := new()
	if c.Flags.parse(os.Args[1:]); c.Flags.verReq {
		fmt.Println(version.Print(c.Flags.Name()))

		return nil
	}

	if err := cnfgfile.Unmarshal(c.Config, c.Flags.ConfigFile); err != nil {
		return fmt.Errorf("config file: %w", err)
	}

	if _, err := cnfg.UnmarshalENV(c.Config, c.Flags.EnvPrefix); err != nil {
		return fmt.Errorf("environment variables: %w", err)
	}

	if c.APIKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	c.setupLogging()
	c.fixSonarrConfig()
	c.fixReadarrConfig()
	c.fixLidarrConfig()
	c.fixRadarrConfig()
	c.Printf("%s v%s Starting! (PID: %v) %v", c.Flags.Name(), version.Version, os.Getpid(), version.Started)
	c.logStartupInfo()

	go c.RunWebServer()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	c.Printf("[%s] Need help? %s\n=====> Exiting! Caught Signal: %v", c.Flags.Name(), helpLink, <-sig)

	return nil
}
