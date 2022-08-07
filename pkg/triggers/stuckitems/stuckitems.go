package stuckitems

import (
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

/* This file contains the procedures to send stuck download queue items to notifiarr. */

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
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

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.Add(&common.Action{
		Name: TrigStuckItems,
		Fn:   a.cmd.sendStuckQueueItems,
		C:    make(chan website.EventType, 1),
		T:    getTicker(a.cmd.Apps),
	})
}

// getTicker only returns a ticker if at least 1 app has stuck items turned on.
func getTicker(apps *apps.Apps) *time.Ticker { //nolint:gocognit,cyclop
	for _, app := range apps.Lidarr {
		if app.StuckItem && app.URL != "" && app.APIKey != "" && app.Timeout.Duration >= 0 {
			return time.NewTicker(stuckDur)
		}
	}

	for _, app := range apps.Radarr {
		if app.StuckItem && app.URL != "" && app.APIKey != "" && app.Timeout.Duration >= 0 {
			return time.NewTicker(stuckDur)
		}
	}

	for _, app := range apps.Readarr {
		if app.StuckItem && app.URL != "" && app.APIKey != "" && app.Timeout.Duration >= 0 {
			return time.NewTicker(stuckDur)
		}
	}

	for _, app := range apps.Sonarr {
		if app.StuckItem && app.URL != "" && app.APIKey != "" && app.Timeout.Duration >= 0 {
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

// Send stuck items to the website.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(event, TrigStuckItems)
}

func (c *cmd) sendStuckQueueItems(event website.EventType) {
	start := time.Now()
	cue := c.getQueues()
	apps := time.Since(start).Round(time.Millisecond)

	if cue == nil || (cue.Lidarr.Empty() && cue.Radarr.Empty() && cue.Readarr.Empty() && cue.Sonarr.Empty()) {
		c.Printf("[%s requested] No stuck items found to send to Notifiarr.", event)
		return
	}

	c.SendData(&website.Request{
		Route:      website.StuckRoute,
		Event:      event,
		LogPayload: true,
		LogMsg: fmt.Sprintf("Stuck Queue Items (apps:%s) (Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d)",
			apps, cue.Lidarr.Len(), cue.Radarr.Len(), cue.Readarr.Len(), cue.Sonarr.Len()),
		Payload: cue,
	})
}

// getQueues fires a routine for each app type and tries to get a lot of data fast!
func (c *cmd) getQueues() *QueuePayload {
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
