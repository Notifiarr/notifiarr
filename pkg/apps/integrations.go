package apps

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/hekmon/transmissionrpc/v2"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"golift.io/deluge"
	"golift.io/nzbget"
	"golift.io/qbit"
	"golift.io/starr"
	"golift.io/starr/debuglog"
)

type PlexConfig struct {
	*plex.Config
	*plex.Server
	ExtraConfig
}

func (c *PlexConfig) Setup(maxBody int, logger mnd.Logger) {
	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = time.Minute
	}

	if logger != nil && logger.DebugEnabled() {
		c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback(starr.Plex.String()),
			Redact:  []string{c.Token},
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper(starr.Plex.String(), c.Config.Client.Transport)
	}

	c.URL = strings.TrimRight(c.URL, "/")
	c.Server = plex.New(c.Config)
}

// Enabled returns true if the server is configured, false otherwise.
func (c *PlexConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.Config.URL != "" && c.Config.Token != "" && c.Timeout.Duration >= 0
}

type TautulliConfig struct {
	ExtraConfig
	tautulli.Config
}

func (c *TautulliConfig) Setup(maxBody int, logger mnd.Logger) {
	if !c.Enabled() {
		return
	}

	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = time.Minute
	}

	if logger != nil && logger.DebugEnabled() {
		c.Config.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback("Tautulli"),
			Redact:  []string{c.APIKey},
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("Tautulli", c.Config.Client.Transport)
	}

	c.URL = strings.TrimRight(c.URL, "/")
}

// Enabled returns true if the instance is enabled and usable.
func (c *TautulliConfig) Enabled() bool {
	return c != nil && c.URL != "" && c.APIKey != "" && c.Timeout.Duration >= 0
}

type DelugeConfig struct {
	ExtraConfig
	*deluge.Config
	*deluge.Deluge `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupDeluge() error {
	for idx, app := range a.Deluge {
		if app == nil || app.Config == nil || app.URL == "" || app.Password == "" {
			return fmt.Errorf("%w: missing url or password: Deluge config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Deluge config %d", ErrInvalidApp, idx+1)
		}

		// a.Deluge[i].Debugf = a.DebugLog.Printf
		if err := a.Deluge[idx].setup(a.MaxBody, a.Logger); err != nil {
			return err
		}
	}

	return nil
}

func (c *DelugeConfig) setup(maxBody int, logger mnd.Logger) error {
	if logger != nil && logger.DebugEnabled() {
		c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback("Deluge"),
			Redact:  []string{c.Password, c.HTTPPass},
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("Deluge", c.Config.Client.Transport)
	}

	var err error

	if c.Deluge, err = deluge.NewNoAuth(c.Config); err != nil {
		return fmt.Errorf("deluge setup failed: %w", err)
	}

	return nil
}

// Enabled returns true if the instance is enabled and usable.
func (c *DelugeConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.Password != "" && c.Timeout.Duration >= 0
}

type SabNZBConfig struct {
	ExtraConfig
	*sabnzbd.Config
}

func (a *Apps) setupSabNZBd() error {
	for idx, app := range a.SabNZB {
		if app == nil || app.Config == nil || app.URL == "" || app.APIKey == "" {
			return fmt.Errorf("%w: missing url or api key: SABnzbd config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: SABnzbd config %d", ErrInvalidApp, idx+1)
		}

		a.SabNZB[idx].Setup(a.MaxBody, a.Logger)
	}

	return nil
}

func (c *SabNZBConfig) Setup(maxBody int, logger mnd.Logger) {
	if !c.Enabled() {
		return
	}

	if logger != nil && logger.DebugEnabled() {
		c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback("SABnzbd"),
			Redact:  []string{c.APIKey},
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("SABnzbd", c.Config.Client.Transport)
	}

	c.URL = strings.TrimRight(c.URL, "/")
}

// Enabled returns true if the instance is enabled and usable.
func (c *SabNZBConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.APIKey != "" && c.Timeout.Duration >= 0
}

type QbitConfig struct {
	ExtraConfig
	*qbit.Config
	*qbit.Qbit `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupQbit() error {
	for idx, app := range a.Qbit {
		if app == nil || app.Config == nil || app.URL == "" {
			return fmt.Errorf("%w: missing url: Qbit config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Qbit config %d", ErrInvalidApp, idx+1)
		}

		// a.Qbit[i].Debugf = a.DebugLog.Printf
		if err := a.Qbit[idx].Setup(a.MaxBody, a.Logger); err != nil {
			return err
		}
	}

	return nil
}

func (c *QbitConfig) Setup(maxBody int, logger mnd.Logger) error {
	if logger != nil && logger.DebugEnabled() {
		c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback("qBittorrent"),
			Redact:  []string{c.Pass, c.HTTPPass},
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("qBittorrent", c.Config.Client.Transport)
	}

	var err error
	if c.Qbit, err = qbit.NewNoAuth(c.Config); err != nil {
		return fmt.Errorf("qbit setup failed: %w", err)
	}

	return nil
}

// Enabled returns true if the instance is enabled and usable.
func (c *QbitConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.Timeout.Duration >= 0
}

type RtorrentConfig struct {
	ExtraConfig
	*xmlrpc.Client
	URL  string `toml:"url" xml:"url" json:"url"`
	User string `toml:"user" xml:"user" json:"user"`
	Pass string `toml:"pass" xml:"pass" json:"pass"`
}

func (a *Apps) setupRtorrent() error {
	for idx, app := range a.Rtorrent {
		if app == nil || app.URL == "" {
			return fmt.Errorf("%w: missing url: rTorrent config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.URL, "http://") && !strings.HasPrefix(app.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: rTorrent config %d", ErrInvalidApp, idx+1)
		}

		a.Rtorrent[idx].Setup(a.MaxBody, a.Logger)
	}

	return nil
}

func (c *RtorrentConfig) Setup(maxBody int, logger mnd.Logger) {
	prefix := "http://"
	if strings.HasPrefix(c.URL, "https://") {
		prefix = "https://"
	}

	// Append the username and password to the URL.
	url := strings.TrimPrefix(strings.TrimPrefix(c.URL, "https://"), "http://")
	if c.User != "" || c.Pass != "" {
		url = prefix + c.User + ":" + c.Pass + "@" + url
	} else {
		url = prefix + url
	}

	client := starr.Client(c.Timeout.Duration, c.ValidSSL)
	client.Transport = NewMetricsRoundTripper("rTorrent", client.Transport)

	if logger != nil && logger.DebugEnabled() {
		client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback("rTorrent"),
			Redact:  []string{c.Pass},
		})
	}

	c.Client = xmlrpc.NewClientWithHTTPClient(url, client)
}

