package starrqueue

import (
	"context"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/lidarr"
)

const TrigLidarrQueue common.TriggerName = "Storing Lidarr instance %d queue."

// StoreLidarr fetches and stores the Lidarr queue immediately for the specified instance.
// Does not send data to the website.
func (a *Action) StoreLidarr(event website.EventType, instance int) {
	if name := TrigLidarrQueue.WithInstance(instance); !a.cmd.Exec(event, name) {
		a.cmd.Errorf("Failed! %s Disbled?", name)
	}
}

type lidarrApp struct {
	app *apps.LidarrConfig
	cmd *cmd
	idx int
}

// storeQueue runs at an interval and saves the queue for an app internally.
func (app *lidarrApp) storeQueue(ctx context.Context, event website.EventType) {
	queue, err := app.app.GetQueueContext(ctx, queueItemsMax, 1)
	if err != nil {
		app.cmd.Errorf("Getting Lidarr Queue (instance %d): %v", app.idx+1, err)
		return
	}

	for _, record := range queue.Records {
		record.Quality = nil
	}

	app.cmd.Debugf("Stored Lidarr Queue (%d items), instance %d %s", len(queue.Records), app.idx+1, app.app.Name)
	data.SaveWithID("lidarr", app.idx, queue)
}

func (c *cmd) setupLidarr() bool {
	var enabled bool

	for idx, app := range c.Apps.Lidarr {
		ci := website.GetClientInfo()
		if !app.Enabled() || ci == nil {
			continue
		}

		var ticker *time.Ticker

		instance := idx + 1
		if ci.Actions.Apps.Lidarr.Finished(instance) {
			ticker = time.NewTicker(finishedDuration)
		} else if ci.Actions.Apps.Lidarr.Stuck(instance) {
			ticker = time.NewTicker(stuckDuration)
		}

		if ticker != nil {
			enabled = true

			c.Add(&common.Action{
				Hide: true,
				Name: TrigLidarrQueue.WithInstance(instance),
				Fn:   (&lidarrApp{app: app, cmd: c, idx: idx}).storeQueue,
				C:    make(chan website.EventType, 1),
				T:    ticker,
			})
		}
	}

	return enabled
}

func (c *cmd) getFinishedItemsLidarr(_ context.Context) itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Lidarr {
		ci := website.GetClientInfo()
		if !app.Enabled() || ci == nil || !ci.Actions.Apps.Lidarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("lidarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*lidarr.Queue)
		instance := idx + 1
		stuckapp := stuck[instance]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = c.Apps.Lidarr[idx].Name // this should be safe.
		stuck[instance] = stuckapp

		c.Debugf("Checking Lidarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}
