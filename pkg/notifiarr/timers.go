package notifiarr

import (
	"time"

	"golift.io/cnfg"
)

const (
	stuckDur = 5*time.Minute + 1327*time.Millisecond
	pollDur  = 4*time.Minute + 47*time.Second + 977*time.Millisecond
)

// timerConfig defines a custom GET timer from the website.
// Used to offload crons to clients.
type timerConfig struct {
	Name     string        `json:"name"`     // name of action.
	Interval cnfg.Duration `json:"interval"` // how often to GET this URI.
	URI      string        `json:"endpoint"` // endpoint for the URI.
	Desc     string        `json:"description"`
	last     time.Time
}

func (t *timerConfig) Ready() bool {
	return t.last.After(time.Now().Add(t.Interval.Duration))
}

func (c *Config) startTimers() {
	if c.Trigger.stop != nil {
		return // Already running.
	}

	c.Trigger.stop = make(chan struct{})
	snapTimer := c.getSnapTimer()
	completedItems, plexSessions := c.getPlexTimers()
	stuckTimer := time.NewTicker(stuckDur)
	pollTimer := time.NewTicker(pollDur)

	syncTimer := &time.Ticker{C: make(<-chan time.Time)}
	cronTimer := &time.Ticker{C: make(<-chan time.Time)}
	gapsTimer := &time.Ticker{C: make(<-chan time.Time)}
	dashTimer := &time.Ticker{C: make(<-chan time.Time)}

	if _, err := c.GetClientInfo(EventStart); err == nil { // gets stored.
		syncTimer, cronTimer, gapsTimer, dashTimer = c.getClientInfoTimers()
	}

	go c.runTimerLoop(snapTimer, syncTimer, completedItems, plexSessions,
		stuckTimer, dashTimer, cronTimer, gapsTimer, pollTimer)
}

func (c *Config) getClientInfoTimers() (*time.Ticker, *time.Ticker, *time.Ticker, *time.Ticker) {
	cronTimer := &time.Ticker{C: make(<-chan time.Time)}
	gapsTimer := &time.Ticker{C: make(<-chan time.Time)}
	syncTimer := &time.Ticker{C: make(<-chan time.Time)}
	dashTimer := &time.Ticker{C: make(<-chan time.Time)}

	if len(c.Actions.Custom) > 0 {
		c.Printf("==> Custom Timers Enabled: %d timers provided", len(c.Actions.Custom))

		cronTimer = time.NewTicker(time.Minute)
	}

	if c.Actions.Gaps.Interval.Duration > 0 {
		c.Printf("==> Collection Gaps Timer Enabled, interval: %s", c.Actions.Gaps.Interval)

		gapsTimer = time.NewTicker(c.Actions.Gaps.Interval.Duration)
	}

	if c.Actions.Sync.Interval.Duration > 0 {
		c.Printf("==> Keeping %d Radarr Custom Formats and %d Sonarr Release Profiles synced, interval: %s",
			c.Actions.Sync.Radarr, c.Actions.Sync.Sonarr, c.Actions.Sync.Interval)

		syncTimer = time.NewTicker(c.Actions.Sync.Interval.Duration)
	}

	if c.Actions.Dashboard.Interval.Duration > 0 {
		c.Printf("==> Sending Current State Data for Dashboard every %s", c.Actions.Dashboard.Interval)
		dashTimer = time.NewTicker(c.Actions.Dashboard.Interval.Duration)
	}

	return syncTimer, cronTimer, gapsTimer, dashTimer
}

func (c *Config) getPlexTimers() (*time.Ticker, *time.Ticker) {
	plexTimer1 := &time.Ticker{C: make(<-chan time.Time)}
	plexTimer2 := &time.Ticker{C: make(<-chan time.Time)}

	if !c.Plex.Configured() {
		return plexTimer1, plexTimer2
	}

	if c.Plex.MoviesPC != 0 || c.Plex.SeriesPC != 0 {
		c.Printf("==> Plex Completed Items Started, URL: %s, interval: 1m, timeout: %v movies: %d%%, series: %d%%",
			c.Plex.URL, c.Plex.Timeout, c.Plex.MoviesPC, c.Plex.SeriesPC)

		plexTimer1 = time.NewTicker(time.Minute + 179*time.Millisecond)
	}

	if c.Plex.Interval.Duration > 0 {
		// Add a little splay to the timers to not hit plex at the same time too often.
		c.Printf("==> Plex Sessions Collection Started, URL: %s, interval: %v, timeout: %v, webhook cooldown: %v",
			c.Plex.URL, c.Plex.Interval, c.Plex.Timeout, c.Plex.Cooldown)

		plexTimer2 = time.NewTicker(c.Plex.Interval.Duration + 139*time.Millisecond)
	}

	return plexTimer1, plexTimer2
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
func (c *Config) runTimerLoop(snapTimer, syncTimer, completedItems, plexSessions,
	stuckTimer, dashTimer, cronTimer, gapsTimer, pollTimer *time.Ticker) {
	defer c.stopTimerLoop(snapTimer, syncTimer, completedItems, plexSessions,
		stuckTimer, dashTimer, cronTimer, gapsTimer, pollTimer)

	for sent := make(map[string]struct{}); ; {
		select {
		case <-c.Trigger.stop:
			return
		case <-gapsTimer.C:
			c.sendGaps(EventCron)
		case event := <-c.Trigger.gaps:
			c.sendGaps(event)
		case event := <-c.Trigger.syncCF:
			c.syncCF(event)
		case <-syncTimer.C:
			c.syncCF(EventCron)
		case event := <-c.Trigger.snap:
			c.sendSnapshot(event)
		case <-snapTimer.C:
			c.sendSnapshot(EventCron)
		case event := <-c.Trigger.plex:
			c.sendPlexSessions(event)
		case <-plexSessions.C:
			c.sendPlexSessions(EventCron)
		case event := <-c.Trigger.stuck:
			c.sendFinishedQueueItems(event)
		case <-completedItems.C:
			c.checkForFinishedItems(sent)
		case event := <-c.Trigger.state:
			c.sendDashboardState(event)
		case <-dashTimer.C:
			c.Print("Gathering current state for dashboard.")
			c.sendDashboardState(EventCron)
		case <-pollTimer.C:
			c.pollForReload()
		case <-stuckTimer.C:
			c.sendFinishedQueueItems(EventCron)
		case <-cronTimer.C:
			c.runCustomTimers()
		}
	}
}

func (c *Config) runCustomTimers() {
	for _, timer := range c.Actions.Custom {
		if !timer.Ready() {
			continue
		}

		c.Printf("Running Custom Cron Timer: %s", timer.Name)

		if _, err := c.GetData(c.BaseURL + "/" + timer.URI); err != nil {
			c.Errorf("Custom Timer Request for %s failed: %v", timer.URI, err)
		}
	}
}

// stopTimerLoop is defered by runTimerLoop.
func (c *Config) stopTimerLoop(timers ...*time.Ticker) {
	defer c.CapturePanic()
	defer close(c.Trigger.stop)
	c.Trigger.stop = nil

	for _, timer := range timers {
		timer.Stop()
	}
}
