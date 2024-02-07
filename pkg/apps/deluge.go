package apps

import (
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/deluge"
	"golift.io/starr"
	"golift.io/starr/debuglog"
)

type DelugeConfig struct {
	ExtraConfig
	*deluge.Config
	*deluge.Deluge `toml:"-" xml:"-" json:"-"`
}

// delugeHandlers is called once on startup to register the web API paths.
func (a *Apps) delugeHandlers() {}

func (a *Apps) setupDeluge() error {
	for idx, app := range a.Deluge {
		if app == nil || app.Config == nil || app.URL == "" || app.Password == "" {
			return fmt.Errorf("%w: missing url or password: Deluge config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Deluge config %d", ErrInvalidApp, idx+1)
		}

		// a.Deluge[i].Debugf = a.DebugLog.Printf
		if err := a.Deluge[idx].setup(a.MaxBody, a.Logger); err != nil {
			return err
		}
	}

	return nil
}

func (c *DelugeConfig) setup(maxBody int, logger mnd.Logger) error {
	if logger != nil && logger.DebugEnabled() {
		c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback("Deluge"),
			Redact:  []string{c.Password, c.HTTPPass},
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("Deluge", c.Config.Client.Transport)
	}

	var err error

	if c.Deluge, err = deluge.NewNoAuth(c.Config); err != nil {
		return fmt.Errorf("deluge setup failed: %w", err)
	}

	return nil
}

// Enabled returns true if the instance is enabled and usable.
func (c *DelugeConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.Password != "" && c.Timeout.Duration >= 0
}
