package notifiarr

import (
	"time"
)

const (
	stuckTimer = 5*time.Minute + 1327*time.Millisecond
)

func (c *Config) startTimers() {
	if c.Trigger.stop != nil {
		return // Already running.
	}

	c.Trigger.stop = make(chan struct{})
	snapTimer := c.getSnapTimer()
	plexTimer1, plexTimer2 := c.getPlexTimers()
	stuckItemTimer := time.NewTicker(stuckTimer)

	syncTimer := &time.Ticker{C: make(<-chan time.Time)}
	cronTimer := &time.Ticker{C: make(<-chan time.Time)}
	gapsTimer := &time.Ticker{C: make(<-chan time.Time)}
	dashTimer := &time.Ticker{C: make(<-chan time.Time)}

	if ci, err := c.GetClientInfo(); err == nil {
		syncTimer, cronTimer, gapsTimer, dashTimer = c.getClientInfoTimers(ci)
	}

	go c.runTimerLoop(snapTimer, syncTimer, plexTimer1, plexTimer2, stuckItemTimer, dashTimer, cronTimer, gapsTimer)
}

func (c *Config) getClientInfoTimers(ci *ClientInfo) (*time.Ticker, *time.Ticker, *time.Ticker, *time.Ticker) {
	cronTimer := &time.Ticker{C: make(<-chan time.Time)}
	gapsTimer := &time.Ticker{C: make(<-chan time.Time)}
	syncTimer := &time.Ticker{C: make(<-chan time.Time)}
	dashTimer := &time.Ticker{C: make(<-chan time.Time)}

	if len(ci.Actions.Custom) > 0 {
		c.Printf("==> Custom Timers Enabled: %d timers provided", len(ci.Actions.Custom))

		cronTimer = time.NewTicker(time.Minute)
	}

	if ci.Actions.Gaps.Minutes > 0 {
		c.Printf("==> Collection Gaps Timer Enabled, interval: %dm", ci.Actions.Gaps.Minutes)

		gapsTimer = time.NewTicker(time.Minute * time.Duration(ci.Actions.Gaps.Minutes))
	}

	if ci.Actions.Sync.Minutes > 0 {
		c.Printf("==> Keeping %d Radarr Custom Formats and %d Sonarr Release Profiles synced, interval: %dm",
			ci.Actions.Sync.Radarr, ci.Actions.Sync.Sonarr, ci.Actions.Sync.Minutes)

		syncTimer = time.NewTicker(time.Minute * time.Duration(ci.Actions.Sync.Minutes))
	}

	if ci.Actions.Dashboard.Minutes > 0 {
		c.Printf("==> Sending Current State Data for Dashboard every %dm", ci.Actions.Dashboard.Minutes)
		dashTimer = time.NewTicker(time.Minute * time.Duration(ci.Actions.Dashboard.Minutes))
	}

	return syncTimer, cronTimer, gapsTimer, dashTimer
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
func (c *Config) runTimerLoop(snapTimer, syncTimer, plexTimer1, plexTimer2,
	stuckTimer, dashTimer, cronTimer, gapsTimer *time.Ticker) {
	defer c.stopTimerLoop(snapTimer, syncTimer, plexTimer1, plexTimer2, stuckTimer, dashTimer, cronTimer, gapsTimer)

	for sent := make(map[string]struct{}); ; {
		select {
		case <-cronTimer.C:
			for _, timer := range c.extras.clientInfo.Actions.Custom {
				if timer.Ready() {
					if _, _, err := c.GetData(c.BaseURL + "/" + timer.URI); err != nil {
						c.Errorf("Custom Timer Request for %s failed: %v", timer.URI, err)
					}
					break //nolint:wsl
				}
			}
		case <-gapsTimer.C:
			c.sendGaps("timer")
		case source := <-c.Trigger.gaps:
			c.sendGaps(source)
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
	defer close(c.Trigger.stop)
	c.Trigger.stop = nil

	c.CapturePanic()

	for _, timer := range timers {
		timer.Stop()
	}
}
