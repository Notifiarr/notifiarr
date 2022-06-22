package stuckitems

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/sonarr"
)

/* This file contains the procedures to send stuck download queue items to notifiarr. */

type Config struct {
	*common.Config
}

const TrigStuckItems common.TriggerName = "Checking app queues and sending stuck items."

const (
	// How often to check starr apps for stuck items.
	stuckDur = 5*time.Minute + 1327*time.Millisecond
)

const (
	errorstr  = "error"
	failed    = "failed"
	warning   = "warning"
	completed = "completed"
)

type ListItem struct {
	Elapsed time.Duration `json:"elapsed"`
	Name    string        `json:"name"`
	Queue   []interface{} `json:"queue"`
}

type ItemList map[int]ListItem

type QueuePayload struct {
	Lidarr  ItemList `json:"lidarr,omitempty"`
	Radarr  ItemList `json:"radarr,omitempty"`
	Readarr ItemList `json:"readarr,omitempty"`
	Sonarr  ItemList `json:"sonarr,omitempty"`
}

const getItemsMax = 100

func (c *Config) Create() {
	c.Add(&common.Action{
		Name: TrigStuckItems,
		Fn:   c.sendStuckQueueItems,
		C:    make(chan website.EventType, 1),
		T:    getTicker(c.Apps),
	})
}

// getTicker only returns a ticker if at least 1 app has stuck items turned on.
func getTicker(apps *apps.Apps) *time.Ticker {
	for _, app := range apps.Lidarr {
		if app.StuckItem {
			return time.NewTicker(stuckDur)
		}
	}

	for _, app := range apps.Radarr {
		if app.StuckItem {
			return time.NewTicker(stuckDur)
		}
	}

	for _, app := range apps.Readarr {
		if app.StuckItem {
			return time.NewTicker(stuckDur)
		}
	}

	for _, app := range apps.Sonarr {
		if app.StuckItem {
			return time.NewTicker(stuckDur)
		}
	}

	var ticker *time.Ticker

	return ticker
}

func (i ItemList) Len() (count int) {
	for _, v := range i {
		count += len(v.Queue)
	}

	return count
}

func (i ItemList) Empty() bool {
	return i.Len() < 1
}

func (c *Config) SendStuckQueueItems(event website.EventType) {
	c.Exec(event, TrigStuckItems)
}

func (c *Config) sendStuckQueueItems(event website.EventType) {
	start := time.Now()
	cue := c.getQueues()
	apps := time.Since(start).Round(time.Millisecond)

	if cue == nil || (cue.Lidarr.Empty() && cue.Radarr.Empty() && cue.Readarr.Empty() && cue.Sonarr.Empty()) {
		c.Printf("[%s requested] No stuck items found to send to Notifiarr.", event)
		return
	}

	c.QueueData(&website.SendRequest{
		Route:      website.StuckRoute,
		Event:      event,
		LogPayload: true,
		LogMsg: fmt.Sprintf("Stuck Queue Items (apps:%s) (Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d)",
			apps, cue.Lidarr.Len(), cue.Radarr.Len(), cue.Readarr.Len(), cue.Sonarr.Len()),
		Payload: cue,
	})
}

// getQueues fires a routine for each app type and tries to get a lot of data fast!
func (c *Config) getQueues() *QueuePayload {
	if c.Serial {
		return &QueuePayload{
			Lidarr:  c.getFinishedItemsLidarr(),
			Radarr:  c.getFinishedItemsRadarr(),
			Readarr: c.getFinishedItemsReadarr(),
			Sonarr:  c.getFinishedItemsSonarr(),
		}
	}

	cue := &QueuePayload{}

	var wg sync.WaitGroup

	wg.Add(4) //nolint:gomnd // 4 is 1 for each app polled.

	go func() {
		cue.Lidarr = c.getFinishedItemsLidarr()
		wg.Done() //nolint:wsl
	}()
	go func() {
		cue.Radarr = c.getFinishedItemsRadarr()
		wg.Done() //nolint:wsl
	}()
	go func() {
		cue.Readarr = c.getFinishedItemsReadarr()
		wg.Done() //nolint:wsl
	}()
	go func() {
		cue.Sonarr = c.getFinishedItemsSonarr()
		wg.Done() //nolint:wsl
	}()
	wg.Wait()

	return cue
}

func (c *Config) getFinishedItemsLidarr() ItemList { //nolint:dupl,cyclop
	stuck := make(ItemList)

	for idx, app := range c.Apps.Lidarr {
		instance := idx + 1

		if !app.StuckItem {
			continue
		}

		if app.Lidarr == nil {
			c.Errorf("Getting Lidarr Queue (%d): Lidarr config is nil? This is probably a bug.", instance)
			continue
		}

		start := time.Now()

		queue, err := app.GetQueue(getItemsMax, getItemsMax)
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

func (c *Config) getFinishedItemsRadarr() ItemList { //nolint:cyclop
	stuck := make(ItemList)

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1

		if !app.StuckItem {
			continue
		}

		if app.Radarr == nil {
			c.Errorf("Getting Radarr Queue (%d): Radarr config is nil? This is probably a bug.", instance)
			continue
		}

		start := time.Now()

		queue, err := app.GetQueue(getItemsMax, 1)
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

func (c *Config) getFinishedItemsReadarr() ItemList { //nolint:dupl,cyclop
	stuck := make(ItemList)

	for idx, app := range c.Apps.Readarr {
		instance := idx + 1

		if !app.StuckItem {
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

func (c *Config) getFinishedItemsSonarr() ItemList { //nolint:cyclop
	stuck := make(ItemList)

	for idx, app := range c.Apps.Sonarr {
		instance := idx + 1

		if !app.StuckItem {
			continue
		}

		if app.Sonarr == nil {
			c.Errorf("Getting Sonarr Queue (%d): Sonarr config is nil? This is probably a bug.", instance)
			continue
		}

		start := time.Now()

		queue, err := app.GetQueue(getItemsMax, 1)
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
