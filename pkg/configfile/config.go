// Package configfile handles all the base configuration-file routines. This
// package also holds the conifiguration for the webserver and notifiarr packages.
// In here you will find config file parsing, validation, and creation.
// The application can re-write its own config file from a built-in template,
// complete with comments. In some circumstances the application writes a brand
// new empty config on startup.
package configfile

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	homedir "github.com/mitchellh/go-homedir"
	"golift.io/cnfg"
	"golift.io/cnfg/cnfgfile"
)

// Return prefixes from FindAndReturn.
const (
	MsgNoConfigFile = "Using env variables only. Config file not found."
	MsgConfigFailed = "Using env variables only. Could not create config file: "
	MsgConfigCreate = "Created new config file: "
	MsgConfigFound  = "Using Config File: "
)

// Config represents the data in our config file.
type Config struct {
	BindAddr   string              `json:"bindAddr" toml:"bind_addr" xml:"bind_addr" yaml:"bindAddr"`
	SSLCrtFile string              `json:"sslCertFile" toml:"ssl_cert_file" xml:"ssl_cert_file" yaml:"sslCertFile"`
	SSLKeyFile string              `json:"sslKeyFile" toml:"ssl_key_file" xml:"ssl_key_file" yaml:"sslKeyFile"`
	AutoUpdate string              `json:"autoUpdate" toml:"auto_update" xml:"auto_update" yaml:"autoUpdate"`
	SendDash   cnfg.Duration       `json:"sendDash" toml:"send_dash" xml:"send_dash" yaml:"sendDash"`
	Mode       string              `json:"mode" toml:"mode" xml:"mode" yaml:"mode"`
	Upstreams  []string            `json:"upstreams" toml:"upstreams" xml:"upstreams" yaml:"upstreams"`
	Timeout    cnfg.Duration       `json:"timeout" toml:"timeout" xml:"timeout" yaml:"timeout"`
	Plex       *plex.Server        `json:"plex" toml:"plex" xml:"plex" yaml:"plex"`
	Snapshot   *snapshot.Config    `json:"snapshot" toml:"snapshot" xml:"snapshot" yaml:"snapshot"`
	Services   *services.Config    `json:"services" toml:"services" xml:"services" yaml:"services"`
	Service    []*services.Service `json:"service" toml:"service" xml:"service" yaml:"service"`
	*logs.LogConfig
	*apps.Apps
	Allow AllowedIPs `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// Get parses a config file and environment variables.
// Sometimes the app runs without a config file entirely.
func (c *Config) Get(configFile, envPrefix string, logger *logs.Logger) (*notifiarr.Config, error) {
	if configFile != "" {
		if err := cnfgfile.Unmarshal(c, configFile); err != nil {
			return nil, fmt.Errorf("config file: %w", err)
		}
	}

	if _, err := cnfg.UnmarshalENV(c, envPrefix); err != nil {
		return nil, fmt.Errorf("environment variables: %w", err)
	}

	// This function returns the notifiarr package Config struct too.
	// This config contains [some of] the same data as the normal Config.
	notifiarr := &notifiarr.Config{
		Apps:    c.Apps,
		Plex:    c.Plex,
		Snap:    c.Snapshot,
		Logger:  logger,
		URL:     notifiarr.ProdURL,
		Timeout: c.Timeout.Duration,
		DashDur: c.SendDash.Duration,
	}
	c.setup(notifiarr)

	// Make sure each app has a sane timeout.
	if err := c.Apps.Setup(c.Timeout.Duration); err != nil {
		return nil, fmt.Errorf("setting up app: %w", err)
	}

	return notifiarr, nil
}

func (c *Config) setup(notifiarr *notifiarr.Config) {
	c.URLBase = path.Join("/", c.URLBase)
	c.Allow = MakeIPs(c.Upstreams)

	if c.Services != nil {
		c.Services.Notifiarr = notifiarr
		c.Services.Apps = c.Apps
	}

	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = mnd.DefaultTimeout
	}

	if c.BindAddr == "" {
		c.BindAddr = mnd.DefaultBindAddr
	} else if !strings.Contains(c.BindAddr, ":") {
		c.BindAddr = "0.0.0.0:" + c.BindAddr
	}

	if ui.HasGUI() && c.LogConfig != nil {
		// Setting AppName forces log files (even if not configured).
		// Used for GUI apps that have no console output.
		c.LogConfig.AppName = mnd.Title
	}
}

// FindAndReturn return a config file. Write one if requested.
func (c *Config) FindAndReturn(configFile string, write bool) (string, bool, string) {
	var confFile string

	defaultConfigFile, configFileList := defaultLocactions()
	for _, f := range append([]string{configFile}, configFileList...) {
		if d, err := homedir.Expand(f); err == nil {
			f = d
		}

		if _, err := os.Stat(f); err == nil {
			confFile = f
			break
		} // else { c.Print("rip:", err) }
	}

	if configFile = ""; confFile != "" {
		configFile, _ = filepath.Abs(confFile)
		return configFile, false, MsgConfigFound + configFile
	}

	if defaultConfigFile == "" || !write {
		return configFile, false, MsgNoConfigFile
	}

	findFile, err := c.Write(defaultConfigFile)
	if err != nil {
		return configFile, true, MsgConfigFailed + err.Error()
	} else if findFile == "" {
		return configFile, false, MsgNoConfigFile
	}

	if err := cnfgfile.Unmarshal(c, findFile); err != nil {
		return findFile, true, MsgConfigCreate + findFile + ": " + err.Error()
	}

	return findFile, true, MsgConfigCreate + findFile
}

// Write config to a file.
func (c *Config) Write(file string) (string, error) {
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
	if err := os.MkdirAll(dir, mnd.Mode0750); err != nil {
		return "", fmt.Errorf("making config dir: %w", err)
	}

	f, err := os.Create(file)
	if err != nil {
		return "", fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	if err := Template.Execute(f, c); err != nil {
		return "", fmt.Errorf("writing config file: %w", err)
	}

	return file, nil
}
