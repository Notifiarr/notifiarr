package apps

import (
	"fmt"
	"strings"

	"golift.io/starr"
	"golift.io/starr/debuglog"
	"golift.io/starr/prowlarr"
)

// prowlarrHandlers is called once on startup to register the web API paths.
func (a *Apps) prowlarrHandlers() {
}

// ProwlarrConfig represents the input data for a Prowlarr server.
type ProwlarrConfig struct {
	extraConfig
	*starr.Config
	*prowlarr.Prowlarr `toml:"-" xml:"-" json:"-"`
	errorf             func(string, ...interface{}) `toml:"-" xml:"-" json:"-"`
}

// will be used when we add http handlers for prowlarr.
/* func getProwlarr(r *http.Request) *prowlarr.Prowlarr {
	app, _ := r.Context().Value(starr.Prowlarr).(*ProwlarrConfig)
	return app.Prowlarr
} */

// Enabled returns true if the Prowlarr instance is enabled and usable.
func (p *ProwlarrConfig) Enabled() bool {
	return p != nil && p.Config != nil && p.URL != "" && p.APIKey != "" && p.Timeout.Duration >= 0
}

func (a *Apps) setupProwlarr() error {
	for idx, app := range a.Prowlarr {
		if app.Config == nil || app.Config.URL == "" {
			return fmt.Errorf("%w: missing url: Prowlarr config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Prowlarr config %d", ErrInvalidApp, idx+1)
		}

		if a.Logger.DebugEnabled() {
			app.Config.Client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  a.Debugf,
				Caller:  metricMakerCallback(string(starr.Prowlarr)),
				Redact:  []string{app.APIKey, app.Password, app.HTTPPass},
			})
		} else {
			app.Config.Client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			app.Config.Client.Transport = NewMetricsRoundTripper(starr.Prowlarr.String(), nil)
		}

		app.errorf = a.Errorf
		app.URL = strings.TrimRight(app.URL, "/")
		app.Prowlarr = prowlarr.New(app.Config)
	}

	return nil
}
