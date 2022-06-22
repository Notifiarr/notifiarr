package triggers

import (
	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/backups"
	"github.com/Notifiarr/notifiarr/pkg/triggers/cfsync"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/crontimer"
	"github.com/Notifiarr/notifiarr/pkg/triggers/dashboard"
	"github.com/Notifiarr/notifiarr/pkg/triggers/filewatch"
	"github.com/Notifiarr/notifiarr/pkg/triggers/gaps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/plexcron"
	"github.com/Notifiarr/notifiarr/pkg/triggers/snapcron"
	"github.com/Notifiarr/notifiarr/pkg/triggers/stuckitems"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// Actions defines all our triggers and timers.
type Actions struct {
	timers     *common.Config
	Backups    *backups.Action
	CFSync     *cfsync.Action
	CronTimer  *crontimer.Action
	Dashboard  *dashboard.Action
	FileWatch  *filewatch.Action
	Gaps       *gaps.Action
	PlexCron   *plexcron.Action
	SnapCron   *snapcron.Action
	StuckItems *stuckitems.Action
}

// Config is the required input data. Everything is mandatory.
type Config struct {
	Serial     bool
	Apps       *apps.Apps
	Plex       *plex.Server
	Website    *website.Server
	Snapshot   *snapshot.Config
	WatchFiles []*filewatch.WatchFile
	mnd.Logger
}

// New turns a populated Config into a pile of Actions.
func New(config *Config) *Actions {
	var (
		ci, _  = config.Website.GetClientInfo()
		common = &common.Config{
			ClientInfo: ci,
			Server:     config.Website,
			Snapshot:   config.Snapshot,
			Apps:       config.Apps,
			Serial:     config.Serial,
			Logger:     config.Logger,
		}
		plex = &plexcron.Action{Config: common, Plex: config.Plex}
	)

	return &Actions{
		timers:     common,
		Backups:    &backups.Action{Config: common},
		CFSync:     &cfsync.Action{Config: common},
		CronTimer:  &crontimer.Action{Config: common},
		Dashboard:  &dashboard.Action{Config: common, PlexCron: plex},
		FileWatch:  &filewatch.Action{Config: common, WatchFiles: config.WatchFiles},
		Gaps:       &gaps.Action{Config: common},
		PlexCron:   plex,
		SnapCron:   &snapcron.Action{Config: common},
		StuckItems: &stuckitems.Action{Config: common},
	}
}

// Start creates all the triggers and runs the timers.
func (c *Actions) Start() {
	// Order may be important here.
	c.PlexCron.Run() // must be stopped.
	c.Backups.Create()
	c.CFSync.Create()
	c.CronTimer.Create()
	c.Dashboard.Create()
	c.FileWatch.Run() // must be stopped.
	c.Gaps.Create()
	c.SnapCron.Create()
	c.StuckItems.Create()
	c.timers.Run() // must be stopped.
}

// Stop all internal cron timers and Triggers.
func (c *Actions) Stop(event website.EventType) {
	// This closes runTimerLoop() and fires stopTimerLoop().
	c.timers.Stop(event)
	// Stops the file and log watchers.
	c.FileWatch.Stop()
	// Closes the Plex session holder.
	c.PlexCron.Stop()
}
