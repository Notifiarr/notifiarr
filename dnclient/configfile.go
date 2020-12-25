package dnclient

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"golift.io/cnfg"
	"golift.io/cnfg/cnfgfile"
)

func (c *Client) getConfig() (string, error) {
	defer c.fixConfigs()

	var f, msg string

	for _, f = range append([]string{c.Flags.ConfigFile}, configFileLocactions()...) {
		d, err := homedir.Expand(f)
		if err == nil {
			f = d
		}

		if _, err := os.Stat(f); err == nil {
			break
		} // else { c.Print("rip:", err) }

		f = ""
	}

	if f != "" {
		msg = "Using Config File: " + f

		if err := cnfgfile.Unmarshal(c.Config, f); err != nil {
			return msg, fmt.Errorf("config file: %w", err)
		}
	} else {
		msg = "No config file found, using env variables only"
	}

	if _, err := cnfg.UnmarshalENV(c.Config, c.Flags.EnvPrefix); err != nil {
		return msg, fmt.Errorf("environment variables: %w", err)
	}

	return msg, nil
}

func (c *Client) fixConfigs() {
	if c.Config.Timeout.Duration == 0 {
		c.Config.Timeout.Duration = DefaultTimeout
	}

	if c.Config.BindAddr == "" {
		c.Config.BindAddr = DefaultBindAddr
	} else if !strings.Contains(c.Config.BindAddr, ":") {
		c.Config.BindAddr = "0.0.0.0:" + c.Config.BindAddr
	}

	for _, ip := range c.Config.Upstreams {
		if !strings.Contains(ip, "/") {
			if strings.Contains(ip, ":") {
				ip += "/128"
			} else {
				ip += "/32"
			}
		}

		if _, i, err := net.ParseCIDR(ip); err == nil {
			c.allow = append(c.allow, i)
		}
	}
}

func configFileLocactions() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			`~\.dnclient\dnclient.conf`,
			`C:\ProgramData\discordnotifier-client\dnclient.conf`,
			`.\dnclient.conf`,
		}
	case "darwin", "freebsd", "netbsd", "openbsd":
		return []string{
			"/usr/local/etc/discordnotifier-client/dnclient.conf",
			"/etc/discordnotifier-client/dnclient.conf",
			"~/.dnclient/dnclient.conf",
			"./dnclient.conf",
		}
	case "android", "dragonfly", "linux", "nacl", "plan9", "solaris":
		fallthrough
	default:
		return []string{
			"/etc/discordnotifier-client/dnclient.conf",
			"/config/dnclient.conf",
			"/usr/local/etc/discordnotifier-client/dnclient.conf",
			"~/.dnclient/dnclient.conf",
			"./dnclient.conf",
		}
	}
}
