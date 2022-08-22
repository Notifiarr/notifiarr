package starrqueue

import (
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/sonarr"
)

const TrigSonarrQueue common.TriggerName = "Storing Sonarr instance %d queue."

// StoreSonarr fetches and stores the Sonarr queue immediately for the specified instance.
// Does not send data to the website.
func (a *Action) StoreSonarr(event website.EventType, instance int) {
	if name := TrigSonarrQueue.WithInstance(instance); !a.cmd.Exec(event, name) {
		a.cmd.Errorf("Failed! %s Disbled?", name)
	}
}

// sonarrApp allows us to have a trigger/timer per instance.
type sonarrApp struct {
	app *apps.SonarrConfig
	cmd *cmd
	idx int
}

// storeQueue runs at an interval and saves the queue for an app internally.
func (app *sonarrApp) storeQueue(event website.EventType) {
	queue, err := app.app.GetQueue(queueItemsMax, 1)
	if err != nil {
		app.cmd.Errorf("Getting Sonarr Queue (instance %d): %v", app.idx+1, err)
		return
	}

	for _, record := range queue.Records {
		record.Quality = nil
		record.Language = nil
	}

	data.SaveWithID("sonarr", app.idx, queue)
}

func (c *cmd) setupSonarr() bool {
	var enable bool

	for idx, app := range c.Apps.Sonarr {
		if !app.Enabled() || !c.HaveClientInfo() {
			continue
		}

		var ticker *time.Ticker

		instance := idx + 1
		if c.ClientInfo.Actions.Apps.Sonarr.Finished(instance) {
			enable = true
			ticker = time.NewTicker(finishedDuration)
		} else if c.ClientInfo.Actions.Apps.Sonarr.Stuck(instance) {
			enable = true
			ticker = time.NewTicker(stuckDuration)
		}

		if ticker != nil {
			c.Add(&common.Action{
				Hide: true,
				Name: TrigSonarrQueue.WithInstance(instance),
				Fn:   (&sonarrApp{app: app, cmd: c, idx: idx}).storeQueue,
				C:    make(chan website.EventType, 1),
				T:    ticker,
			})
		}
	}

	return enable
}

func (c *cmd) getFinishedItemsSonarr() itemList { //nolint:cyclop
	stuck := make(itemList)

	for idx, app := range c.Apps.Sonarr {
		if !app.Enabled() || !c.HaveClientInfo() || !c.ClientInfo.Actions.Apps.Sonarr.Stuck(idx+1) {
			continue
		}

		item := data.GetWithID("sonarr", idx)
		if item == nil || item.Data == nil {
			continue
		}

		queue, _ := item.Data.(*sonarr.Queue)
		instance := idx + 1
		stuckapp := stuck[instance]
		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*sonarr.QueueRecord)

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if repeatStomper[item.DownloadID] != nil {
				continue
			}

			repeatStomper[item.DownloadID] = item
			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = c.Apps.Sonarr[idx].Name // this should be safe.
		stuck[instance] = stuckapp

		c.Debugf("Checking Sonarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}
