package apps

import (
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"golift.io/starr"
	"golift.io/starr/debuglog"
)

// rtorrentHandlers is called once on startup to register the web API paths.
func (a *Apps) rtorrentHandlers() {}

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
