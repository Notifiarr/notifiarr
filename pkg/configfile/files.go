package configfile

import (
	"os"
	"runtime"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
)

// First string is default config file.
// It is created (later) if no config files are found.
func defaultLocactions() (string, []string) { //nolint:cyclop,funlen
	defaultUnixConf := ""

	switch {
	case mnd.IsDocker:
		// Provide a default config on Docker if /config dir exists.
		if f, err := os.Stat("/config"); err == nil && f.IsDir() {
			defaultUnixConf = "/config/notifiarr.conf"
		}
	case mnd.IsSynology:
		// Provide a default config on Synology.
		defaultUnixConf = "/etc/notifiarr/notifiarr.conf"
	case ui.HasGUI():
		// Provide a default config for *nix Desktop users.
		defaultUnixConf = "~/.config/notifiarr/notifiarr.conf"
	}

	switch runtime.GOOS {
	case mnd.Windows:
		return `C:\ProgramData\notifiarr\notifiarr.conf`, []string{
			`~\.dnclient\dnclient.conf`,
			`~\.notifiarr\notifiarr.conf`,
			`C:\ProgramData\notifiarr\notifiarr.conf`,
			`C:\ProgramData\discordnotifier-client\dnclient.conf`,
			`.\notifiarr.conf`,
		}
	case "darwin":
		return "~/.notifiarr/notifiarr.conf", []string{
			"/usr/local/etc/notifiarr/notifiarr.conf",
			"/usr/local/etc/discordnotifier-client/dnclient.conf",
			"/etc/notifiarr/notifiarr.conf",
			"/etc/discordnotifier-client/dnclient.conf",
			"~/.notifiarr/notifiarr.conf",
			"~/.dnclient/dnclient.conf",
			"./notifiarr.conf",
		}
	case "freebsd", "netbsd", "openbsd":
		return defaultUnixConf, []string{
			"/usr/local/etc/notifiarr/notifiarr.conf",
			"/usr/local/etc/discordnotifier-client/dnclient.conf",
			"/etc/notifiarr/notifiarr.conf",
			"/etc/discordnotifier-client/dnclient.conf",
			"~/.dnotifiarr/notifiarr.conf",
			"~/.dnclient/dnclient.conf",
			"./notifiarr.conf",
		}
	case "android", "dragonfly", "linux", "nacl", "plan9", "solaris":
		fallthrough
	default:
		return defaultUnixConf, []string{
			"./notifiarr.conf",
			"~/.config/notifiarr/notifiarr.conf",
			"~/.notifiarr/notifiarr.conf",
			"~/.dnclient/dnclient.conf",
			"/etc/notifiarr/notifiarr.conf",
			"/etc/discordnotifier-client/dnclient.conf",
			"/config/notifiarr.conf",
			"/config/dnclient.conf",
			"/usr/local/etc/notifiarr/notifiarr.conf",
			"/usr/local/etc/discordnotifier-client/dnclient.conf",
		}
	}
}
