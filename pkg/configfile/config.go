// Package configfile handles all the base configuration-file routines. This
// package also holds the configuration for the webserver and notifiarr packages.
// In here you will find config file parsing, validation, and creation.
// The application can re-write its own config file from a built-in template,
// complete with comments. In some circumstances the application writes a brand
// new empty config on startup.
package configfile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands"
	"github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig"
	"github.com/Notifiarr/notifiarr/pkg/triggers/filewatch"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/dsnet/compress/bzip2"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/shirou/gopsutil/v4/host"
	"golift.io/cnfg"
	"golift.io/cnfgfile"
)

// Return prefixes from FindAndReturn.
const (
	MsgNoConfigFile = "Using env variables only. Config file not found."
	MsgConfigFailed = "Using env variables only. Could not create config file: "
	MsgConfigCreate = "Created new config file '%s'."
	MsgConfigFound  = "Using Config File: "
	DefaultUsername = "admin"
	DefaultHeader   = "X-Webauth-User"
)

// Config represents the data in our config file.
type Config struct {
	HostID     string                   `json:"hostId"      toml:"host_id"       xml:"host_id"       yaml:"hostId"`
	UIPassword CryptPass                `json:"uiPassword"  toml:"ui_password"   xml:"ui_password"   yaml:"uiPassword"`
	BindAddr   string                   `json:"bindAddr"    toml:"bind_addr"     xml:"bind_addr"     yaml:"bindAddr"`
	NoCompress bool                     `json:"noCompress"  toml:"no_compress"   xml:"no_compress"   yaml:"noCompress"`
	SSLCrtFile string                   `json:"sslCertFile" toml:"ssl_cert_file" xml:"ssl_cert_file" yaml:"sslCertFile"`
	SSLKeyFile string                   `json:"sslKeyFile"  toml:"ssl_key_file"  xml:"ssl_key_file"  yaml:"sslKeyFile"`
	Upstreams  []string                 `json:"upstreams"   toml:"upstreams"     xml:"upstreams"     yaml:"upstreams"`
	AutoUpdate string                   `json:"autoUpdate"  toml:"auto_update"   xml:"auto_update"   yaml:"autoUpdate"`
	UnstableCh bool                     `json:"unstableCh"  toml:"unstable_ch"   xml:"unstable_ch"   yaml:"unstableCh"`
	Timeout    cnfg.Duration            `json:"timeout"     toml:"timeout"       xml:"timeout"       yaml:"timeout"`
	Retries    int                      `json:"retries"     toml:"retries"       xml:"retries"       yaml:"retries"`
	Snapshot   snapshot.Config          `json:"snapshot"    toml:"snapshot"      xml:"snapshot"      yaml:"snapshot"`
	Services   services.Config          `json:"services"    toml:"services"      xml:"services"      yaml:"services"`
	Service    []services.ServiceConfig `json:"service"     toml:"service"       xml:"service"       yaml:"service"`
	EnableApt  bool                     `json:"apt"         toml:"apt"           xml:"apt"           yaml:"apt"`
	WatchFiles []*filewatch.WatchFile   `json:"watchFiles"  toml:"watch_file"    xml:"watch_file"    yaml:"watchFiles"`
	Endpoints  []*epconfig.Endpoint     `json:"endpoints"   toml:"endpoint"      xml:"endpoint"      yaml:"endpoints"`
	Commands   []*commands.Command      `json:"commands"    toml:"command"       xml:"command"       yaml:"commands"`
	Version    uint                     `json:"version"     toml:"version"       xml:"version"       yaml:"version"`
	logs.LogConfig
	apps.AppsConfig
}

// NewConfig returns a fresh config with only defaults and a logger ready to go.
func NewConfig() *Config {
	return &Config{
		AppsConfig: apps.AppsConfig{
			BaseConfig: apps.BaseConfig{
				URLBase: "/",
			},
		},
		Services: services.Config{},
		BindAddr: mnd.DefaultBindAddr,
		Snapshot: snapshot.Config{
			Timeout: cnfg.Duration{Duration: snapshot.DefaultTimeout},
		},
		LogConfig: logs.LogConfig{
			LogFiles:  mnd.DefaultLogFiles,
			LogFileMb: mnd.DefaultLogFileMb,
		},
		Timeout: cnfg.Duration{Duration: mnd.DefaultTimeout},
	}
}

