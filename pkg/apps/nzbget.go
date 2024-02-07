package apps

import (
	"fmt"
	"strings"

	"golift.io/nzbget"
	"golift.io/starr"
	"golift.io/starr/debuglog"
)

type NZBGetConfig struct {
	ExtraConfig
	*nzbget.Config
	*nzbget.NZBGet `toml:"-" xml:"-" json:"-"`
}

// nzbgetHandlers is called once on startup to register the web API paths.
func (a *Apps) nzbgetHandlers() {}

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
