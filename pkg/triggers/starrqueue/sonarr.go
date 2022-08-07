package starrqueue

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/sonarr"
)

const TrigSonarrQueue common.TriggerName = "Checking Sonarr queue and sending incomplete items."

// SonarrStuckItems sends Sonarr's stuck items to the website.
func (a *Action) SonarrStuckItems(event website.EventType) {
	a.cmd.Exec(event, TrigSonarrQueue)
}

func (c *cmd) setupSonarr() {
	var ticker *time.Ticker

	for _, app := range c.Apps.Sonarr {
		if app.StuckItem && app.URL != "" && app.APIKey != "" && app.Timeout.Duration >= 0 {
			ticker = time.NewTicker(queueDuration + time.Duration(rand.Intn(randomSeconds))) //nolint:gosec
			break
		}
	}

	c.Add(&common.Action{
		Name: TrigSonarrQueue,
		Fn:   c.sonarrStuckItems,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) sonarrStuckItems(event website.EventType) {
	start := time.Now()

	if cue := c.getFinishedItemsSonarr(); !cue.Empty() {
		c.SendData(&website.Request{
			Route:      website.StuckRoute,
			Event:      event,
			LogPayload: true,
			LogMsg: fmt.Sprintf("Stuck Sonarr Queue (elapsed:%s): Sonarr: %d",
				time.Since(start).Round(time.Millisecond), cue.Len()),
			Payload: map[string]ItemList{"sonarr": cue},
		})
	}
}

func (c *cmd) getFinishedItemsSonarr() ItemList { //nolint:cyclop
	stuck := make(ItemList)

	for idx, app := range c.Apps.Sonarr {
		instance := idx + 1

		if !app.StuckItem || app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0 {
			continue
		}

		if app.Sonarr == nil {
			c.Errorf("Getting Sonarr Queue (%d): Sonarr config is nil? This is probably a bug.", instance)
			continue
		}

		start := time.Now()

		queue, err := app.GetQueue(queueItemsMax, 1)
		if err != nil {
			c.Errorf("Getting Sonarr Queue (%d): %v", instance, err)
			continue
		}

		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*sonarr.QueueRecord)
		stuckapp := stuck[instance]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			} else if repeatStomper[item.DownloadID] != nil {
				continue
			}

			item.Quality = nil
			item.Language = nil
			repeatStomper[item.DownloadID] = item
			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = app.Name
		stuckapp.Elapsed = time.Since(start)
		stuck[instance] = stuckapp

		c.Debugf("Checking Sonarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}
