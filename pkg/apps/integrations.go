package apps

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/hekmon/transmissionrpc/v3"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"golift.io/deluge"
	"golift.io/nzbget"
	"golift.io/qbit"
	"golift.io/starr"
	"golift.io/starr/debuglog"
	"golift.io/version"
)

type PlexConfig struct {
	plex.Config
	ExtraConfig
}

type Plex struct {
	PlexConfig
	plex.Server `json:"-" toml:"-" xml:"-"`
}

func (c *PlexConfig) Setup(maxBody int) Plex {
	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = time.Minute
	}

	client := &http.Client{}
	c.URL = strings.TrimRight(c.URL, "/")

	if mnd.Log.DebugEnabled() {
		client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  mnd.Log.Debugf,
			Caller:  metricMakerCallback(starr.Plex.String()),
			Redact:  []string{c.Token},
		})
	} else {
		client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		client.Transport = NewMetricsRoundTripper(starr.Plex.String(), client.Transport)
	}

	return Plex{
		PlexConfig: *c,
		Server:     *plex.New(&c.Config, client),
	}
}

// Enabled returns true if the server is configured, false otherwise.
func (c *PlexConfig) Enabled() bool {
	return c != nil && c.Config.URL != "" && c.Config.Token != "" && c.Timeout.Duration >= 0
}

type TautulliConfig struct {
	ExtraConfig
	tautulli.Config
}

type Tautulli struct {
	TautulliConfig
	*tautulli.Tautulli `json:"-" toml:"-" xml:"-"`
}

func (c *TautulliConfig) Setup(maxBody int) Tautulli {
	if c.Timeout.Duration == 0 {
		c.Timeout.Duration = time.Minute
	}

	if !c.Enabled() {
		return Tautulli{}
	}

	c.URL = strings.TrimRight(c.URL, "/")
	client := &http.Client{}

	if mnd.Log.DebugEnabled() {
		client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  mnd.Log.Debugf,
			Caller:  metricMakerCallback("Tautulli"),
			Redact:  []string{c.APIKey},
		})
	} else {
		client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		client.Transport = NewMetricsRoundTripper("Tautulli", client.Transport)
	}

	return Tautulli{
		TautulliConfig: *c,
		Tautulli:       tautulli.New(c.Config, client),
	}
}

// Enabled returns true if the instance is enabled and usable.
func (c *TautulliConfig) Enabled() bool {
	return c != nil && c.URL != "" && c.APIKey != "" && c.Timeout.Duration >= 0
}

type DelugeConfig struct {
	ExtraConfig
	deluge.Config
}

type Deluge struct {
	DelugeConfig
	*deluge.Deluge `json:"-" toml:"-" xml:"-"`
}

func (a *AppsConfig) setupDeluge() ([]Deluge, error) {
	output := make([]Deluge, len(a.Deluge))

	for idx := range a.Deluge {
		app := &a.Deluge[idx]
		if app.URL == "" {
			return nil, fmt.Errorf("%w: missing url: Deluge config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return nil, fmt.Errorf("%w: URL must begin with http:// or https://: Deluge config %d", ErrInvalidApp, idx+1)
		}

		deluge, err := app.setup(a.MaxBody)
		if err != nil {
			return nil, err
		}

		output[idx] = Deluge{
			DelugeConfig: *app,
			Deluge:       deluge,
		}
	}

	return output, nil
}

func (c *DelugeConfig) setup(maxBody int) (*deluge.Deluge, error) {
	if mnd.Log.DebugEnabled() {
		c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  mnd.Log.Debugf,
			Caller:  metricMakerCallback("Deluge"),
			Redact:  []string{c.Password, c.HTTPPass},
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("Deluge", c.Config.Client.Transport)
	}

	deluge, err := deluge.NewNoAuth(&c.Config)
	if err != nil {
		return nil, fmt.Errorf("deluge setup failed: %w", err)
	}

	return deluge, nil
}

// Enabled returns true if the instance is enabled and usable.
func (c *Deluge) Enabled() bool {
	return c != nil && c.URL != "" && c.Timeout.Duration >= 0
}

type SabNZBConfig struct {
	ExtraConfig
	*sabnzbd.Config
}

type SabNZB struct {
	SabNZBConfig
	*sabnzbd.SabNZB `json:"-" toml:"-" xml:"-"`
}

func (a *AppsConfig) setupSabNZBd() ([]SabNZB, error) {
	output := make([]SabNZB, len(a.SabNZB))

	for idx := range a.SabNZB {
		sabnzb, err := a.SabNZB[idx].Setup(a.MaxBody, idx)
		if err != nil {
			return nil, err
		}

		output[idx] = *sabnzb
	}

	return output, nil
}

func (c *SabNZBConfig) Setup(maxBody, index int) (*SabNZB, error) {
	client := &http.Client{}
	c.URL = strings.TrimRight(c.URL, "/")

	if err := checkUrl(c.URL, "SABnzbd", index); err != nil {
		return nil, err
	}

	if mnd.Log.DebugEnabled() {
		client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  mnd.Log.Debugf,
			Caller:  metricMakerCallback("SABnzbd"),
			Redact:  []string{c.APIKey},
		})
	} else {
		client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		client.Transport = NewMetricsRoundTripper("SABnzbd", client.Transport)
	}

	return &SabNZB{
		SabNZBConfig: *c,
		SabNZB:       sabnzbd.New(*c.Config, client),
	}, nil
}

