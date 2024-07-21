// cmdconfig contains the input config for commands.
// This is in its own package to avoid an import cycle with the clientinfo package.
package cmdconfig

import "golift.io/cnfg"

type Config struct {
	Name    string        `json:"name"   toml:"name"    xml:"name"    yaml:"name"`
	Hash    string        `json:"hash"   toml:"hash"    xml:"hash"    yaml:"hash"`
	Command string        `json:"-"      toml:"command" xml:"command" yaml:"command"`
	Shell   bool          `json:"shell"  toml:"shell"   xml:"shell"   yaml:"shell"`
	Log     bool          `json:"log"    toml:"log"     xml:"log"     yaml:"log"`
	Notify  bool          `json:"notify" toml:"notify"  xml:"notify"  yaml:"notify"`
	Timeout cnfg.Duration `json:"-"      toml:"timeout" xml:"timeout" yaml:"timeout"`
	Args    int           `json:"args"   toml:"-"       xml:"-"       yaml:"-"`
}
