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
	Interval     cnfg.Duration `toml:"interval"`
	URL          string        `toml:"url"`
	Token        string        `toml:"token"`
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
func (s *Server) Start(apikey string) error {
	if s == nil || s.URL == "" || s.Token == "" {
		return ErrNoURLToken
	}

	if s.Interval.Duration < minimumInterval && s.Interval.Duration != 0 {
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

	go s.startCron(apikey)

	return nil
}

func (s *Server) startCron(apikey string) {
	if s.Interval.Duration == 0 {
		return
	}

	time.Sleep(time.Second)
	t := time.NewTicker(s.Interval.Duration)
	s.stopChan = make(chan struct{})

	defer func() {
		t.Stop()
		close(s.stopChan)
		s.stopChan = nil
	}()

	s.Printf("==> Plex Sessions Collection Started, URL: %s, interval: %v, timeout: %v, cooldown: %v",
		s.URL, s.Interval, s.Timeout, s.Cooldown)

	for {
		select {
		case <-t.C:
			if body, err := s.SendMeta(nil, apikey); err != nil {
				s.Errorf("Sending Plex Session to Notifiarr: %v: %v", err, string(body))
				continue
			}

			s.Printf("Plex Sessions sent to Notifiarr, sending again in %s", s.Interval)
		case <-s.stopChan:
			return
		}
	}
}

func (s *Server) Stop() {
	if s != nil || s.stopChan != nil {
		s.stopChan <- struct{}{}
	}
}