// Enabled returns true if the instance is enabled and usable.
func (c *SabNZBConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.APIKey != "" && c.Timeout.Duration >= 0
}

type QbitConfig struct {
	ExtraConfig
	qbit.Config
}

type Qbit struct {
	QbitConfig
	*qbit.Qbit `json:"-" toml:"-" xml:"-"`
}

func (a *AppsConfig) setupQbit() ([]Qbit, error) {
	output := make([]Qbit, len(a.Qbit))

	for idx := range a.Qbit {
		app := &a.Qbit[idx]
		if app.URL == "" {
			return nil, fmt.Errorf("%w: missing url: Qbit config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return nil, fmt.Errorf("%w: URL must begin with http:// or https://: Qbit config %d", ErrInvalidApp, idx+1)
		}

		qbit, err := app.Setup(a.MaxBody)
		if err != nil {
			return nil, err
		}

		output[idx] = Qbit{
			QbitConfig: *app,
			Qbit:       qbit,
		}
	}

	return output, nil
}

func (q *QbitConfig) Setup(maxBody int) (*qbit.Qbit, error) {
	if mnd.Log.DebugEnabled() {
		q.Client = starr.ClientWithDebug(q.Timeout.Duration, q.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  mnd.Log.Debugf,
			Caller:  metricMakerCallback("qBittorrent"),
			Redact:  []string{q.Pass, q.HTTPPass},
		})
	} else {
		q.Config.Client = starr.Client(q.Timeout.Duration, q.ValidSSL)
		q.Config.Client.Transport = NewMetricsRoundTripper("qBittorrent", q.Config.Client.Transport)
	}

	qbit, err := qbit.NewNoAuth(&q.Config)
	if err != nil {
		return nil, fmt.Errorf("qbit setup failed: %w", err)
	}

	return qbit, nil
}

// Enabled returns true if the instance is enabled and usable.
func (q *Qbit) Enabled() bool {
	return q != nil && q.URL != "" && q.Timeout.Duration >= 0
}

type RtorrentConfig struct {
	ExtraConfig
	URL  string `json:"url"      toml:"url"  xml:"url"`
	User string `json:"username" toml:"user" xml:"user"`
	Pass string `json:"password" toml:"pass" xml:"pass"`
}

type Rtorrent struct {
	RtorrentConfig
	*xmlrpc.Client `json:"-" toml:"-" xml:"-"`
}

