package starrqueue

import (
	"context"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
)

const TrigReadarrQueue common.TriggerName = "Storing Readarr instance %d queue."

type readarrApp struct {
	app *apps.Readarr
	cmd *cmd
	idx int
}

// storeQueue runs at an interval and saves the queue for an app internally.
func (app *readarrApp) storeQueue(ctx context.Context, input *common.ActionInput) {
	logs.Log.Trace(input.ReqID, "start: readarrApp.storeQueue", app.idx, app.app.Name, input.Type)
	defer logs.Log.Trace(input.ReqID, "end: readarrApp.storeQueue", app.idx, app.app.Name, input.Type)

	queue, err := app.app.GetQueueContext(ctx, queueItemsMax, 1)
	if err != nil {
		mnd.Log.Errorf(input.ReqID, "[%s requested] Getting Readarr Queue (instance %d): %v", input.Type, app.idx+1, err)
		return
	}

	for _, record := range queue.Records {
		record.Quality = nil
	}

	mnd.Log.Printf(input.ReqID, "[%s requested] Stored Readarr Queue (%d items), instance %d %s",
		input.Type, len(queue.Records), app.idx+1, app.app.Name)
	data.SaveWithID("readarr", app.idx, queue)
}

func (c *cmd) setupReadarr(reqID string) bool {
	logs.Log.Trace(reqID, "start: setupReadarr")
	defer logs.Log.Trace(reqID, "end: setupReadarr")

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
			Key:  "TrigReadarrQueue",
			Hide: true,
			Name: TrigReadarrQueue.WithInstance(instance),
			Fn:   (&readarrApp{app: &app, cmd: c, idx: idx}).storeQueue,
			C:    make(chan *common.ActionInput, 1),
			D:    cnfg.Duration{Duration: dur},
		})
	}

	return enable
}
