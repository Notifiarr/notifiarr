package notifiarr

import (
	"strings"
	"sync"
	"time"

	"golift.io/starr/sonarr"
)

/* This file contains the procedures to send stuck download queue items to notifiarr. */

const (
	errorstr  = "error"
	failed    = "failed"
	warning   = "warning"
	completed = "completed"
)

type custom struct {
	Elapsed time.Duration `json:"elapsed"`
	Repeat  uint          `json:"repeat"`
	Name    string        `json:"name"`
	Queue   []interface{} `json:"queue"`
}

type ItemList map[int]custom

type QueuePayload struct {
	Type    string   `json:"type"`
	Lidarr  ItemList `json:"lidarr,omitempty"`
	Radarr  ItemList `json:"radarr,omitempty"`
	Readarr ItemList `json:"readarr,omitempty"`
	Sonarr  ItemList `json:"sonarr,omitempty"`
}

const getItemsMax = 100

func (i ItemList) Len() (count int) {
	for _, v := range i {
		count += len(v.Queue)
	}

	return count
}

func (i ItemList) Empty() bool {
	return i.Len() < 1
}

func (c *Config) SendFinishedQueueItems(url string) {
	start := time.Now()
	q := c.getQueues()
	apps := time.Since(start).Round(time.Millisecond)

	if q.Lidarr.Empty() && q.Radarr.Empty() && q.Readarr.Empty() && q.Sonarr.Empty() {
		return
	}

	_, _, err := c.SendData(url+ClientRoute, q, true)
	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		c.Errorf("Sending Stuck Queue Items (apps:%s total:%s) (Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d): %v",
			apps, elapsed, q.Lidarr.Len(), q.Radarr.Len(), q.Readarr.Len(), q.Sonarr.Len(), err)
	} else {
		c.Printf("Sent Stuck Items (apps:%s total:%s): Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d",
			apps, elapsed, q.Lidarr.Len(), q.Radarr.Len(), q.Readarr.Len(), q.Sonarr.Len())
	}
}

// getQueues fires a routine for each app type and tries to get a lot of data fast!
func (c *Config) getQueues() *QueuePayload {
	q := &QueuePayload{Type: "queue"}

	var wg sync.WaitGroup

	wg.Add(4) //nolint:gomnd // 4 is 1 for each app polled.

	go func() {
		q.Lidarr = c.getFinishedItemsLidarr()
		wg.Done() //nolint:wsl
	}()
	go func() {
		q.Radarr = c.getFinishedItemsRadarr()
		wg.Done() //nolint:wsl
	}()
	go func() {
		q.Readarr = c.getFinishedItemsReadarr()
		wg.Done() //nolint:wsl
	}()
	go func() {
		q.Sonarr = c.getFinishedItemsSonarr()
		wg.Done() //nolint:wsl
	}()
	wg.Wait()

	return q
}

func (c *Config) getFinishedItemsLidarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Lidarr {
		if l.CheckQ == nil {
			continue
		}

		start := time.Now()

		queue, err := l.GetQueue(getItemsMax)
		if err != nil {
			c.Errorf("Getting Lidarr Queue (%d): %v", i, err)
			continue
		}

		instance := stuck[i+1]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			item.Quality = nil
			instance.Queue = append(instance.Queue, item)
		}

		instance.Name = l.Name
		instance.Repeat = *l.CheckQ
		instance.Elapsed = time.Since(start)
		stuck[i+1] = instance

		c.Debugf("Checking Lidarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1].Queue))
	}

	return stuck
}

func (c *Config) getFinishedItemsRadarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Radarr {
		if l.CheckQ == nil {
			continue
		}

		start := time.Now()

		queue, err := l.GetQueue(getItemsMax, 1)
		if err != nil {
			c.Errorf("Getting Radarr Queue (%d): %v", i, err)
			continue
		}

		instance := stuck[i+1]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			item.Quality = nil
			item.CustomFormats = nil
			item.Languages = nil
			instance.Queue = append(instance.Queue, item)
		}

		instance.Name = l.Name
		instance.Repeat = *l.CheckQ
		instance.Elapsed = time.Since(start)
		stuck[i+1] = instance

		c.Debugf("Checking Radarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1].Queue))
	}

	return stuck
}

func (c *Config) getFinishedItemsReadarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Readarr {
		if l.CheckQ == nil {
			continue
		}

		start := time.Now()

		queue, err := l.GetQueue(getItemsMax)
		if err != nil {
			c.Errorf("Getting Readarr Queue (%d): %v", i, err)
			continue
		}

		instance := stuck[i+1]

		for _, item := range queue.Records {
			if s := strings.ToLower(item.Status); s != completed && s != warning &&
				s != failed && s != errorstr && item.ErrorMessage == "" && len(item.StatusMessages) == 0 {
				continue
			}

			item.Quality = nil
			instance.Queue = append(instance.Queue, item)
		}

		instance.Name = l.Name
		instance.Repeat = *l.CheckQ
		instance.Elapsed = time.Since(start)
		stuck[i+1] = instance

		c.Debugf("Checking Readarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1].Queue))
	}

	return stuck
}

func (c *Config) getFinishedItemsSonarr() ItemList {
	stuck := make(ItemList)

	for i, l := range c.Apps.Sonarr {
		if l.CheckQ == nil {
			continue
		}

		start := time.Now()

		queue, err := l.GetQueue(getItemsMax, 1)
		if err != nil {
			c.Errorf("Getting Sonarr Queue (%d): %v", i, err)
			continue
		}

		// repeatStomper is used to collapse duplicate download IDs.
		repeatStomper := make(map[string]*sonarr.QueueRecord)
		instance := stuck[i+1]

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
			instance.Queue = append(instance.Queue, item)
		}

		instance.Name = l.Name
		instance.Repeat = *l.CheckQ
		instance.Elapsed = time.Since(start)
		stuck[i+1] = instance

		c.Debugf("Checking Sonarr (%d) Queue for Stuck Items, queue size: %d, stuck: %d",
			i+1, len(queue.Records), len(stuck[i+1].Queue))
	}

	return stuck
}
