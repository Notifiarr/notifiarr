package starrqueue

import (
	"context"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/radarr"
)

const TrigRadarrQueue common.TriggerName = "Storing Radarr instance %d queue."

// StoreRadarr fetches and stores the Radarr queue immediately for the specified instance.
// Does not send data to the website.
func (a *Action) StoreRadarr(event website.EventType, instance int) {
	if name := TrigRadarrQueue.WithInstance(instance); !a.cmd.Exec(event, name) {
		a.cmd.Errorf("Failed! %s Disbled?", name)
	}
}

type radarrApp struct {
	app *apps.RadarrConfig
	cmd *cmd
	idx int
}

// storeQueue runs at an interval and saves the queue for an app internally.
func (app *radarrApp) storeQueue(ctx context.Context, event website.EventType) {
	queue, err := app.app.GetQueueContext(ctx, queueItemsMax, 1)
	if err != nil {
		app.cmd.Errorf("Getting Radarr Queue (instance %d): %v", app.idx+1, err)
		return
	}

	for _, item := range queue.Records {
		item.Quality = nil
		item.CustomFormats = nil
		item.Languages = nil
	}

	app.cmd.Debugf("Stored Radarr Queue (%d items), instance %d %s", len(queue.Records), app.idx+1, app.app.Name)
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
				C:    make(chan website.EventType, 1),
				T:    ticker,
			})
		}
	}

	return enabled
}

func (c *cmd) getFinishedItemsRadarr(_ context.Context) itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Radarr {
		ci := website.GetClientInfo()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Radarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("radarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*radarr.Queue)
		instance := idx + 1
		stuckapp := stuck[instance]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = c.Apps.Radarr[idx].Name // this should be safe.
		stuck[instance] = stuckapp

		c.Debugf("Checking Radarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}
