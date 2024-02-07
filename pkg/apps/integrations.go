package apps

import (
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr"
	"golift.io/starr/debuglog"
)

type PlexConfig struct {
	*plex.Config
	*plex.Server
	ExtraConfig
}

const (
	Qbit         starr.App = "Qbit"
	SabNZB       starr.App = "SabNZB"
	Rtorrent     starr.App = "rTorrent"
	NZBGet       starr.App = "NZBGet"
	Deluge       starr.App = "Deluge"
	Transmission starr.App = "Transmission"
)

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
