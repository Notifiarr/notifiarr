package apps

import (
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"golift.io/deluge"
	"golift.io/nzbget"
	"golift.io/qbit"
	"golift.io/starr"
	"golift.io/starr/debuglog"
)

type TautulliConfig struct {
	extraConfig
	*tautulli.Config
}

func (c *TautulliConfig) Setup(maxBody int, debugf func(string, ...interface{})) {
	if !c.Enabled() {
		return
	}

	c.Config.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("Tautulli"),
	})

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
	for idx := range a.Deluge {
		if !a.Deluge[idx].Enabled() {
			return fmt.Errorf("%w: missing url: Deluge config %d", ErrInvalidApp, idx+1)
		}

		// a.Deluge[i].Debugf = a.DebugLog.Printf
		if err := a.Deluge[idx].setup(a.MaxBody, a.Debugf); err != nil {
			return err
		}
	}

	return nil
}

func (c *DelugeConfig) setup(maxBody int, debugf func(string, ...interface{})) error {
	c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("Deluge"),
	})

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
	for idx := range a.SabNZB {
		if !a.SabNZB[idx].Enabled() {
			return fmt.Errorf("%w: missing url: SabNZBd config %d", ErrInvalidApp, idx+1)
		}

		a.SabNZB[idx].Setup(a.MaxBody, a.Debugf)
	}

	return nil
}

func (c *SabNZBConfig) Setup(maxBody int, debugf func(string, ...interface{})) {
	if !c.Enabled() {
		return
	}

	c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("SABnzbd"),
	})

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
	for idx := range a.Qbit {
		if !a.Qbit[idx].Enabled() {
			return fmt.Errorf("%w: missing url: Qbit config %d", ErrInvalidApp, idx+1)
		}

		// a.Qbit[i].Debugf = a.DebugLog.Printf
		if err := a.Qbit[idx].Setup(a.MaxBody, a.Debugf); err != nil {
			return err
		}
	}

	return nil
}

func (c *QbitConfig) Setup(maxBody int, debugf func(string, ...interface{})) error {
	c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("qBittorrent"),
	})

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
	for idx := range a.Rtorrent {
		if !a.Rtorrent[idx].Enabled() {
			return fmt.Errorf("%w: missing url: rTorrent config %d", ErrInvalidApp, idx+1)
		}

		a.Rtorrent[idx].Setup(a.MaxBody, a.Debugf)
	}

	return nil
}

func (c *RtorrentConfig) Setup(maxBody int, debugf func(string, ...interface{})) {
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

	c.Client = xmlrpc.NewClientWithHTTPClient(url, starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("rTorrent"),
	}))
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
	for idx, nzb := range a.NZBGet {
		if !nzb.Enabled() {
			return fmt.Errorf("%w: missing url: NZBGet config %d", ErrInvalidApp, idx+1)
		}

		nzb.Client = starr.ClientWithDebug(nzb.Timeout.Duration, nzb.ValidSSL, debuglog.Config{
			MaxBody: a.MaxBody,
			Debugf:  a.Debugf,
			Caller:  metricMaker("NZBGet"),
		})

		a.NZBGet[idx].NZBGet = nzbget.New(a.NZBGet[idx].Config)
	}

	return nil
}

// Enabled returns true if the instance is enabled and usable.
func (c *NZBGetConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.Timeout.Duration >= 0
}
