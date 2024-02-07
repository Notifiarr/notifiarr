package apps

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/hekmon/transmissionrpc/v3"
	"golift.io/starr"
	"golift.io/starr/debuglog"
	"golift.io/version"
)

// transmissionHandlers is called once on startup to register the web API paths.
func (a *Apps) transmissionHandlers() {}

// XmissionConfig is the Transmission input configuration.
type XmissionConfig struct {
	URL  string `toml:"url" xml:"url" json:"url"`
	User string `toml:"user" xml:"user" json:"user"`
	Pass string `toml:"pass" xml:"pass" json:"pass"`
	ExtraConfig
	*transmissionrpc.Client `toml:"-" xml:"-" json:"-"`
}

// Enabled returns true if the instance is enabled and usable.
func (c *XmissionConfig) Enabled() bool {
	return c != nil && c.URL != "" && c.Timeout.Duration >= 0
}

func (a *Apps) setupTransmission() error {
	for idx, app := range a.Transmission {
		if app == nil || app.URL == "" {
			return fmt.Errorf("%w: missing url: Transmission config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.URL, "http://") && !strings.HasPrefix(app.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Transmission config %d", ErrInvalidApp, idx+1)
		}

		endpoint, err := url.Parse(app.URL)
		if err != nil {
			return fmt.Errorf("%w: invalid URL: Transmission config %d", err, idx+1)
		} else if app.User != "" {
			endpoint.User = url.UserPassword(app.User, app.Pass)
		}

		var client *http.Client
		if a.Logger != nil && a.Logger.DebugEnabled() {
			client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  a.Debugf,
				Caller:  metricMakerCallback("Transmission"),
				Redact:  []string{app.Pass},
			})
		} else {
			client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			client.Transport = NewMetricsRoundTripper("Transmission", client.Transport)
		}

		a.Transmission[idx].Client, _ = transmissionrpc.New(endpoint, &transmissionrpc.Config{
			UserAgent:    fmt.Sprintf("%s v%s-%s %s", mnd.Title, version.Version, version.Revision, version.Branch),
			CustomClient: client,
		})
	}

	return nil
}
