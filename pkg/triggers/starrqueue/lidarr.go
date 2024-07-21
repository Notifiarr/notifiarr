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

const TrigLidarrQueue common.TriggerName = "Storing Lidarr instance %d queue."

// StoreLidarr fetches and stores the Lidarr queue immediately for the specified instance.
// Does not send data to the website.
func (a *Action) StoreLidarr(event website.EventType, instance int) {
	if name := TrigLidarrQueue.WithInstance(instance); !a.cmd.Exec(&common.ActionInput{Type: event}, name) {
		a.cmd.Errorf("[%s requested] Failed! %s Disabled?", event, name)
	}
}

type lidarrApp struct {
	app *apps.LidarrConfig
	cmd *cmd
	idx int
}

// storeQueue runs at an interval and saves the queue for an app internally.
func (app *lidarrApp) storeQueue(ctx context.Context, input *common.ActionInput) {
	queue, err := app.app.GetQueueContext(ctx, queueItemsMax, 1)
	if err != nil {
		app.cmd.Errorf("[%s requested] Getting Lidarr Queue (instance %d): %v", input.Type, app.idx+1, err)
		return
	}

	for _, record := range queue.Records {
		record.Quality = nil
	}

	app.cmd.Debugf("[%s requested] Stored Lidarr Queue (%d items), instance %d %s",
		input.Type, len(queue.Records), app.idx+1, app.app.Name)
	data.SaveWithID("lidarr", app.idx, queue)
}

func (c *cmd) setupLidarr() bool {
	var enabled bool

	for idx, app := range c.Apps.Lidarr {
		info := clientinfo.Get()
		if !app.Enabled() || info == nil {
			continue
		}

		var dur time.Duration

		instance := idx + 1
		if info.Actions.Apps.Lidarr.Finished(instance) {
			dur = finishedDuration
		} else if info.Actions.Apps.Lidarr.Stuck(instance) {
			dur = stuckDuration
		}

		if dur != 0 {
			enabled = true

			c.Add(&common.Action{
				Hide: true,
				Name: TrigLidarrQueue.WithInstance(instance),
				Fn:   (&lidarrApp{app: app, cmd: c, idx: idx}).storeQueue,
				C:    make(chan *common.ActionInput, 1),
				D:    cnfg.Duration{Duration: dur},
			})
		}
	}

	return enabled
}
