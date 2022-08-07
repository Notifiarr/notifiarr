package stuckitems

import (
	"strings"
	"time"
)

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

		queue, err := app.GetQueue(getItemsMax, getItemsMax)
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
