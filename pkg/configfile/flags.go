package configfile

import (
	"os"
	"runtime"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	flag "github.com/spf13/pflag"
)

/* This file handles application cli flags. */

// Flags are our CLI input flags.
type Flags struct {
	*flag.FlagSet
	VerReq     bool
	TestSnaps  bool
	SendSnaps  bool
	Restart    bool
	Updated    bool
	CFsync     bool
	PSlist     bool
	Write      string
	Curl       string
	ConfigFile string
	EnvPrefix  string
}

// ParseArgs stores the cli flag data into the Flags pointer.
func (f *Flags) ParseArgs(args []string) {
	f.StringVarP(&f.ConfigFile, "config", "c", os.Getenv(mnd.DefaultEnvPrefix+"_CONFIG_FILE"), f.Name()+" Config File.")
	f.BoolVar(&f.TestSnaps, "snaps", false, f.Name()+"Test Snapshots.")
	f.BoolVar(&f.SendSnaps, "send", false, f.Name()+"Send Snapshots; must also pass --snaps for this to work.")
	f.StringVarP(&f.EnvPrefix, "prefix", "p", mnd.DefaultEnvPrefix, "Environment Variable Prefix.")
	f.BoolVarP(&f.VerReq, "version", "v", false, "Print the version and exit.")
	f.BoolVar(&f.CFsync, "cfsync", false, "Trigger Custom Format sync and exit.")
	f.StringVar(&f.Curl, "curl", "", "GET a URL and display headers and payload.")
	f.BoolVar(&f.PSlist, "ps", false, "Print the system process list; useful for 'process' service checks.")
	f.StringVar(&f.Write, "write", "", "Write new config file to provided path. Use - to overwrite '--config' file.")

	if runtime.GOOS == mnd.Windows {
		f.BoolVar(&f.Restart, "restart", false, "This is used by auto-update, do not call it.")
		f.BoolVar(&f.Updated, "updated", false, "This flag causes the app to print an 'updated' message.")
	}

	f.Parse(args) // nolint: errcheck
}
