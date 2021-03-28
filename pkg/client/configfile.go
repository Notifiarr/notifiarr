package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Go-Lift-TV/notifiarr/pkg/bindata"
	"github.com/Go-Lift-TV/notifiarr/pkg/notifiarr"
	"github.com/Go-Lift-TV/notifiarr/pkg/snapshot"
	"github.com/Go-Lift-TV/notifiarr/pkg/ui"
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

// getConfig attempts to find or create a config file.
// Sometimes the app runs without a config entirely.
func (c *Client) getConfig() (string, error) {
	defer c.setupConfig()

	confFile := ""
	msg := msgNoConfigFile

	defaultConfigFile, configFileList := configFileLocactions()
	for _, f := range append([]string{c.Flags.ConfigFile}, configFileList...) {
		d, err := homedir.Expand(f)
		if err == nil {
			f = d
		}

		if _, err := os.Stat(f); err == nil {
			confFile = f
			break
		} // else { c.Print("rip:", err) }
	}

	if confFile != "" {
		c.Flags.ConfigFile, _ = filepath.Abs(confFile)
		msg = msgConfigFound + c.Flags.ConfigFile

		if err := cnfgfile.Unmarshal(c.Config, c.Flags.ConfigFile); err != nil {
			return msg, fmt.Errorf("config file: %w", err)
		}
	} else if findFile, err := c.createConfigFile(defaultConfigFile); err != nil {
		msg = msgConfigFailed + err.Error()
	} else if findFile != "" {
		c.Flags.ConfigFile = findFile
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
		URL:    notifiarr.ProdURL,
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
	if file == "" {
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

	if a, err := bindata.Asset("../../examples/notifiarr.conf.example"); err != nil {
		return "", fmt.Errorf("getting config file: %w", err)
	} else if _, err = f.Write(a); err != nil {
		return "", fmt.Errorf("writing config file: %w", err)
	}

	if err := cnfgfile.Unmarshal(c.Config, file); err != nil {
		return file, fmt.Errorf("config file: %w", err)
	}

	return file, nil
}

// reloadConfiguration is called from a menu tray item or when a HUP signal is received.
// Re-reads the configuration file and stops/starts all the internal routines.
func (c *Client) reloadConfiguration(msg string) {
	c.Print("==> Reloading Configuration: " + msg)
	c.notify.Stop()
	c.Config.Services.Stop()

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

	if err := c.Config.Services.Start(c.Config.Service); err != nil {
		c.Errorf("Reloading Config: %v", err)
		panic(err)
	}

	c.Print("==> Configuration Reloaded!")

	if failed := c.checkPlex(); failed {
		_, _ = ui.Info(Title, "Configuration Reloaded!\nERROOR: Plex DISABLED due to bad config.")
	} else {
		_, _ = ui.Info(Title, "Configuration Reloaded!")
	}
}

// First string is default config file.
// It is created (later) if no config files are found.
func configFileLocactions() (string, []string) {
	defaultConf := ""

	if os.Getenv("NOTIFIARR_IN_DOCKER") == "true" {
		// Provide a default config on Docker if /config dir exists.
		if f, err := os.Stat("/config"); err == nil && f.IsDir() {
			defaultConf = "/config/notifiarr/notifiarr.conf"
		}
	} else if _, err := os.Stat(snapshot.SynologyConf); err == nil {
		// Provide a default config on Synology.
		defaultConf = "/etc/notifiarr/notifiarr.conf"
	}

	switch runtime.GOOS {
	case "windows":
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
		return defaultConf, []string{
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
		return defaultConf, []string{
			"/etc/notifiarr/notifiarr.conf",
			"/etc/discordnotifier-client/dnclient.conf",
			"/config/notifiarr.conf",
			"/config/dnclient.conf",
			"/usr/local/etc/notifiarr/notifiarr.conf",
			"/usr/local/etc/discordnotifier-client/dnclient.conf",
			"~/.notifiarr/notifiarr.conf",
			"~/.dnclient/dnclient.conf",
			"./notifiarr.conf",
		}
	}
}
