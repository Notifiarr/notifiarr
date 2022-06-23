package triggers

import (
	"reflect"

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

// Actions defines all our triggers and timers.
// Any action here will automatically have its interface methods called.
type Actions struct {
	timers *common.Config
	// Order is important here.
	PlexCron   *plexcron.Action
	Backups    *backups.Action
	CFSync     *cfsync.Action
	CronTimer  *crontimer.Action
	Dashboard  *dashboard.Action
	FileWatch  *filewatch.Action
	Gaps       *gaps.Action
	SnapCron   *snapcron.Action
	StuckItems *stuckitems.Action
}

// New turns a populated Config into a pile of Actions.
func New(config *Config) *Actions {
	ci, _ := config.Website.GetClientInfo()
	common := &common.Config{
		ClientInfo: ci,
		Server:     config.Website,
		Snapshot:   config.Snapshot,
		Apps:       config.Apps,
		Serial:     config.Serial,
		Logger:     config.Logger,
	}
	plex := plexcron.New(common, config.Plex)

	return &Actions{
		PlexCron:   plex,
		Backups:    backups.New(common),
		CFSync:     cfsync.New(common),
		CronTimer:  crontimer.New(common),
		Dashboard:  dashboard.New(common, plex),
		FileWatch:  filewatch.New(common, config.WatchFiles),
		Gaps:       gaps.New(common),
		SnapCron:   snapcron.New(common),
		StuckItems: stuckitems.New(common),
		timers:     common,
	}
}

// These methods use reflection so they never really need to be updated.
// They execute all Create(), Run() and Stop() procedures defined in our Actions.

type create interface {
	Create()
}

type run interface {
	Run()
	Stop()
}

// Start creates all the triggers and runs the timers.
func (c *Actions) Start() {
	defer c.timers.Run() // unexported fields do not get picked up by reflection.

	actions := reflect.ValueOf(c).Elem()
	for i := 0; i < actions.NumField(); i++ {
		if !actions.Field(i).CanInterface() {
			continue
		}

		// A panic here means you screwed up the code somewhere else.
		if action, ok := actions.Field(i).Interface().(create); ok {
			action.Create()
		}
		// No 'else if' so you can have both if you need them.
		if action, ok := actions.Field(i).Interface().(run); ok {
			action.Run()
		}
	}
}

// Stop all internal cron timers and Triggers.
func (c *Actions) Stop(event website.EventType) {
	defer c.timers.Stop(event)

	actions := reflect.ValueOf(c).Elem()
	// Stop them in reverse order they were started.
	for i := actions.NumField() - 1; i >= 0; i-- {
		if !actions.Field(i).CanInterface() {
			continue
		}

		if action, ok := actions.Field(i).Interface().(run); ok {
			action.Stop()
		}
	}
}
