// Pcakage triggers provides a simple interface to setup all sub-module triggers.
// Adding a new trigger here should be two new lines of code and a new import.
package triggers

import (
	"context"
	"os"
	"reflect"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/autoupdate"
	"github.com/Notifiarr/notifiarr/pkg/triggers/backups"
	"github.com/Notifiarr/notifiarr/pkg/triggers/cfsync"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/crontimer"
	"github.com/Notifiarr/notifiarr/pkg/triggers/dashboard"
	"github.com/Notifiarr/notifiarr/pkg/triggers/emptytrash"
	"github.com/Notifiarr/notifiarr/pkg/triggers/endpoints"
	"github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig"
	"github.com/Notifiarr/notifiarr/pkg/triggers/fileupload"
	"github.com/Notifiarr/notifiarr/pkg/triggers/filewatch"
	"github.com/Notifiarr/notifiarr/pkg/triggers/gaps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/mdblist"
	"github.com/Notifiarr/notifiarr/pkg/triggers/plexcron"
	"github.com/Notifiarr/notifiarr/pkg/triggers/snapcron"
	"github.com/Notifiarr/notifiarr/pkg/triggers/starrqueue"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/go-co-op/gocron/v2"
)

// Config is the required input data. Everything is mandatory.
type Config struct {
	Apps       *apps.Apps
	Snapshot   *snapshot.Config
	WatchFiles []*filewatch.WatchFile
	Endpoints  []*epconfig.Endpoint
	LogFiles   []string
	Commands   []*commands.Command
	ClientInfo *clientinfo.Config
	ConfigFile string
	AutoUpdate string
	UnstableCh bool
	common.Services
}

// Actions defines all our triggers and timers.
// Any action here will automatically have its interface methods called.
type Actions struct {
	*common.Config
	AutoUpdate *autoupdate.Action
	Backups    *backups.Action
	CFSync     *cfsync.Action
	Commands   *commands.Action
	CronTimer  *crontimer.Action
	Dashboard  *dashboard.Action
	EmptyTrash *emptytrash.Action
	FileUpload *fileupload.Action
	FileWatch  *filewatch.Action
	Gaps       *gaps.Action
	MDbList    *mdblist.Action
	Endpoints  *endpoints.Action
	PlexCron   *plexcron.Action
	SnapCron   *snapcron.Action
	StarrQueue *starrqueue.Action
}

// New turns a populated Config into a pile of Actions.
func New(config *Config) *Actions {
	common := &common.Config{
		Snapshot: config.Snapshot,
		Apps:     config.Apps,
		CI:       config.ClientInfo,
		Services: config.Services,
	}
	common.Scheduler, _ = gocron.NewScheduler()
	plex := plexcron.New(common, &config.Apps.Plex)

	return &Actions{
		AutoUpdate: autoupdate.New(common, config.AutoUpdate, config.ConfigFile, config.UnstableCh),
		Backups:    backups.New(common),
		CFSync:     cfsync.New(common),
		Commands:   commands.New(common, config.Commands),
		Config:     common,
		CronTimer:  crontimer.New(common),
		Dashboard:  dashboard.New(common, plex),
		EmptyTrash: emptytrash.New(common),
		FileUpload: fileupload.New(common),
		FileWatch:  filewatch.New(common, config.WatchFiles, config.LogFiles),
		Gaps:       gaps.New(common),
		MDbList:    mdblist.New(common),
		Endpoints:  endpoints.New(common, config.Endpoints),
		PlexCron:   plex,
		SnapCron:   snapcron.New(common),
		StarrQueue: starrqueue.New(common),
	}
}

// These methods use reflection so they never really need to be updated.
// They execute all Create(), Run() and Stop() procedures defined in our Actions.

// Start creates all the triggers and runs the timers.
func (a *Actions) Start(ctx context.Context, reloadCh, stopCh chan os.Signal) {
	if reloadCh != nil {
		a.SetReloadCh(reloadCh)
	}

	if stopCh != nil {
		a.SetStopCh(stopCh)
	}

	defer a.Run(ctx)

	actions := reflect.ValueOf(a).Elem()
	for idx := range actions.NumField() {
		if !actions.Field(idx).CanInterface() {
			continue
		}

		// A panic here means you screwed up the code somewhere else.
		if action, ok := actions.Field(idx).Interface().(common.Create); ok {
			action.Create()
		}
		// No 'else if' so you can have both if you need them.
		if action, ok := actions.Field(idx).Interface().(common.Run); ok {
			go action.Run(ctx)
		}
	}
}

// Stop all internal cron timers and Triggers.
func (a *Actions) Stop(event website.EventType) context.Context {
	ctx := a.Config.Stop(event)

	actions := reflect.ValueOf(a).Elem()
	// Stop them in reverse order they were started.
	for i := range actions.NumField() {
		if !actions.Field(i).CanInterface() {
			continue
		}

		if action, ok := actions.Field(i).Interface().(common.Run); ok {
			action.Stop()
		}
	}

	return ctx
}
