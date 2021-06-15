// Package plex provides the methods the Notifiarr client uses to interface with Plex.
// This package also provides a web handler for incoming plex webhooks, and another
// two handlers for requests from Notifiarr.com to list sessions and kill a session.
// The purpose is to keep track of Plex viewers and send meaningful alerts to their
// respective Disord server about user behavior.
// ie. user started watching something, paused it, resumed it, and finished something.
// This package can be disabled by not providing a Plex Media Server URL or Token.
package plex

import (
	"fmt"
	"net/http"
	"time"

	"golift.io/cnfg"
)

// Server is the Plex configuration from a config file.
// Without a URL or Token, nothing works and this package is unused.
type Server struct {
	Timeout    cnfg.Duration `toml:"timeout" xml:"timeout"`
	Interval   cnfg.Duration `toml:"interval" xml:"interval"`
	URL        string        `toml:"url" xml:"url"`
	Token      string        `toml:"token" xml:"token"`
	AccountMap string        `toml:"account_map" xml:"account_map"`
	Name       string        `toml:"server" xml:"server"`
	ReturnJSON bool          `toml:"return_json" xml:"return_json"`
	Cooldown   cnfg.Duration `toml:"cooldown" xml:"cooldown"`
	SeriesPC   uint          `toml:"series_percent_complete" xml:"series_percent_complete"`
	MoviesPC   uint          `toml:"movies_percent_complete" xml:"movies_percent_complete"`
	client     *http.Client
}

const (
	defaultTimeout  = 10 * time.Second
	minimumTimeout  = 2 * time.Second
	defaultCooldown = 15 * time.Second
	minimumCooldown = 10 * time.Second
	minimumInterval = 5 * time.Minute
	minimumComplete = 70
	maximumComplete = 99
)

// WaitTime is the recommended wait time to pull plex sessions after a webhook.
const WaitTime = 10 * time.Second

// ErrNoURLToken is returned when there is no token or URL.
var ErrNoURLToken = fmt.Errorf("token or URL for Plex missing")

// Validate checks input values and starts the cron interval if it's configured.
func (s *Server) Validate() error { //nolint:cyclop
	if s == nil || s.URL == "" || s.Token == "" {
		return ErrNoURLToken
	}

	if s.Interval.Duration < minimumInterval && s.Interval.Duration != 0 {
		s.Interval.Duration = minimumInterval
	}

	if s.SeriesPC > maximumComplete {
		s.SeriesPC = maximumComplete
	} else if s.SeriesPC != 0 && s.SeriesPC < minimumComplete {
		s.SeriesPC = minimumComplete
	}

	if s.MoviesPC > maximumComplete {
		s.MoviesPC = maximumComplete
	} else if s.MoviesPC != 0 && s.MoviesPC < minimumComplete {
		s.MoviesPC = minimumComplete
	}

	if s.Timeout.Duration == 0 {
		s.Timeout.Duration = defaultTimeout
	} else if s.Timeout.Duration < minimumTimeout {
		s.Timeout.Duration = minimumTimeout
	}

	if s.Cooldown.Duration == 0 {
		s.Cooldown.Duration = defaultCooldown
	} else if s.Cooldown.Duration > minimumCooldown {
		s.Cooldown.Duration = minimumCooldown
	}

	if s.Cooldown.Duration < s.Timeout.Duration {
		s.Cooldown.Duration = s.Timeout.Duration
	}

	return nil
}
