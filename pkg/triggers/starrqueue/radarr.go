package starrqueue

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const TrigRadarrQueue common.TriggerName = "Checking Radarr queue and sending incomplete items."

// RadarrStuckItems sends Radarr's stuck items to the website.
func (a *Action) RadarrStuckItems(event website.EventType) {
	a.cmd.Exec(event, TrigRadarrQueue)
}

func (c *cmd) setupRadarr() {
	var ticker *time.Ticker

	for _, app := range c.Apps.Radarr {
		if app.StuckItem && app.URL != "" && app.APIKey != "" && app.Timeout.Duration >= 0 {
			ticker = time.NewTicker(queueDuration + time.Duration(rand.Intn(randomSeconds))) //nolint:gosec
			break
		}
	}

	c.Add(&common.Action{
		Name: TrigRadarrQueue,
		Fn:   c.radarrStuckItems,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) radarrStuckItems(event website.EventType) {
	start := time.Now()

	if cue := c.getFinishedItemsRadarr(); !cue.Empty() {
		c.SendData(&website.Request{
			Route:      website.StuckRoute,
			Event:      event,
			LogPayload: true,
			LogMsg: fmt.Sprintf("Radarr Queue: %d stuck items (elapsed:%s)",
				cue.Len(), time.Since(start).Round(time.Millisecond)),
			Payload: map[string]ItemList{"radarr": cue},
		})
	}
}

func (c *cmd) getFinishedItemsRadarr() ItemList { //nolint:cyclop
	stuck := make(ItemList)

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1

		if !app.StuckItem || app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0 {
			continue
		}

		if app.Radarr == nil {
			c.Errorf("Getting Radarr Queue (%d): Radarr config is nil? This is probably a bug.", instance)
			continue
		}

		start := time.Now()

		queue, err := app.GetQueue(queueItemsMax, 1)
		if err != nil {
			c.Errorf("Getting Radarr Queue (%d): %v", instance, err)
			continue
		}

		stuckapp := stuck[instance]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			item.Quality = nil
			item.CustomFormats = nil
			item.Languages = nil
			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = app.Name
		stuckapp.Elapsed = time.Since(start)
		stuck[instance] = stuckapp

		c.Debugf("Checking Radarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}
