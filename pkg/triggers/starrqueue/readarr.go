package starrqueue

import (
	"context"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
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
	if name := TrigReadarrQueue.WithInstance(instance); !a.cmd.Exec(event, name) {
		a.cmd.Errorf("Failed! %s Disbled?", name)
	}
}

// storeQueue runs at an interval and saves the queue for an app internally.
func (app *readarrApp) storeQueue(ctx context.Context, event website.EventType) {
	queue, err := app.app.GetQueueContext(ctx, queueItemsMax, 1)
	if err != nil {
		app.cmd.Errorf("Getting Readarr Queue (instance %d): %v", app.idx+1, err)
		return
	}

	for _, record := range queue.Records {
		record.Quality = nil
	}

	app.cmd.Debugf("Stored Readarr Queue (%d items), instance %d %s", len(queue.Records), app.idx+1, app.app.Name)
	data.SaveWithID("readarr", app.idx, queue)
}

func (c *cmd) setupReadarr() bool {
	var enable bool

	for idx, app := range c.Apps.Readarr {
		ci := website.GetClientInfo()
		if !app.Enabled() || ci == nil {
			continue
		}

		var ticker *time.Ticker

		instance := idx + 1

		switch {
		case ci.Actions.Apps.Readarr.Finished(instance):
			enable = true
			ticker = time.NewTicker(finishedDuration)
		case ci.Actions.Apps.Readarr.Stuck(instance):
			enable = true
			ticker = time.NewTicker(stuckDuration)
		default:
			continue
		}

		c.Add(&common.Action{
			Hide: true,
			Name: TrigReadarrQueue.WithInstance(instance),
			Fn:   (&readarrApp{app: app, cmd: c, idx: idx}).storeQueue,
			C:    make(chan website.EventType, 1),
			T:    ticker,
		})
	}

	return enable
}