// Enabled returns true if the instance is enabled and usable.
func (c *RtorrentConfig) Enabled() bool {
	return c != nil && c.URL != "" && c.Timeout.Duration >= 0
}

type NZBGetConfig struct {
	ExtraConfig
	*nzbget.Config
	*nzbget.NZBGet `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupNZBGet() error {
	for idx, app := range a.NZBGet {
		if app == nil || app.Config == nil || app.URL == "" {
			return fmt.Errorf("%w: missing url: NZBGet config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: NZBGet config %d", ErrInvalidApp, idx+1)
		}

		if a.Logger != nil && a.Logger.DebugEnabled() {
			app.Client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  a.Debugf,
				Caller:  metricMakerCallback("NZBGet"),
				Redact:  []string{app.Pass},
			})
		} else {
			app.Client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			app.Client.Transport = NewMetricsRoundTripper("NZBGet", app.Client.Transport)
		}

		a.NZBGet[idx].NZBGet = nzbget.New(a.NZBGet[idx].Config)
	}

	return nil
}

// Enabled returns true if the instance is enabled and usable.
func (c *NZBGetConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.Timeout.Duration >= 0
}

// XmissionConfig is the Transmission input configuration.
type XmissionConfig struct {
	URL  string `toml:"url" xml:"url" json:"url"`
	User string `toml:"user" xml:"user" json:"user"`
	Pass string `toml:"pass" xml:"pass" json:"pass"`
	ExtraConfig
	*transmissionrpc.Client `toml:"-" xml:"-" json:"-"`
}

// Enabled returns true if the instance is enabled and usable.
func (c *XmissionConfig) Enabled() bool {
	return c != nil && c.URL != "" && c.Timeout.Duration >= 0
}

func (a *Apps) setupTransmission() error {
	for idx, app := range a.Transmission {
		if app == nil || app.URL == "" {
			return fmt.Errorf("%w: missing url: Transmission config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.URL, "http://") && !strings.HasPrefix(app.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Transmission config %d", ErrInvalidApp, idx+1)
		}

		var client *http.Client
		if a.Logger != nil && a.Logger.DebugEnabled() {
			client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  a.Debugf,
				Caller:  metricMakerCallback("Transmission"),
				Redact:  []string{app.Pass},
			})
		} else {
			client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			client.Transport = NewMetricsRoundTripper("Transmission", client.Transport)
		}

		a.Transmission[idx].Client = transmissionrpc.NewClient(transmissionrpc.Config{
			URL:       app.URL,
			Username:  app.User,
			Password:  app.Pass,
			UserAgent: mnd.Title,
			Client:    client,
		})
	}

	return nil
}