// CopyConfig returns a copy of the configuration data.
// Useful for writing a config file with different values than what's running.
func (c *Config) CopyConfig() (*Config, error) {
	var newConfig Config

	buf := bytes.Buffer{}
	if err := Template.Execute(&buf, c); err != nil {
		return nil, fmt.Errorf("encoding config into toml for copying (this is a bug!): %w: %#v, %#v", err, c.Endpoints[0], c)
	}

	dec := toml.NewDecoder(&buf)
	if _, err := dec.Decode(&newConfig); err != nil {
		var parseErr toml.ParseError
		if errors.As(err, &parseErr) {
			return nil, fmt.Errorf("decoding config from toml for copying (this is a bug!): %w: %s",
				err, parseErr.ErrorWithUsage())
		}

		return nil, fmt.Errorf("decoding config from toml for copying: %w", err)
	}

	return &newConfig, nil
}

// Get parses a config file and environment variables.
// Sometimes the app runs without a config file entirely.
// You should only run this after getting a config with NewConfig().
func (c *Config) Get(flag *Flags) error {
	if flag.ConfigFile != "" {
		files := append([]string{flag.ConfigFile}, flag.ExtraConf...)
		if err := cnfgfile.Unmarshal(c, files...); err != nil {
			return fmt.Errorf("config file: %w", err)
		}
	} else if len(flag.ExtraConf) != 0 {
		if err := cnfgfile.Unmarshal(c, flag.ExtraConf...); err != nil {
			return fmt.Errorf("extra config file: %w", err)
		}
	}

	if _, err := cnfg.UnmarshalENV(c, flag.EnvPrefix); err != nil {
		return fmt.Errorf("environment variables: %w", err)
	}

	return nil
}

// ExpandHomedir expands a ~ to a homedir, or returns the original path in case of any error.
func ExpandHomedir(filePath string) string {
	expanded, err := homedir.Expand(filePath)
	if err != nil {
		return filePath
	}

	return expanded
}

type SetupResult struct {
	Triggers *triggers.Actions
	Output   map[string]string
	Services *services.Services
	Apps     *apps.Apps
}

func (c *Config) Setup(ctx context.Context, flag *Flags) (*SetupResult, error) {
	output, err := cnfgfile.Parse(c, &cnfgfile.Opts{
		Name:          mnd.Title,
		TransformPath: ExpandHomedir,
		Prefix:        "filepath:",
	})
	if err != nil {
		return nil, fmt.Errorf("filepath variables: %w", err)
	}

	c.fixConfig()

	if err := c.setupPassword(); err != nil {
		return nil, err
	}

	result := &SetupResult{Output: output, Services: services.New(&c.Services)}
	if err := result.Services.Add(c.Service); err != nil {
		return nil, fmt.Errorf("service checks: %w", err)
	}

	// Make sure each app has a sane timeout.
	if result.Apps, err = apps.New(&c.AppsConfig); err != nil {
		return nil, fmt.Errorf("setting up app: %w", err)
	}

	// Add apps to the service checks.
	result.Services.AddApps(result.Apps, c.Snapshot.MySQL)

	// Make sure the port is not in use before starting the web server.
	c.BindAddr, err = CheckPort(c.BindAddr)

	// This function returns the notifiarr package Config struct too.
	// This config contains [some of] the same data as the normal Config.
	website.New(ctx, &website.Config{
		Apps:       result.Apps,
		Timeout:    c.Timeout,
		Retries:    c.Retries,
		HostID:     c.HostID,
		BindAddr:   c.BindAddr,
		NoCompress: c.NoCompress,
	})

	result.Triggers = c.setup(ctx, flag, result.Services, result.Apps)

	return result, err
}

func (c *Config) fixConfig() {
	if c.Retries < 0 {
		c.Retries = 0
	} else if c.Retries == 0 {
		c.Retries = website.DefaultRetries
	}

	// Windows has no stdout, so turn it off.
	c.LogConfig.Quiet = mnd.IsWindows || c.LogConfig.Quiet
	c.Services.Plugins = &c.Snapshot.Plugins
}

