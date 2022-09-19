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

func (t *TautulliConfig) Setup(maxBody int, debugf func(string, ...interface{})) {
	if t == nil || t.Config == nil {
		return
	}

	t.Config.Client = starr.ClientWithDebug(t.Timeout.Duration, t.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("Tautulli"),
	})

	t.URL = strings.TrimRight(t.URL, "/")
}

type DelugeConfig struct {
	extraConfig
	*deluge.Config
	*deluge.Deluge `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupDeluge() error {
	for idx := range a.Deluge {
		if a.Deluge[idx] == nil || a.Deluge[idx].Config == nil || a.Deluge[idx].Config.URL == "" {
			return fmt.Errorf("%w: missing url: Deluge config %d", ErrInvalidApp, idx+1)
		}

		// a.Deluge[i].Debugf = a.DebugLog.Printf
		if err := a.Deluge[idx].setup(a.MaxBody, a.Debugf); err != nil {
			return err
		}
	}

	return nil
}

func (d *DelugeConfig) setup(maxBody int, debugf func(string, ...interface{})) error {
	d.Client = starr.ClientWithDebug(d.Timeout.Duration, d.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("Deluge"),
	})

	var err error

	if d.Deluge, err = deluge.NewNoAuth(d.Config); err != nil {
		return fmt.Errorf("deluge setup failed: %w", err)
	}

	return nil
}

type SabNZBConfig struct {
	extraConfig
	*sabnzbd.Config
}

func (a *Apps) setupSabNZBd() error {
	for idx := range a.SabNZB {
		if a.SabNZB[idx] == nil || a.SabNZB[idx].URL == "" {
			return fmt.Errorf("%w: missing url: SabNZBd config %d", ErrInvalidApp, idx+1)
		}

		a.SabNZB[idx].Setup(a.MaxBody, a.Debugf)
	}

	return nil
}

func (s *SabNZBConfig) Setup(maxBody int, debugf func(string, ...interface{})) {
	if s == nil || s.Config == nil {
		return
	}

	s.Client = starr.ClientWithDebug(s.Timeout.Duration, s.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("SABnzbd"),
	})

	s.URL = strings.TrimRight(s.URL, "/")
}

type QbitConfig struct {
	extraConfig
	*qbit.Config
	*qbit.Qbit `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupQbit() error {
	for idx := range a.Qbit {
		if a.Qbit[idx].Config == nil || a.Qbit[idx].URL == "" {
			return fmt.Errorf("%w: missing url: Qbit config %d", ErrInvalidApp, idx+1)
		}

		// a.Qbit[i].Debugf = a.DebugLog.Printf
		if err := a.Qbit[idx].Setup(a.MaxBody, a.Debugf); err != nil {
			return err
		}
	}

	return nil
}

func (q *QbitConfig) Setup(maxBody int, debugf func(string, ...interface{})) error {
	q.Client = starr.ClientWithDebug(q.Timeout.Duration, q.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("qBittorrent"),
	})

	var err error
	if q.Qbit, err = qbit.NewNoAuth(q.Config); err != nil {
		return fmt.Errorf("qbit setup failed: %w", err)
	}

	return nil
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
		if a.Rtorrent[idx] == nil || a.Rtorrent[idx].URL == "" {
			return fmt.Errorf("%w: missing url: rTorrent config %d", ErrInvalidApp, idx+1)
		}

		a.Rtorrent[idx].Setup(a.MaxBody, a.Debugf)
	}

	return nil
}

func (r *RtorrentConfig) Setup(maxBody int, debugf func(string, ...interface{})) {
	prefix := "http://"
	if strings.HasPrefix(r.URL, "https://") {
		prefix = "https://"
	}

	// Append the username and password to the URL.
	url := strings.TrimPrefix(strings.TrimPrefix(r.URL, "https://"), "http://")
	if r.User != "" || r.Pass != "" {
		url = prefix + r.User + ":" + r.Pass + "@" + url
	} else {
		url = prefix + url
	}

	r.Client = xmlrpc.NewClientWithHTTPClient(url, starr.ClientWithDebug(r.Timeout.Duration, r.ValidSSL, debuglog.Config{
		MaxBody: maxBody,
		Debugf:  debugf,
		Caller:  metricMaker("rTorrent"),
	}))
}

type NZBGetConfig struct {
	extraConfig
	*nzbget.Config
	*nzbget.NZBGet `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupNZBGet() error {
	for idx, nzb := range a.NZBGet {
		if nzb.Config == nil || nzb.URL == "" {
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
