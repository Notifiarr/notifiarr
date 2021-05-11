package configfile

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	homedir "github.com/mitchellh/go-homedir"
	"golift.io/cnfg"
	"golift.io/cnfg/cnfgfile"
)

// Application Defaults.
const (
	DefaultTimeout  = time.Minute
	DefaultBindAddr = "0.0.0.0:5454"
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
	BindAddr   string              `json:"bind_addr" toml:"bind_addr" xml:"bind_addr" yaml:"bind_addr"`
	SSLCrtFile string              `json:"ssl_cert_file" toml:"ssl_cert_file" xml:"ssl_cert_file" yaml:"ssl_cert_file"`
	SSLKeyFile string              `json:"ssl_key_file" toml:"ssl_key_file" xml:"ssl_key_file" yaml:"ssl_key_file"`
	AutoUpdate string              `json:"auto_update" toml:"auto_update" xml:"auto_update" yaml:"auto_update"`
	Upstreams  []string            `json:"upstreams" toml:"upstreams" xml:"upstreams" yaml:"upstreams"`
	Timeout    cnfg.Duration       `json:"timeout" toml:"timeout" xml:"timeout" yaml:"timeout"`
	Plex       *plex.Server        `json:"plex" toml:"plex" xml:"plex" yaml:"plex"`
	Snapshot   *snapshot.Config    `json:"snapshot" toml:"snapshot" xml:"snapshot" yaml:"snapshot"`
	Services   *services.Config    `json:"services" toml:"services" xml:"services" yaml:"services"`
	Service    []*services.Service `json:"service" toml:"service" xml:"service" yaml:"service"`
	*logs.Logs
	*apps.Apps
	Allow AllowedIPs `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// Get parses a config file and environment variables.
// Sometimes the app runs without a config file entirely.
func (c *Config) Get(configFile, envPrefix string) error {
	defer c.setup()

	if configFile != "" {
		if err := cnfgfile.Unmarshal(c, configFile); err != nil {
			return fmt.Errorf("config file: %w", err)
		}
	}

	if _, err := cnfg.UnmarshalENV(c, envPrefix); err != nil {
		return fmt.Errorf("environment variables: %w", err)
	}

	return nil
}

func (c *Config) setup() {
	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = DefaultTimeout
	}

	if c.AutoUpdate != "" && runtime.GOOS != "windows" {
		c.AutoUpdate = ""
	}

	// Make sure each app has a sane timeout.
	c.Apps.Setup(c.Timeout.Duration)

	if c.BindAddr == "" {
		c.BindAddr = DefaultBindAddr
	} else if !strings.Contains(c.BindAddr, ":") {
		c.BindAddr = "0.0.0.0:" + c.BindAddr
	}

	for _, ip := range c.Upstreams {
		if !strings.Contains(ip, "/") {
			if strings.Contains(ip, ":") {
				ip += "/128"
			} else {
				ip += "/32"
			}
		}

		if _, i, err := net.ParseCIDR(ip); err == nil {
			c.Allow = append(c.Allow, i)
		}
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
	if err := os.MkdirAll(dir, 0750); err != nil {
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
