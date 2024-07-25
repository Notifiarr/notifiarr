package configfile

import (
	"os"
	"runtime"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	flag "github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

/* This file handles application cli flags. */

// Flags are our CLI input flags.
type Flags struct {
	*flag.FlagSet `json:"-"`
	VerReq        bool          `json:"verReq"`
	LongVerReq    bool          `json:"longVerReq"`
	Restart       bool          `json:"restart"`
	AptHook       bool          `json:"aptHook"`
	Updated       bool          `json:"updated"`
	PSlist        bool          `json:"pslist"`
	Fortune       bool          `json:"fortune"`
	Write         string        `json:"write"`
	Reset         bool          `json:"reset"`
	Curl          string        `json:"curl"`
	ConfigFile    string        `json:"configFile"`
	ExtraConf     []string      `json:"extraConf"`
	EnvPrefix     string        `json:"envPrefix"`
	Headers       []string      `json:"headers"`
	Assets        string        `json:"staticDif"`
	Delay         time.Duration `json:"delay"`
}

// ParseArgs stores the cli flag data into the Flags pointer.
func (f *Flags) ParseArgs(args []string) {
	appName := cases.Title(language.AmericanEnglish).String(f.Name())

	f.StringVarP(&f.ConfigFile, "config", "c",
		os.Getenv(mnd.DefaultEnvPrefix+"_CONFIG_FILE"), appName+" Config File.")
	f.StringSliceVarP(&f.ExtraConf, "extraconfig", "e", nil, "This app supports multiple config files. "+
		"Separate with commas, or pass -e more than once.")
	f.StringVarP(&f.EnvPrefix, "prefix", "p", mnd.DefaultEnvPrefix, "Environment Variable Prefix.")
	f.BoolVar(&f.LongVerReq, "version", false, "Print the long version and exit.")
	f.BoolVarP(&f.VerReq, "v", "v", false, "Print the version and exit.")
	f.BoolVar(&f.Fortune, "fortune", false, "Print a fortune and exit.")
	f.StringVar(&f.Curl, "curl", "", "GET a URL and display headers and payload.")
	f.StringSliceVar(&f.Headers, "header", nil, "Use with --curl to add a request header.")
	f.BoolVar(&f.PSlist, "ps", false, "Print the system process list; useful for 'process' service checks.")
	f.BoolVar(&f.Reset, "reset", false, "Reset the admin password and write it to the config file.")
	f.StringVarP(&f.Write, "write", "w", "", "Write new config file to provided path. Use - to overwrite '--config' file.")
	f.StringVarP(&f.Assets, "assets", "a", "", "Provide path to custom web assets: static files and templates")
	f.BoolVar(&f.AptHook, "apthook", false, "Process a payload from a dpkg Pre-Install-Pkgs hook.")
	f.DurationVar(&f.Delay, "delay", time.Millisecond, "Delay web server startup by this duration.")

	if runtime.GOOS == mnd.Windows {
		f.BoolVar(&f.Restart, "restart", false, "This is used by auto-update, do not call it.")
		f.BoolVar(&f.Updated, "updated", false, "This flag causes the app to print an 'updated' message.")
	}

	f.Parse(args) //nolint: errcheck
}
