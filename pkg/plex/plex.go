package plex

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/logs"
	"golift.io/cnfg"
)

type Server struct {
	Timeout      cnfg.Duration `toml:"timeout"`
	Interval     cnfg.Duration `toml:"timeout"`
	URL          string        `toml:"url"`
	Token        string        `toml:"token"`
	Secret       string        `toml:"secret"`
	AccountMap   string        `toml:"account_map"`
	Name         string        `toml:"server"`
	ReturnJSON   bool          `toml:"return_json"`
	Cooldown     cnfg.Duration `toml:"cooldown"`
	client       *http.Client
	stopChan     chan struct{}
	*logs.Logger `toml:"zfs_pools"`
}

const (
	defaultTimeout  = 10 * time.Second
	minimumTimeout  = 2 * time.Second
	defaultCooldown = 15 * time.Second
	minimumCooldown = 10 * time.Second
	minimumInterval = 5 * time.Minute
	plexWaitTime    = 10 * time.Second
)

var ErrNoURLToken = fmt.Errorf("token or URL for Plex missing")

// Start checks input values and starts the cron interval if it's configured.
func (s *Server) Start() error {
	if s == nil || s.URL == "" || s.Token == "" {
		return ErrNoURLToken
	}

	if s.Interval.Duration < minimumInterval {
		s.Interval.Duration = minimumInterval
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

	go s.startCron()

	return nil
}

func (s *Server) startCron() {
	if s.Interval.Duration == 0 {
		return
	}

	t := time.NewTicker(s.Interval.Duration)
	s.stopChan = make(chan struct{})
	s.Printf("==> Plex Sessions Collection Started, interval: %v", s.Interval)

	for {
		select {
		case <-t.C:
			if body, err := s.SendMeta(nil); err != nil {
				s.Errorf("Sending Plex Session to Notifiarr: %v: %v", err, string(body))
				continue
			}

			s.Printf("Plex Sessions sent to Notifiarr, sending again in %s", s.Interval)
		case <-s.stopChan:
			t.Stop()
			return
		}
	}
}

func (s *Server) Stop() {
	if s == nil || s.stopChan == nil {
		return
	}

	s.stopChan <- struct{}{}
	close(s.stopChan)
	s.stopChan = nil
}
