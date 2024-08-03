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
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands"
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
	MsgConfigCreate = "Created new config file '%s'. Your Web UI '%s' user password is '%s' " +
		"and will not be printed again. Log in, and change it."
	MsgConfigFound  = "Using Config File: "
	DefaultUsername = "admin"
	DefaultHeader   = "X-Webauth-User"
)

// Config represents the data in our config file.
type Config struct {
	HostID     string                 `json:"hostId"      toml:"host_id"       xml:"host_id"       yaml:"hostId"`
	UIPassword CryptPass              `json:"uiPassword"  toml:"ui_password"   xml:"ui_password"   yaml:"uiPassword"`
	BindAddr   string                 `json:"bindAddr"    toml:"bind_addr"     xml:"bind_addr"     yaml:"bindAddr"`
	SSLCrtFile string                 `json:"sslCertFile" toml:"ssl_cert_file" xml:"ssl_cert_file" yaml:"sslCertFile"`
	SSLKeyFile string                 `json:"sslKeyFile"  toml:"ssl_key_file"  xml:"ssl_key_file"  yaml:"sslKeyFile"`
	Upstreams  []string               `json:"upstreams"   toml:"upstreams"     xml:"upstreams"     yaml:"upstreams"`
	AutoUpdate string                 `json:"autoUpdate"  toml:"auto_update"   xml:"auto_update"   yaml:"autoUpdate"`
	UnstableCh bool                   `json:"unstableCh"  toml:"unstable_ch"   xml:"unstable_ch"   yaml:"unstableCh"`
	Timeout    cnfg.Duration          `json:"timeout"     toml:"timeout"       xml:"timeout"       yaml:"timeout"`
	Retries    int                    `json:"retries"     toml:"retries"       xml:"retries"       yaml:"retries"`
	Snapshot   *snapshot.Config       `json:"snapshot"    toml:"snapshot"      xml:"snapshot"      yaml:"snapshot"`
	Services   *services.Config       `json:"services"    toml:"services"      xml:"services"      yaml:"services"`
	Service    []*services.Service    `json:"service"     toml:"service"       xml:"service"       yaml:"service"`
	EnableApt  bool                   `json:"apt"         toml:"apt"           xml:"apt"           yaml:"apt"`
	WatchFiles []*filewatch.WatchFile `json:"watchFiles"  toml:"watch_file"    xml:"watch_file"    yaml:"watchFiles"`
	Commands   []*commands.Command    `json:"commands"    toml:"command"       xml:"command"       yaml:"commands"`
	*logs.LogConfig
	*apps.Apps
	*website.Server `json:"-" toml:"-" xml:"-" yaml:"-"`
	Allow           AllowedIPs `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// NewConfig returns a fresh config with only defaults and a logger ready to go.
func NewConfig(logger mnd.Logger) *Config {
	return &Config{
		Apps: &apps.Apps{
			URLBase: "/",
			Logger:  logger,
		},
		Services: &services.Config{
			Interval: cnfg.Duration{Duration: services.DefaultSendInterval},
			Logger:   logger,
		},
		BindAddr: mnd.DefaultBindAddr,
		Snapshot: &snapshot.Config{
			Timeout: cnfg.Duration{Duration: snapshot.DefaultTimeout},
			Plugins: snapshot.Plugins{
				Nvidia: &snapshot.NvidiaConfig{},
			},
		},
		LogConfig: &logs.LogConfig{
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
		return nil, fmt.Errorf("encoding config into toml for copying (this is a bug!): %w", err)
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
func (c *Config) Get(flag *Flags) (*Config, error) {
	if flag.ConfigFile != "" {
		files := append([]string{flag.ConfigFile}, flag.ExtraConf...)
		if err := cnfgfile.Unmarshal(c, files...); err != nil {
			return nil, fmt.Errorf("config file: %w", err)
		}
	} else if len(flag.ExtraConf) != 0 {
		if err := cnfgfile.Unmarshal(c, flag.ExtraConf...); err != nil {
			return nil, fmt.Errorf("extra config file: %w", err)
		}
	}

	if _, err := cnfg.UnmarshalENV(c, flag.EnvPrefix); err != nil {
		return nil, fmt.Errorf("environment variables: %w", err)
	}

	return c.CopyConfig()
}

// ExpandHomedir expands a ~ to a homedir, or returns the original path in case of any error.
func ExpandHomedir(filePath string) string {
	expanded, err := homedir.Expand(filePath)
	if err != nil {
		return filePath
	}

	return expanded
}

func (c *Config) Setup(flag *Flags, logger *logs.Logger) (*triggers.Actions, map[string]string, error) {
	output, err := cnfgfile.Parse(c, &cnfgfile.Opts{
		Name:          mnd.Title,
		TransformPath: ExpandHomedir,
		Prefix:        "filepath:",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("filepath variables: %w", err)
	}

	if err := c.setupPassword(); err != nil {
		return nil, nil, err
	}

	c.fixConfig()
	logger.LogConfig = c.LogConfig // this is sorta hacky.

	if err := c.Services.Setup(c.Service); err != nil {
		return nil, nil, fmt.Errorf("service checks: %w", err)
	}

	// Make sure each app has a sane timeout.
	if err = c.Apps.Setup(); err != nil {
		return nil, nil, fmt.Errorf("setting up app: %w", err)
	}

	// Make sure the port is not in use before starting the web server.
	c.BindAddr, err = CheckPort(c.BindAddr)
	if flag.Delay > time.Second {
		err = nil // dont check the port if delay is set.
	}

	// This function returns the notifiarr package Config struct too.
	// This config contains [some of] the same data as the normal Config.
	c.Server = website.New(&website.Config{
		Apps:     c.Apps,
		Logger:   c.Apps.Logger,
		BaseURL:  website.BaseURL,
		Timeout:  c.Timeout,
		Retries:  c.Retries,
		HostID:   c.HostID,
		BindAddr: c.BindAddr,
	})
	c.Services.SetWebsite(c.Server)

	return c.setup(logger, flag), output, err
}

func (c *Config) fixConfig() {
	if c.Retries < 0 {
		c.Retries = 0
	} else if c.Retries == 0 {
		c.Retries = website.DefaultRetries
	}

	if c.UIPassword.Val() == "" && len(c.APIKey) == website.APIKeyLength {
		_ = c.UIPassword.Set(DefaultUsername + ":" + c.APIKey)
	}

	c.Services.Apps = c.Apps
	c.Services.Plugins = &c.Snapshot.Plugins
}

func (c *Config) setup(logger *logs.Logger, flag *Flags) *triggers.Actions {
	c.URLBase = strings.TrimSuffix(path.Join("/", c.URLBase), "/") + "/"
	c.Allow = MakeIPs(c.Upstreams)

	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = mnd.DefaultTimeout
	}

	if ui.HasGUI() && c.LogConfig != nil {
		// Setting AppName forces log files (even if not configured).
		// Used for GUI apps that have no console output.
		c.LogConfig.AppName = mnd.Title
	}

	// Ordering.....
	clientinfo := &clientinfo.Config{
		Server: c.Server,
		Apps:   c.Apps,
	}
	triggers := triggers.New(&triggers.Config{
		Apps:       c.Apps,
		Website:    c.Server,
		Snapshot:   c.Snapshot,
		WatchFiles: c.WatchFiles,
		LogFiles:   c.LogConfig.GetActiveLogFilePaths(),
		Commands:   c.Commands,
		ClientInfo: clientinfo,
		ConfigFile: flag.ConfigFile,
		AutoUpdate: c.AutoUpdate,
		UnstableCh: c.UnstableCh,
		Services:   c.Services,
		Logger:     logger,
	})
	clientinfo.CmdList = triggers.Commands.List()

	return triggers
}

// FindAndReturn return a config file. Write one if requested.
func (c *Config) FindAndReturn(ctx context.Context, configFile string, write bool) (string, string, string) {
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
		return configFile, "", MsgConfigFound + configFile + mnd.DurationAge(stat.ModTime())
	}

	if defaultConfigFile != "" && write {
		return c.writeDefaultConfigFile(ctx, defaultConfigFile, confFile)
	}

	return configFile, "", MsgNoConfigFile
}

func (c *Config) writeDefaultConfigFile(ctx context.Context, defaultFile, configFile string) (string, string, string) {
	// If we are writing a config file, set a password.
	newPassword := c.APIKey
	if len(newPassword) != website.APIKeyLength {
		newPassword = GeneratePassword()
	}

	// Save the original password as plain text.
	c.UIPassword = CryptPass(DefaultUsername + ":" + newPassword)

	findFile, err := c.Write(ctx, defaultFile, false)
	if err != nil {
		return configFile, newPassword, MsgConfigFailed + err.Error()
	} else if findFile == "" {
		return configFile, "", MsgNoConfigFile
	}

	msg := fmt.Sprintf(MsgConfigCreate, findFile, DefaultUsername, newPassword)

	if err := cnfgfile.Unmarshal(c, findFile); err != nil {
		return findFile, newPassword, msg + ": " + err.Error()
	}

	return findFile, newPassword, msg
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

	if encode {
		bzWr, err := bzip2.NewWriter(newFile, &bzip2.WriterConfig{Level: 1})
		if err != nil {
			return "", fmt.Errorf("encoding config file: %w", err)
		}

		defer bzWr.Close()
		writer = bzWr
	}

	if err := Template.Execute(writer, c); err != nil {
		return "", fmt.Errorf("writing config file: %w", err)
	}

	return file, nil
}
