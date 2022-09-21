package starrqueue

import (
	"context"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const TrigRadarrQueue common.TriggerName = "Storing Radarr instance %d queue."

// StoreRadarr fetches and stores the Radarr queue immediately for the specified instance.
// Does not send data to the website.
func (a *Action) StoreRadarr(event website.EventType, instance int) {
	if name := TrigRadarrQueue.WithInstance(instance); !a.cmd.Exec(&common.ActionInput{Type: event}, name) {
		a.cmd.Errorf("[%s requested] Failed! %s Disbled?", event, name)
	}
}

type radarrApp struct {
	app *apps.RadarrConfig
	cmd *cmd
	idx int
}

// storeQueue runs at an interval and saves the queue for an app internally.
func (app *radarrApp) storeQueue(ctx context.Context, input *common.ActionInput) {
	queue, err := app.app.GetQueueContext(ctx, queueItemsMax, 1)
	if err != nil {
		app.cmd.Errorf("[%s requested] Getting Radarr Queue (instance %d): %v", input.Type, app.idx+1, err)
		return
	}

	for _, item := range queue.Records {
		item.Quality = nil
		item.CustomFormats = nil
		item.Languages = nil
	}

	app.cmd.Debugf("[%s requested] Stored Radarr Queue (%d items), instance %d %s",
		input.Type, len(queue.Records), app.idx+1, app.app.Name)
	data.SaveWithID("radarr", app.idx, queue)
}

func (c *cmd) setupRadarr() bool {
	var enabled bool

	for idx, app := range c.Apps.Radarr {
		ci := website.GetClientInfo()
		if !app.Enabled() || ci == nil {
			continue
		}

		var ticker *time.Ticker

		instance := idx + 1
		if ci.Actions.Apps.Radarr.Finished(instance) {
			ticker = time.NewTicker(finishedDuration)
		} else if ci.Actions.Apps.Radarr.Stuck(instance) {
			ticker = time.NewTicker(stuckDuration)
		}

		if ticker != nil {
			enabled = true

			c.Add(&common.Action{
				Hide: true,
				Name: TrigRadarrQueue.WithInstance(instance),
				Fn:   (&radarrApp{app: app, cmd: c, idx: idx}).storeQueue,
				C:    make(chan *common.ActionInput, 1),
				T:    ticker,
			})
		}
	}

	return enabled
}
