package apps

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"golift.io/starr"
	"golift.io/starr/prowlarr"
)

// prowlarrHandlers is called once on startup to register the web API paths.
func (a *Apps) prowlarrHandlers() {
}

// ProwlarrConfig represents the input data for a Prowlarr server.
type ProwlarrConfig struct {
	starrConfig
	*starr.Config
	*prowlarr.Prowlarr `toml:"-" xml:"-" json:"-"`
	errorf             func(string, ...interface{}) `toml:"-" xml:"-" json:"-"`
}

func (a *Apps) setupProwlarr() error {
	for idx, app := range a.Prowlarr {
		if app.Config == nil || app.Config.URL == "" {
			return fmt.Errorf("%w: missing url: Prowlarr config %d", ErrInvalidApp, idx+1)
		}

		app.Config.Client = &http.Client{
			Timeout: app.Timeout.Duration,
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: exp.NewMetricsRoundTripper(string(starr.Prowlarr), &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: app.Config.ValidSSL}, //nolint:gosec
			}),
		}
		app.Debugf = a.Debugf
		app.errorf = a.Errorf
		app.URL = strings.TrimRight(app.URL, "/")
		app.Prowlarr = prowlarr.New(app.Config)
	}

	return nil
}
