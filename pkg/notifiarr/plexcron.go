package notifiarr

import (
	"strings"
	"time"
)

// This cron tab runs every 10-60 minutes to send a report of who's currently watching.
func (c *Config) startPlexCron() {
	if c.Plex == nil || c.Plex.Interval.Duration == 0 || c.Plex.URL == "" || c.Plex.Token == "" {
		return
	}

	// Add a little splay to the timers to not hit plex at the same time too often.
	timer1 := time.NewTicker(c.Plex.Interval.Duration + 139*time.Millisecond)
	timer2 := time.NewTicker(time.Minute + 179*time.Millisecond)
	c.stopPlex = make(chan struct{})

	defer func() {
		timer1.Stop()
		close(c.stopPlex)
		c.stopPlex = nil
	}()

	c.Printf("==> Plex Sessions Collection Started, URL: %s, interval: %v, timeout: %v, cooldown: %v",
		c.Plex.URL, c.Plex.Interval, c.Plex.Timeout, c.Plex.Cooldown)

	if c.Plex.MoviesPC != 0 || c.Plex.SeriesPC != 0 {
		defer timer2.Stop()
		c.Printf("==> Plex Completed Items Started, URL: %s, interval: 1m, timeout: %v",
			c.Plex.URL, c.Plex.Timeout)
	} else {
		timer2.Stop() // nothing to check, so turn off this timer.
	}

	c.plexCron(timer1, timer2)
}

// Do not call this directly. Called from above, only.
func (c *Config) plexCron(timer1, timer2 *time.Ticker) {
	for {
		select {
		case <-timer1.C:
			if body, err := c.SendMeta(PlexCron, c.URL, nil, 0); err != nil {
				c.Errorf("Sending Plex Session to %s: %v: %v", c.URL, err, string(body))
			} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
				c.Printf("Plex Sessions sent to %s, sending again in %s, reply: %s", c.URL, c.Plex.Interval, fields[3])
			} else {
				c.Printf("Plex Sessions sent to %s, sending again in %s, reply: %s", c.URL, c.Plex.Interval, string(body))
			}
		case <-timer2.C:
			c.checkForFinishedItems()
		case <-c.stopPlex:
			return
		}
	}
}

// This cron tab runs every minute to send a report when a user gets to the end of a movie or tv show.
// This is basically a hack to "watch" Plex for when an active item gets to around 90% complete.
// This usually means the user has finished watching the item and we can send a "done" notice.
// Plex does not send a webhook or identify in any other way when an item is "finished".
func (c *Config) checkForFinishedItems() {
	sessions, err := c.Plex.GetSessions()
	if err != nil {
		c.Errorf("[PLEX] Getting Sessions from %s: %v", c.URL, err)
		return
	}

	switch l := len(sessions); l {
	case 0:
		c.Debugf("[PLEX] No Sessions Collected from %s", c.URL)
	case 1:
		c.Debugf("[PLEX] Collected 1 session from %s:", c.URL)
	default:
		c.Debugf("[PLEX] Collected %d sessions from %s:", l, c.URL)
	}

	for _, s := range sessions {
		c.Debugf("[PLEX] %s => %s: %s (%s) %.1f%%",
			s.User.Title, s.Type, s.Title, s.Player.State, s.ViewOffset/s.Duration*100) //nolint:gomnd
	}
}
