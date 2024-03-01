package apps

import (
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr"
	"golift.io/starr/debuglog"
)

type SabNZBConfig struct {
	ExtraConfig
	*sabnzbd.Config
}

// sabnzbHandlers is called once on startup to register the web API paths.
func (a *Apps) sabnzbHandlers() {}

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
