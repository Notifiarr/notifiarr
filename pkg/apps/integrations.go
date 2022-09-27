package apps

import (
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
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
	extraConfig
}

func (c *PlexConfig) Setup(maxBody int, logger mnd.Logger) {
	if logger != nil && logger.DebugEnabled() {
		c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback(starr.Plex.String()),
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper(starr.Plex.String(), nil)
	}

	c.URL = strings.TrimRight(c.URL, "/")
	c.Server = plex.New(c.Config)
}

// Enabled returns true if the server is configured, false otherwise.
func (c *PlexConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.Config.URL != "" && c.Config.Token != "" && c.Timeout.Duration >= 0
}

type TautulliConfig struct {
	extraConfig
	*tautulli.Config
}

func (c *TautulliConfig) Setup(maxBody int, logger mnd.Logger) {
	if !c.Enabled() {
		return
	}

	if logger != nil && logger.DebugEnabled() {
		c.Config.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback("Tautulli"),
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("Tautulli", nil)
	}

	c.URL = strings.TrimRight(c.URL, "/")
}

// Enabled returns true if the instance is enabled and usable.
func (c *TautulliConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.APIKey != "" && c.Timeout.Duration >= 0
}

type DelugeConfig struct {
	extraConfig
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
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("Deluge", nil)
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
	extraConfig
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
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("SABnzbd", nil)
	}

	c.URL = strings.TrimRight(c.URL, "/")
}

// Enabled returns true if the instance is enabled and usable.
func (c *SabNZBConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.APIKey != "" && c.Timeout.Duration >= 0
}

type QbitConfig struct {
	extraConfig
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
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("qBittorrent", nil)
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
	extraConfig
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
	client.Transport = NewMetricsRoundTripper("rTorrent", nil)

	if logger != nil && logger.DebugEnabled() {
		client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback("rTorrent"),
		})
	}

	c.Client = xmlrpc.NewClientWithHTTPClient(url, client)
}

// Enabled returns true if the instance is enabled and usable.
func (c *RtorrentConfig) Enabled() bool {
	return c != nil && c.URL != "" && c.Timeout.Duration >= 0
}

type NZBGetConfig struct {
	extraConfig
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
			})
		} else {
			app.Client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			app.Client.Transport = NewMetricsRoundTripper("NZBGet", nil)
		}

		a.NZBGet[idx].NZBGet = nzbget.New(a.NZBGet[idx].Config)
	}

	return nil
}

// Enabled returns true if the instance is enabled and usable.
func (c *NZBGetConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.Timeout.Duration >= 0
}
