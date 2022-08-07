package starrqueue

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const TrigLidarrQueue common.TriggerName = "Checking Lidarr queue and sending incomplete items."

// LidarrStuckItems sends Lidarr's stuck items to the website.
func (a *Action) LidarrStuckItems(event website.EventType) {
	a.cmd.Exec(event, TrigLidarrQueue)
}

func (c *cmd) setupLidarr() {
	var ticker *time.Ticker

	for _, app := range c.Apps.Lidarr {
		if app.StuckItem && app.URL != "" && app.APIKey != "" && app.Timeout.Duration >= 0 {
			ticker = time.NewTicker(queueDuration + time.Duration(rand.Intn(randomSeconds))) //nolint:gosec
			break
		}
	}

	c.Add(&common.Action{
		Name: TrigLidarrQueue,
		Fn:   c.lidarrStuckItems,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) lidarrStuckItems(event website.EventType) {
	start := time.Now()

	if cue := c.getFinishedItemsLidarr(); cue.Empty() {
		c.SendData(&website.Request{
			Route:      website.StuckRoute,
			Event:      event,
			LogPayload: true,
			LogMsg: fmt.Sprintf("Stuck Lidarr Queue (elapsed:%s): Lidarr: %d",
				time.Since(start).Round(time.Millisecond), cue.Len()),
			Payload: map[string]ItemList{"lidarr": cue},
		})
	}
}

func (c *cmd) getFinishedItemsLidarr() ItemList { //nolint:dupl,cyclop
	stuck := make(ItemList)

	for idx, app := range c.Apps.Lidarr {
		instance := idx + 1

		if !app.StuckItem || app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0 {
			continue
		}

		if app.Lidarr == nil {
			c.Errorf("Getting Lidarr Queue (%d): Lidarr config is nil? This is probably a bug.", instance)
			continue
		}

		start := time.Now()

		queue, err := app.GetQueue(queueItemsMax, queueItemsMax)
		if err != nil {
			c.Errorf("Getting Lidarr Queue (%d): %v", instance, err)
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

		c.Debugf("Checking Lidarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			instance, len(queue.Records), len(stuck[instance].Queue))
	}

	return stuck
}