func (a *AppsConfig) setupRtorrent() ([]Rtorrent, error) {
	output := make([]Rtorrent, len(a.Rtorrent))

	for idx := range a.Rtorrent {
		app := &a.Rtorrent[idx]
		if app.URL == "" {
			return nil, fmt.Errorf("%w: missing url: rTorrent config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.URL, "http://") && !strings.HasPrefix(app.URL, "https://") {
			return nil, fmt.Errorf("%w: URL must begin with http:// or https://: rTorrent config %d", ErrInvalidApp, idx+1)
		}

		output[idx] = Rtorrent{
			RtorrentConfig: *app,
			Client:         app.Setup(a.MaxBody),
		}
	}

	return output, nil
}

func (r *RtorrentConfig) Setup(maxBody int) *xmlrpc.Client {
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

	client := starr.Client(r.Timeout.Duration, r.ValidSSL)
	client.Transport = NewMetricsRoundTripper("rTorrent", client.Transport)

	if mnd.Log.DebugEnabled() {
		client = starr.ClientWithDebug(r.Timeout.Duration, r.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  mnd.Log.Debugf,
			Caller:  metricMakerCallback("rTorrent"),
			Redact:  []string{r.Pass},
		})
	}

	return xmlrpc.NewClientWithHTTPClient(url, client)
}

// Enabled returns true if the instance is enabled and usable.
func (r Rtorrent) Enabled() bool {
	return r.URL != "" && r.Timeout.Duration >= 0
}

type NZBGetConfig struct {
	ExtraConfig
	nzbget.Config
}

type NZBGet struct {
	NZBGetConfig
	*nzbget.NZBGet `json:"-" toml:"-" xml:"-"`
}

func (a *AppsConfig) setupNZBGet() ([]NZBGet, error) {
	output := make([]NZBGet, len(a.NZBGet))

	for idx := range a.NZBGet {
		app := &a.NZBGet[idx]
		if err := checkUrl(app.URL, "NZBGet", idx); err != nil {
			return nil, err
		}

		output[idx] = NZBGet{
			NZBGetConfig: *app,
			NZBGet:       app.Setup(a.MaxBody),
		}
	}

	return output, nil
}

func (n *NZBGetConfig) Setup(maxBody int) *nzbget.NZBGet {
	if mnd.Log.DebugEnabled() {
		n.Client = starr.ClientWithDebug(n.Timeout.Duration, n.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  mnd.Log.Debugf,
			Caller:  metricMakerCallback("NZBGet"),
			Redact:  []string{n.Pass},
		})
	} else {
		n.Config.Client = starr.Client(n.Timeout.Duration, n.ValidSSL)
		n.Config.Client.Transport = NewMetricsRoundTripper("NZBGet", n.Config.Client.Transport)
	}

	return nzbget.New(&n.Config)
}

// Enabled returns true if the instance is enabled and usable.
func (n NZBGet) Enabled() bool {
	return n.URL != "" && n.Timeout.Duration >= 0
}

// XmissionConfig is the Transmission input configuration.
type XmissionConfig struct {
	URL  string `json:"url"      toml:"url"  xml:"url"`
	User string `json:"username" toml:"user" xml:"user"`
	Pass string `json:"password" toml:"pass" xml:"pass"`
	ExtraConfig
}

type Xmission struct {
	XmissionConfig
	*transmissionrpc.Client `json:"-" toml:"-" xml:"-"`
}

// Enabled returns true if the instance is enabled and usable.
func (x XmissionConfig) Enabled() bool {
	return x.URL != "" && x.Timeout.Duration >= 0
}

func (a *AppsConfig) setupTransmission() ([]Xmission, error) {
	output := make([]Xmission, len(a.Transmission))

	for idx := range a.Transmission {
		app := &a.Transmission[idx]
		if err := checkUrl(app.URL, "Transmission", idx); err != nil {
			return nil, err
		}

		endpoint, _ := url.Parse(app.URL)
		if app.User != "" {
			endpoint.User = url.UserPassword(app.User, app.Pass)
		}

		var client *http.Client
		if mnd.Log.DebugEnabled() {
			client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  mnd.Log.Debugf,
				Caller:  metricMakerCallback("Transmission"),
				Redact:  []string{app.Pass},
			})
		} else {
			client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			client.Transport = NewMetricsRoundTripper("Transmission", client.Transport)
		}

		rpc, err := transmissionrpc.New(endpoint, &transmissionrpc.Config{
			UserAgent:    fmt.Sprintf("%s v%s-%s %s", mnd.Title, version.Version, version.Revision, version.Branch),
			CustomClient: client,
		})
		if err != nil {
			return nil, err
		}

		output[idx] = Xmission{
			XmissionConfig: *app,
			Client:         rpc,
		}
	}

	return output, nil
}
