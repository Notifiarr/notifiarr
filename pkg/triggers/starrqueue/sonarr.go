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

const TrigSonarrQueue common.TriggerName = "Storing Sonarr instance %d queue."

// StoreSonarr fetches and stores the Sonarr queue immediately for the specified instance.
// Does not send data to the website.
func (a *Action) StoreSonarr(event website.EventType, instance int) {
	if name := TrigSonarrQueue.WithInstance(instance); !a.cmd.Exec(&common.ActionInput{Type: event}, name) {
		a.cmd.Errorf("[%s requested] Failed! %s Disabled?", event, name)
	}
}

// sonarrApp allows us to have a trigger/timer per instance.
type sonarrApp struct {
	app *apps.SonarrConfig
	cmd *cmd
	idx int
}

// storeQueue runs at an interval and saves the queue for an app internally.
func (app *sonarrApp) storeQueue(ctx context.Context, input *common.ActionInput) {
	queue, err := app.app.GetQueueContext(ctx, queueItemsMax, 1)
	if err != nil {
		app.cmd.Errorf("[%s requested] Getting Sonarr Queue (instance %d): %v", input.Type, app.idx+1, err)
		return
	}

	for _, record := range queue.Records {
		record.Quality = nil
		record.Language = nil
	}

	app.cmd.Debugf("[%s requested] Stored Sonarr Queue (%d items), instance %d %s",
		input.Type, len(queue.Records), app.idx+1, app.app.Name)
	data.SaveWithID("sonarr", app.idx, queue)
}

func (c *cmd) setupSonarr() bool {
	var enable bool

	for idx, app := range c.Apps.Sonarr {
		info := clientinfo.Get()
		if !app.Enabled() || info == nil {
			continue
		}

		var dur time.Duration

		instance := idx + 1
		if info.Actions.Apps.Sonarr.Finished(instance) {
			enable = true
			dur = finishedDuration
		} else if info.Actions.Apps.Sonarr.Stuck(instance) {
			enable = true
			dur = stuckDuration
		}

		if dur != 0 {
			c.Add(&common.Action{
				Hide: true,
				Name: TrigSonarrQueue.WithInstance(instance),
				Fn:   (&sonarrApp{app: app, cmd: c, idx: idx}).storeQueue,
				C:    make(chan *common.ActionInput, 1),
				D:    cnfg.Duration{Duration: dur},
			})
		}
	}

	return enable
}
