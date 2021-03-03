package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/bindata"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/notifiarr"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/ui"
	homedir "github.com/mitchellh/go-homedir"
	"golift.io/cnfg"
	"golift.io/cnfg/cnfgfile"
)

const (
	msgNoConfigFile = "Using env variables only. Config file not found."
	msgConfigFailed = "Using env variables only. Could not create config file: "
	msgConfigCreate = "Created new config file: "
	msgConfigFound  = "Using Config File: "
)

func (c *Client) getConfig() (string, error) {
	defer c.setupConfig()

	var f, msg string

	def, cfl := configFileLocactions()
	for _, f = range append([]string{c.Flags.ConfigFile}, cfl...) {
		d, err := homedir.Expand(f)
		if err == nil {
			f = d
		}

		if _, err := os.Stat(f); err == nil {
			break
		} // else { c.Print("rip:", err) }

		f = ""
	}

	msg = msgNoConfigFile

	if f != "" {
		c.Flags.ConfigFile, _ = filepath.Abs(f)
		msg = msgConfigFound + c.Flags.ConfigFile

		if err := cnfgfile.Unmarshal(c.Config, c.Flags.ConfigFile); err != nil {
			return msg, fmt.Errorf("config file: %w", err)
		}
	} else if f, err := c.createConfigFile(def); err != nil {
		msg = msgConfigFailed + err.Error()
	} else if f != "" {
		c.Flags.ConfigFile = f
		msg = msgConfigCreate + c.Flags.ConfigFile
	}

	if _, err := cnfg.UnmarshalENV(c.Config, c.Flags.EnvPrefix); err != nil {
		return msg, fmt.Errorf("environment variables: %w", err)
	}

	return msg, nil
}

func (c *Client) setupConfig() {
	if c.Config.Timeout.Duration == 0 {
		c.Config.Timeout.Duration = DefaultTimeout
	}

	// Make sure each app has a sane timeout.
	c.Config.Apps.Setup(c.Config.Timeout.Duration)
	c.notify = &notifiarr.Config{
		Apps:   c.Config.Apps,
		Plex:   c.Config.Plex,
		Snap:   c.Config.Snapshot,
		Logger: c.Logger,
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

func (c *Client) createConfigFile(file string) (string, error) {
	if !ui.HasGUI() {
		return "", nil
	}

	file, err := homedir.Expand(file)
	if err != nil {
		return "", fmt.Errorf("expanding home: %w", err)
	}

	if file, err = filepath.Abs(file); err != nil {
		return "", fmt.Errorf("absolute file: %w", err)
	}

	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("making config dir: %w", err)
	}

	f, err := os.Create(file)
	if err != nil {
		return "", fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	if a, err := bindata.Asset("../../examples/dnclient.conf.example"); err != nil {
		return "", fmt.Errorf("getting config file: %w", err)
	} else if _, err = f.Write(a); err != nil {
		return "", fmt.Errorf("writing config file: %w", err)
	}

	if err := cnfgfile.Unmarshal(c.Config, file); err != nil {
		return file, fmt.Errorf("config file: %w", err)
	}

	return file, nil
}

func (c *Client) reloadConfiguration(msg string) {
	c.Print("==> Reloading Configuration: " + msg)
	c.notify.Stop()

	if err := c.StopWebServer(); err != nil && !errors.Is(err, ErrNoServer) {
		c.Errorf("Unable to reload configuration: %v", err)
		return
	} else if !errors.Is(err, ErrNoServer) {
		defer c.StartWebServer()
	}

	if _, err := c.getConfig(); err != nil {
		c.Errorf("Reloading Config: %v", err)
		panic(err)
	}

	c.PrintStartupInfo()
	c.notify.Start(c.Flags.Mode)
	c.Print("==> Configuration Reloaded!")

	if failed := c.checkPlex(); failed {
		_, _ = ui.Info(Title, "Configuration Reloaded!\nERROOR: Plex DISABLED due to bad config.")
	} else {
		_, _ = ui.Info(Title, "Configuration Reloaded!")
	}
}

func configFileLocactions() (string, []string) {
	switch runtime.GOOS {
	case "windows":
		return `C:\ProgramData\discordnotifier-client\dnclient.conf`, []string{
			`~\.dnclient\dnclient.conf`,
			`C:\ProgramData\discordnotifier-client\dnclient.conf`,
			`.\dnclient.conf`,
		}
	case "darwin":
		return "~/.dnclient/dnclient.conf", []string{
			"/usr/local/etc/discordnotifier-client/dnclient.conf",
			"/etc/discordnotifier-client/dnclient.conf",
			"~/.dnclient/dnclient.conf",
			"./dnclient.conf",
		}
	case "freebsd", "netbsd", "openbsd":
		return "", []string{
			"/usr/local/etc/discordnotifier-client/dnclient.conf",
			"/etc/discordnotifier-client/dnclient.conf",
			"~/.dnclient/dnclient.conf",
			"./dnclient.conf",
		}
	case "android", "dragonfly", "linux", "nacl", "plan9", "solaris":
		fallthrough
	default:
		return "", []string{
			"/etc/discordnotifier-client/dnclient.conf",
			"/config/dnclient.conf",
			"/usr/local/etc/discordnotifier-client/dnclient.conf",
			"~/.dnclient/dnclient.conf",
			"./dnclient.conf",
		}
	}
}
