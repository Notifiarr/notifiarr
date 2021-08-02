package notifiarr

import (
	"time"
)

const (
	cfSyncTimer  = 30 * time.Minute
	minDashTimer = 30 * time.Minute
	stuckTimer   = 5*time.Minute + 327*time.Millisecond
)

func (c *Config) startTimers() {
	if c.Trigger.stop != nil {
		return // Already running.
	}

	c.Trigger.stop = make(chan struct{})
	snapTimer := c.getSnapTimer()
	syncTimer := c.getSyncTimer()
	plexTimer1, plexTimer2 := c.getPlexTimers()
	stuckItemTimer := time.NewTicker(stuckTimer)
	dashTimer := c.getDashTimer()

	go c.runTimerLoop(snapTimer, syncTimer, plexTimer1, plexTimer2, stuckItemTimer, dashTimer)
}

func (c *Config) getDashTimer() *time.Ticker {
	dashTimer := &time.Ticker{}

	if c.DashDur > 0 && c.DashDur < minDashTimer {
		c.DashDur = minDashTimer
	}

	if c.DashDur > 0 {
		c.Printf("==> Sending Current State Data for Dashboard every %v", c.DashDur)
		dashTimer = time.NewTicker(c.DashDur)
	}

	return dashTimer
}

func (c *Config) getSyncTimer() *time.Ticker {
	ci, err := c.GetClientInfo()
	if err != nil || (ci.Message.CFSync < 1 && ci.Message.RPSync < 1) {
		return &time.Ticker{C: make(<-chan time.Time)}
	}

	c.Printf("==> Keeping %d Radarr Custom Formats and %d Sonarr Release Profiles synced",
		ci.Message.CFSync, ci.Message.RPSync)

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

// runTimerLoop does all of the timer/cron routines for starr apps and plex.
// Many of the menu items and trigger handlers feed into this routine too.
// nolint:cyclop
func (c *Config) runTimerLoop(snapTimer, syncTimer, plexTimer1, plexTimer2, stuckTimer, dashTimer *time.Ticker) {
	defer c.stopTimerLoop(snapTimer, syncTimer, plexTimer1, plexTimer2, stuckTimer, dashTimer)

	for sent := make(map[string]struct{}); ; {
		select {
		case reply := <-c.Trigger.syncCF:
			c.syncCF(reply)
		case <-syncTimer.C:
			c.syncCF(nil)
		case source := <-c.Trigger.snap:
			c.sendSnapshot(source)
		case <-snapTimer.C:
			c.sendSnapshot(SnapCron)
		case source := <-c.Trigger.plex:
			c.sendPlexSessions(source)
		case <-plexTimer1.C:
			c.sendPlexSessions(PlexCron)
		case <-stuckTimer.C:
			c.sendFinishedQueueItems(c.BaseURL)
		case url := <-c.Trigger.stuck:
			c.sendFinishedQueueItems(url)
		case <-plexTimer2.C:
			c.checkForFinishedItems(sent)
		case <-c.Trigger.state:
			c.Print("API Trigger: Gathering current state for dashboard.")
			c.getState()
		case <-dashTimer.C:
			c.Print("Gathering current state for dashboard.")
			c.getState()
		case <-c.Trigger.stop:
			return
		}
	}
}

// stopTimerLoop is defered by runTimerLoop.
func (c *Config) stopTimerLoop(timers ...*time.Ticker) {
	c.CapturePanic()

	for _, timer := range timers {
		timer.Stop()
	}

	close(c.Trigger.stop)
	c.Trigger.stop = nil
}