func (c *Config) setup(ctx context.Context, flag *Flags, svc *services.Services, apps *apps.Apps) *triggers.Actions {
	c.URLBase = strings.TrimSuffix(path.Join("/", c.URLBase), "/") + "/"

	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = mnd.DefaultTimeout
	}

	if ui.HasGUI() {
		// Setting AppName forces log files (even if not configured).
		// Used for GUI apps that have no console output.
		c.LogConfig.AppName = mnd.Title
	}

	// Ordering.....
	clientinfo := &clientinfo.Config{
		Apps:      apps,
		Endpoints: c.Endpoints,
	}
	triggers := triggers.New(ctx, &triggers.Config{
		Apps:       apps,
		Snapshot:   &c.Snapshot,
		WatchFiles: c.WatchFiles,
		LogFiles:   c.LogConfig.GetActiveLogFilePaths(),
		Commands:   c.Commands,
		ClientInfo: clientinfo,
		ConfigFile: flag.ConfigFile,
		AutoUpdate: c.AutoUpdate,
		UnstableCh: c.UnstableCh,
		Services:   svc,
		Endpoints:  c.Endpoints,
	})
	clientinfo.CmdList = triggers.Commands.List()

	return triggers
}

// FindAndReturn return a config file. Write one if requested.
func (c *Config) FindAndReturn(ctx context.Context, configFile string) (string, string) {
	var (
		confFile string
		stat     os.FileInfo
	)

	defaultConfigFile, configFileList := defaultLocactions()
	for _, fileName := range append([]string{configFile}, configFileList...) {
		d, err := homedir.Expand(fileName)
		if err == nil {
			fileName = d
		}

		if stat, err = os.Stat(fileName); err == nil {
			confFile = fileName
			break
		} //  else { log.Printf("rip: %v", err) }
	}

	if configFile = ""; confFile != "" {
		configFile, _ = filepath.Abs(confFile)
		return configFile, MsgConfigFound + configFile + mnd.DurationAge(stat.ModTime())
	}

	if defaultConfigFile != "" {
		// If we get to this point, we have not found a config file, but we have a "path" for a default, so write it there.
		return c.writeDefaultConfigFile(ctx, defaultConfigFile, confFile)
	}

	return configFile, MsgNoConfigFile
}

func (c *Config) writeDefaultConfigFile(ctx context.Context, defaultFile, configFile string) (string, string) {
	findFile, err := c.Write(ctx, defaultFile, false)
	if err != nil {
		return configFile, MsgConfigFailed + err.Error()
	} else if findFile == "" {
		return configFile, MsgNoConfigFile
	}

	msg := fmt.Sprintf(MsgConfigCreate, findFile)

	if err := cnfgfile.Unmarshal(c, findFile); err != nil {
		return findFile, msg + ": " + err.Error()
	}

	return findFile, msg
}

// Write config to a file.
func (c *Config) Write(ctx context.Context, file string, encode bool) (string, error) { //nolint:cyclop
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

	if _, err = os.Stat(file); err == nil {
		return "", fmt.Errorf("%w: %s", os.ErrExist, file)
	}

	newFile, err := os.Create(file)
	if err != nil {
		return "", fmt.Errorf("creating config file: %w", err)
	}
	defer newFile.Close()

	ctx, cancel := context.WithTimeout(ctx, mnd.DefaultTimeout)
	defer cancel()

	if c.HostID == "" {
		c.HostID, _ = host.HostIDWithContext(ctx)
	}

	var writer io.Writer = newFile

	if encode && os.Getenv("DN_ENCODE_CONFIG_FILE") != mnd.False {
		bzWr, err := bzip2.NewWriter(newFile, &bzip2.WriterConfig{Level: 1})
		if err != nil {
			return "", fmt.Errorf("encoding config file: %w", err)
		}

		defer bzWr.Close()
		writer = bzWr
	}

	c.Version++

	if err := Template.Execute(writer, c); err != nil {
		return "", fmt.Errorf("writing config file: %w", err)
	}

	return file, nil
}
