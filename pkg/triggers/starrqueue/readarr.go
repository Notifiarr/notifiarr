package starrqueue

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const TrigReadarrQueue common.TriggerName = "Checking Readarr queue and sending incomplete items."

// ReadarrStuckItems sends Readarr's stuck items to the website.
func (a *Action) ReadarrStuckItems(event website.EventType) {
	a.cmd.Exec(event, TrigReadarrQueue)
}

func (c *cmd) setupReadarr() {
	var ticker *time.Ticker

	for _, app := range c.Apps.Readarr {
		if app.StuckItem && app.URL != "" && app.APIKey != "" && app.Timeout.Duration >= 0 {
			ticker = time.NewTicker(queueDuration + time.Duration(rand.Intn(randomSeconds))) //nolint:gosec
			break
		}
	}

	c.Add(&common.Action{
		Name: TrigReadarrQueue,
		Fn:   c.readarrStuckItems,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) readarrStuckItems(event website.EventType) {
	start := time.Now()

	if cue := c.getFinishedItemsReadarr(); !cue.Empty() {
		c.SendData(&website.Request{
			Route:      website.StuckRoute,
			Event:      event,
			LogPayload: true,
			LogMsg: fmt.Sprintf("Readarr Queue: %d stuck items (elapsed:%s)",
				cue.Len(), time.Since(start).Round(time.Millisecond)),
			Payload: map[string]ItemList{"readarr": cue},
		})
	}
}

func (c *cmd) getFinishedItemsReadarr() ItemList { //nolint:dupl,cyclop
	stuck := make(ItemList)

	for idx, app := range c.Apps.Readarr {
		instance := idx + 1

		if !app.StuckItem || app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0 {
			continue
		}

		if app.Readarr == nil {
			c.Errorf("Getting Readarr Queue (%d): Readarr config is nil? This is probably a bug.", instance)
			continue
		}

		start := time.Now()

		queue, err := app.GetQueue(queueItemsMax, queueItemsMax)
		if err != nil {
			c.Errorf("Getting Readarr Queue (%d): %v", instance, err)
			continue
		}

		stuckapp := stuck[instance]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			item.Quality = nil
			stuckapp.Queue = append(stuckapp.Queue, item)
		}

		stuckapp.Name = app.Name
		stuckapp.Elapsed = time.Since(start)
		stuck[instance] = stuckapp

		c.Debugf("Checking Readarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}
