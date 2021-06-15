package notifiarr

import (
	"time"
)

const cfSyncTimer = 30 * time.Minute

func (c *Config) startTimers() {
	if c.stopTimers != nil {
		return // Already running.
	}

	var (
		snapTimer  = &time.Ticker{C: make(<-chan time.Time)}
		syncTimer  = time.NewTicker(cfSyncTimer)
		plexTimer1 = &time.Ticker{C: make(<-chan time.Time)}
		plexTimer2 = &time.Ticker{C: make(<-chan time.Time)}
	)

	c.stopTimers = make(chan struct{})

	if c.Snap.Interval.Duration > 0 {
		snapTimer = time.NewTicker(c.Snap.Interval.Duration)
		c.logSnapshotStartup()
	}

	if c.Plex != nil && c.Plex.Interval.Duration > 0 && c.Plex.URL != "" && c.Plex.Token != "" {
		// Add a little splay to the timers to not hit plex at the same time too often.
		plexTimer1 = time.NewTicker(c.Plex.Interval.Duration + 139*time.Millisecond)

		c.Printf("==> Plex Sessions Collection Started, URL: %s, interval: %v, timeout: %v, webhook cooldown: %v",
			c.Plex.URL, c.Plex.Interval, c.Plex.Timeout, c.Plex.Cooldown)

		if c.Plex.MoviesPC != 0 || c.Plex.SeriesPC != 0 {
			plexTimer2 = time.NewTicker(time.Minute + 179*time.Millisecond)

			c.Printf("==> Plex Completed Items Started, URL: %s, interval: 1m, timeout: %v movies: %d%%, series: %d%%",
				c.Plex.URL, c.Plex.Timeout, c.Plex.MoviesPC, c.Plex.SeriesPC)
		}
	}

	go c.runTimerLoop(snapTimer, syncTimer, plexTimer1, plexTimer2)
}

func (c *Config) runTimerLoop(snapTimer, syncTimer, plexTimer1, plexTimer2 *time.Ticker) {
	defer func() {
		snapTimer.Stop()
		syncTimer.Stop()
		plexTimer1.Stop()
		plexTimer2.Stop()
		close(c.stopTimers)
		c.stopTimers = nil
	}()

	sent := make(map[string]struct{})

	for {
		select {
		case <-syncTimer.C:
			c.SyncRadarrCF()
			// c.SyncSonarrCF() // later...
		case <-snapTimer.C:
			c.sendSnapshot()
		case <-plexTimer1.C:
			c.sendPlexSessions()
		case <-plexTimer2.C:
			c.checkForFinishedItems(sent)
		case <-c.stopTimers:
			return
		}
	}
}
