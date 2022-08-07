package apps

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"golift.io/cnfg"
	"golift.io/deluge"
	"golift.io/nzbget"
	"golift.io/qbit"
)

/* Notifiarr Client provides minimal support for pulling data from Download clients. */

type DelugeConfig struct {
	*deluge.Config
	Name           string        `toml:"name" xml:"name" json:"name"`
	Interval       cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	Timeout        cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`
	VerifySSL      bool          `toml:"verify_ssl" xml:"verify_ssl" json:"verifySsl"`
	*deluge.Deluge `toml:"-" xml:"-" json:"-"`
}

type QbitConfig struct {
	*qbit.Config
	Name       string        `toml:"name" xml:"name" json:"name"`
	Interval   cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	Timeout    cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`
	VerifySSL  bool          `toml:"verify_ssl" xml:"verify_ssl" json:"verifySsl"`
	*qbit.Qbit `toml:"-" xml:"-" json:"-"`
}

type RtorrentConfig struct {
	*xmlrpc.Client
	Name      string        `toml:"name" xml:"name" json:"name"`
	URL       string        `toml:"url" xml:"url" json:"url"`
	User      string        `toml:"user" xml:"user" json:"user"`
	Pass      string        `toml:"pass" xml:"pass" json:"pass"`
	Interval  cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	Timeout   cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`
	VerifySSL bool          `toml:"verify_ssl" xml:"verify_ssl" json:"verifySsl"`
}

type TautulliConfig struct {
	Name      string        `toml:"name" xml:"name" json:"name"`
	Interval  cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	Timeout   cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`
	URL       string        `toml:"url" xml:"url" json:"url"`
	APIKey    string        `toml:"api_key" xml:"api_key" json:"apiKey"`
	VerifySSL bool          `toml:"verify_ssl" xml:"verify_ssl" json:"verifySsl"`
}

type NZBGetConfig struct {
	*nzbget.Config
	Name           string        `toml:"name" xml:"name" json:"name"`
	Interval       cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	Timeout        cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`
	VerifySSL      bool          `toml:"verify_ssl" xml:"verify_ssl" json:"verifySsl"`
	*nzbget.NZBGet `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupDeluge() error {
	for idx := range a.Deluge {
		if a.Deluge[idx] == nil || a.Deluge[idx].Config == nil || a.Deluge[idx].Config.URL == "" {
			return fmt.Errorf("%w: missing url: Deluge config %d", ErrInvalidApp, idx+1)
		}

		// a.Deluge[i].Debugf = a.DebugLog.Printf
		if err := a.Deluge[idx].setup(); err != nil {
			return err
		}
	}

	return nil
}

func (d *DelugeConfig) setup() error {
	d.Config.Client = &http.Client{
		Timeout: d.Timeout.Duration,
		Transport: exp.NewMetricsRoundTripper("Deluge", &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: d.VerifySSL}, //nolint:gosec
		}),
	}

	var err error

	if d.Deluge, err = deluge.NewNoAuth(d.Config); err != nil {
		return fmt.Errorf("deluge setup failed: %w", err)
	}

	return nil
}

func (a *Apps) setupQbit() error {
	for idx := range a.Qbit {
		if a.Qbit[idx].Config == nil || a.Qbit[idx].URL == "" {
			return fmt.Errorf("%w: missing url: Qbit config %d", ErrInvalidApp, idx+1)
		}

		// a.Qbit[i].Debugf = a.DebugLog.Printf
		if err := a.Qbit[idx].setup(); err != nil {
			return err
		}
	}

	return nil
}

func (q *QbitConfig) setup() (err error) {
	q.Config.Client = &http.Client{
		Timeout: q.Timeout.Duration,
		Transport: exp.NewMetricsRoundTripper("qBittorrent", &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: q.VerifySSL}, //nolint:gosec
		}),
	}

	q.Qbit, err = qbit.NewNoAuth(q.Config)
	if err != nil {
		return fmt.Errorf("qbit setup failed: %w", err)
	}

	return nil
}

func (a *Apps) setupRtorrent() error {
	for idx := range a.Rtorrent {
		if a.Rtorrent[idx] == nil || a.Rtorrent[idx].URL == "" {
			return fmt.Errorf("%w: missing url: rTorrent config %d", ErrInvalidApp, idx+1)
		}

		a.Rtorrent[idx].Setup()
	}

	return nil
}

func (r *RtorrentConfig) Setup() {
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

	r.Client = xmlrpc.NewClientWithHTTPClient(url, &http.Client{
		Timeout: r.Timeout.Duration,
		Transport: exp.NewMetricsRoundTripper("rTorrent", &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: r.VerifySSL}, //nolint:gosec
		}),
	})
}

func (a *Apps) setupNZBGet() error {
	for idx, nzb := range a.NZBGet {
		if nzb.Config == nil || nzb.URL == "" {
			return fmt.Errorf("%w: missing url: NZBGet config %d", ErrInvalidApp, idx+1)
		}

		nzb.Client = &http.Client{
			Timeout: nzb.Timeout.Duration,
			Transport: exp.NewMetricsRoundTripper("NZBGet", &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: nzb.VerifySSL}, //nolint:gosec
			}),
		}

		a.NZBGet[idx].NZBGet = nzbget.New(a.NZBGet[idx].Config)
	}

	return nil
}
