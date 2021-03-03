package notifiarr

import (
	"strings"
	"time"
)

func (c *Config) startPlexCron() {
	if c.Plex == nil || c.Plex.Interval.Duration == 0 || c.Plex.URL == "" || c.Plex.Token == "" {
		return
	}

	time.Sleep(time.Second)
	t := time.NewTicker(c.Plex.Interval.Duration)
	c.stopPlex = make(chan struct{})

	defer func() {
		t.Stop()
		close(c.stopPlex)
		c.stopPlex = nil
	}()

	c.Printf("==> Plex Sessions Collection Started, URL: %s, interval: %v, timeout: %v, cooldown: %v",
		c.Plex.URL, c.Plex.Interval, c.Plex.Timeout, c.Plex.Cooldown)

	for {
		select {
		case <-t.C:
			if body, err := c.SendMeta(nil, c.URL, 0); err != nil {
				c.Errorf("Sending Plex Session to %s: %v: %v", c.URL, err, string(body))
			} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
				c.Printf("Plex Sessions sent to %s, sending again in %s, reply: %s", c.URL, c.Plex.Interval, fields[3])
			} else {
				c.Printf("Plex Sessions sent to %s, sending again in %s, reply: %s", c.URL, c.Plex.Interval, string(body))
			}
		case <-c.stopPlex:
			return
		}
	}
}
