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

type Actions struct {
	common     *common.Config
	clientInfo *website.ClientInfo
	Backups    *backups.Config
	CronTimer  *crontimer.Config
	PlexCron   *plexcron.Config
	CFSync     *cfsync.Config
	SnapCron   *snapcron.Config
	Dashboard  *dashboard.Config
	Gaps       *gaps.Config
	StuckItems *stuckitems.Config
	FileWatch  *filewatch.Config
}

type Config struct {
	Serial     bool
	Apps       *apps.Apps
	Plex       *plex.Server
	Website    *website.Server
	Snapshot   *snapshot.Config
	WatchFiles []*filewatch.WatchFile
	mnd.Logger
}

func New(config *Config) *Actions {
	var (
		ci, _  = config.Website.GetClientInfo()
		common = &common.Config{
			Server:     config.Website,
			ClientInfo: ci,
			Snapshot:   config.Snapshot,
			Apps:       config.Apps,
			Serial:     config.Serial,
			Logger:     config.Logger,
		}
		plex = &plexcron.Config{Config: common, Plex: config.Plex}
	)

	return &Actions{
		common:     common,
		clientInfo: ci,
		Backups:    &backups.Config{Config: common},
		CronTimer:  &crontimer.Config{Config: common},
		PlexCron:   plex,
		CFSync:     &cfsync.Config{Config: common},
		SnapCron:   &snapcron.Config{Config: common},
		Dashboard:  &dashboard.Config{Config: common, PlexCron: plex},
		Gaps:       &gaps.Config{Config: common},
		StuckItems: &stuckitems.Config{Config: common},
		FileWatch: &filewatch.Config{
			Config:     common,
			WatchFiles: config.WatchFiles,
		},
	}
}

// Start runs the timers.
func (c *Actions) Start() {
	// Order may be important here.
	c.PlexCron.Create()
	c.StuckItems.Create()
	c.SnapCron.Create()
	c.Gaps.Create()
	c.CFSync.Create()
	c.Dashboard.Create()
	c.Backups.Create()
	c.CronTimer.Create()
	c.common.RunTimers()
	c.FileWatch.Run()
}

// Stop all internal cron timers and Triggers.
func (c *Actions) Stop(event website.EventType) {
	// This closes runTimerLoop() and fires stopTimerLoop().
	c.common.Close(event)
	// Stops the file and log watchers.
	c.FileWatch.Stop()
	// Closes the Plex session holder.
	c.PlexCron.Close()
}
