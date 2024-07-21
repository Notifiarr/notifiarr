package starrqueue

import (
	"context"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
)

const TrigReadarrQueue common.TriggerName = "Storing Readarr instance %d queue."

type readarrApp struct {
	app *apps.ReadarrConfig
	cmd *cmd
	idx int
}

// StoreReadarr fetches and stores the Readarr queue immediately for the specified instance.
// Does not send data to the website.
func (a *Action) StoreReadarr(event website.EventType, instance int) {
	if name := TrigReadarrQueue.WithInstance(instance); !a.cmd.Exec(&common.ActionInput{Type: event}, name) {
		a.cmd.Errorf("[%s requested] Failed! %s Disabled?", event, name)
	}
}

// storeQueue runs at an interval and saves the queue for an app internally.
func (app *readarrApp) storeQueue(ctx context.Context, input *common.ActionInput) {
	queue, err := app.app.GetQueueContext(ctx, queueItemsMax, 1)
	if err != nil {
		app.cmd.Errorf("[%s requested] Getting Readarr Queue (instance %d): %v", input.Type, app.idx+1, err)
		return
	}

	for _, record := range queue.Records {
		record.Quality = nil
	}

	app.cmd.Debugf("[%s requested] Stored Readarr Queue (%d items), instance %d %s",
		input.Type, len(queue.Records), app.idx+1, app.app.Name)
	data.SaveWithID("readarr", app.idx, queue)
}

func (c *cmd) setupReadarr() bool {
	var enable bool

	for idx, app := range c.Apps.Readarr {
		info := clientinfo.Get()
		if !app.Enabled() || info == nil {
			continue
		}

		var dur time.Duration

		instance := idx + 1

		switch {
		case info.Actions.Apps.Readarr.Finished(instance):
			enable = true
			dur = finishedDuration
		case info.Actions.Apps.Readarr.Stuck(instance):
			enable = true
			dur = stuckDuration
		default:
			continue
		}

		c.Add(&common.Action{
			Hide: true,
			Name: TrigReadarrQueue.WithInstance(instance),
			Fn:   (&readarrApp{app: app, cmd: c, idx: idx}).storeQueue,
			C:    make(chan *common.ActionInput, 1),
			D:    cnfg.Duration{Duration: dur},
		})
	}

	return enable
}
