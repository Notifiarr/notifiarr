// Pcakage triggers provides a simple interface to setup all sub-module triggers.
// Adding a new trigger here should be two new lines of code and a new import.
package triggers

import (
	"context"
	"os"
	"reflect"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/backups"
	"github.com/Notifiarr/notifiarr/pkg/triggers/cfsync"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/crontimer"
	"github.com/Notifiarr/notifiarr/pkg/triggers/dashboard"
	"github.com/Notifiarr/notifiarr/pkg/triggers/emptytrash"
	"github.com/Notifiarr/notifiarr/pkg/triggers/fileupload"
	"github.com/Notifiarr/notifiarr/pkg/triggers/filewatch"
	"github.com/Notifiarr/notifiarr/pkg/triggers/gaps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/mdblist"
	"github.com/Notifiarr/notifiarr/pkg/triggers/plexcron"
	"github.com/Notifiarr/notifiarr/pkg/triggers/snapcron"
	"github.com/Notifiarr/notifiarr/pkg/triggers/starrqueue"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
)

// Config is the required input data. Everything is mandatory.
type Config struct {
	Apps       *apps.Apps
	Website    *website.Server
	Snapshot   *snapshot.Config
	WatchFiles []*filewatch.WatchFile
	LogFiles   []string
	Commands   []*commands.Command
	CIC        *clientinfo.Config
	common.Services
	*logs.Logger
}

// Actions defines all our triggers and timers.
// Any action here will automatically have its interface methods called.
type Actions struct {
	Timers *common.Config
	// Order is important here.
	PlexCron   *plexcron.Action
	Backups    *backups.Action
	CFSync     *cfsync.Action
	CronTimer  *crontimer.Action
	Dashboard  *dashboard.Action
	FileWatch  *filewatch.Action
	Gaps       *gaps.Action
	SnapCron   *snapcron.Action
	StarrQueue *starrqueue.Action
	Commands   *commands.Action
	EmptyTrash *emptytrash.Action
	MDbList    *mdblist.Action
	FileUpload *fileupload.Action
}

// New turns a populated Config into a pile of Actions.
func New(config *Config) *Actions {
	common := &common.Config{
		Server:   config.Website,
		Snapshot: config.Snapshot,
		Apps:     config.Apps,
		Logger:   config.Logger,
		CIC:      config.CIC,
		Services: config.Services,
	}
	plex := plexcron.New(common, config.Apps.Plex)

	return &Actions{
		PlexCron:   plex,
		Backups:    backups.New(common),
		CFSync:     cfsync.New(common),
		CronTimer:  crontimer.New(common),
		Dashboard:  dashboard.New(common, plex),
		FileWatch:  filewatch.New(common, config.WatchFiles, config.LogFiles),
		Gaps:       gaps.New(common),
		SnapCron:   snapcron.New(common),
		StarrQueue: starrqueue.New(common),
		Commands:   commands.New(common, config.Commands),
		EmptyTrash: emptytrash.New(common),
		MDbList:    mdblist.New(common),
		FileUpload: fileupload.New(common),
		Timers:     common,
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
func (a *Actions) Start(ctx context.Context, reloadCh chan os.Signal) {
	a.Timers.SetReloadCh(reloadCh)
	defer a.Timers.Run(ctx)

	actions := reflect.ValueOf(a).Elem()
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
func (a *Actions) Stop(event website.EventType) {
	a.Timers.Stop(event)

	actions := reflect.ValueOf(a).Elem()
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
