package notifiarr

import (
	"time"
)

const cfSyncTimer = 30 * time.Minute

func (c *Config) startTimers() {
	if c.stopTimers != nil {
		return // Already running.
	}

	c.stopTimers = make(chan struct{})
	snapTimer := c.getSnapTimer()
	syncTimer := c.getSyncTimer()
	plexTimer1, plexTimer2 := c.getPlexTimers()

	go c.runTimerLoop(snapTimer, syncTimer, plexTimer1, plexTimer2)
}

func (c *Config) getSyncTimer() *time.Ticker {
	ci, err := c.GetClientInfo()
	if err != nil || !ci.IsASub() || (ci.Message.CFSync < 1 && ci.Message.QRSync < 1) {
		return &time.Ticker{C: make(<-chan time.Time)}
	}

	c.Printf("==> Keeping %d Radarr Custom Formats and %d Sonarr Release Profiles synced",
		ci.Message.CFSync, ci.Message.QRSync)

	return time.NewTicker(cfSyncTimer)
}

func (c *Config) getPlexTimers() (*time.Ticker, *time.Ticker) {
	empty := &time.Ticker{C: make(<-chan time.Time)}

	if !c.Plex.Configured() || c.Plex.Interval.Duration < 1 {
		return empty, empty
	}

	// Add a little splay to the timers to not hit plex at the same time too often.
	plexTimer1 := time.NewTicker(c.Plex.Interval.Duration + 139*time.Millisecond)
	c.Printf("==> Plex Sessions Collection Started, URL: %s, interval: %v, timeout: %v, webhook cooldown: %v",
		c.Plex.URL, c.Plex.Interval, c.Plex.Timeout, c.Plex.Cooldown)

	if c.Plex.MoviesPC != 0 || c.Plex.SeriesPC != 0 {
		c.Printf("==> Plex Completed Items Started, URL: %s, interval: 1m, timeout: %v movies: %d%%, series: %d%%",
			c.Plex.URL, c.Plex.Timeout, c.Plex.MoviesPC, c.Plex.SeriesPC)
		return plexTimer1, time.NewTicker(time.Minute + 179*time.Millisecond)
	}

	return plexTimer1, empty
}

func (c *Config) getSnapTimer() *time.Ticker {
	if c.Snap.Interval.Duration < 1 {
		return &time.Ticker{C: make(<-chan time.Time)}
	}

	c.logSnapshotStartup()

	return time.NewTicker(c.Snap.Interval.Duration)
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
			c.SyncSonarrCF()
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
