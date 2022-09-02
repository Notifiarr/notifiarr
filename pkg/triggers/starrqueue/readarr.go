package starrqueue

import (
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/readarr"
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
func (app *readarrApp) storeQueue(event website.EventType) {
	queue, err := app.app.GetQueue(queueItemsMax, 1)
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
		if !app.Enabled() || !c.HaveClientInfo() {
			continue
		}

		var ticker *time.Ticker

		instance := idx + 1

		switch {
		case c.ClientInfo.Actions.Apps.Readarr.Finished(instance):
			enable = true
			ticker = time.NewTicker(finishedDuration)
		case c.ClientInfo.Actions.Apps.Readarr.Stuck(instance):
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

func (c *cmd) getFinishedItemsReadarr() itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Readarr {
		if !app.Enabled() || !c.HaveClientInfo() || !c.ClientInfo.Actions.Apps.Readarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("readarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*readarr.Queue)
		instance := idx + 1
		stuckapp := stuck[instance]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = c.Apps.Readarr[idx].Name // this should be safe.
		stuck[instance] = stuckapp

		c.Debugf("Checking Readarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}
