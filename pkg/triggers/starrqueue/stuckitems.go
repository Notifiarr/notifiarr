package starrqueue

import (
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

/* This file contains the procedures to send stuck download queue items to notifiarr. */

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	lidarr  map[int]*lidarr.Queue
	radarr  map[int]*radarr.Queue
	readarr map[int]*readarr.Queue
	sonarr  map[int]*sonarr.Queue
}

const (
	// How often to check starr apps for queue list when stuck items is enabled.
	stuckDuration = 5 * time.Minute
	// How often to check starr apps for queue list when finished items is enabled.
	finishedDuration = time.Minute
	// This is the max number of queued items to inspect/send.
	queueItemsMax = 100
)

const (
	errorstr  = "error"
	failed    = "failed"
	warning   = "warning"
	completed = "completed"
)

const TrigStuckItems common.TriggerName = "Sending cached stuck items to website."

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Run initializes the library.
func (a *Action) Run() {
	a.cmd.lidarr = make(map[int]*lidarr.Queue)
	a.cmd.radarr = make(map[int]*radarr.Queue)
	a.cmd.readarr = make(map[int]*readarr.Queue)
	a.cmd.sonarr = make(map[int]*sonarr.Queue)
	lidarr := a.cmd.setupLidarr()
	radarr := a.cmd.setupRadarr()
	readarr := a.cmd.setupReadarr()
	sonarr := a.cmd.setupSonarr()

	if lidarr || radarr || readarr || sonarr {
		a.cmd.Add(&common.Action{
			Name: TrigStuckItems,
			Fn:   a.cmd.sendStuckQueues,
			C:    make(chan website.EventType, 1),
			T:    time.NewTicker(stuckDuration),
		})
	}
}

// Stop just frees up memory. The are no routines to stop.
func (a *Action) Stop() {
	for idx := range a.cmd.lidarr {
		for i := range a.cmd.lidarr[idx].Records {
			a.cmd.lidarr[idx].Records[i] = nil
		}

		a.cmd.lidarr[idx] = nil
		delete(a.cmd.lidarr, idx)
	}

	for idx := range a.cmd.radarr {
		for i := range a.cmd.radarr[idx].Records {
			a.cmd.radarr[idx].Records[i] = nil
		}

		a.cmd.radarr[idx] = nil
		delete(a.cmd.radarr, idx)
	}

	for idx := range a.cmd.readarr {
		for i := range a.cmd.readarr[idx].Records {
			a.cmd.readarr[idx].Records[i] = nil
		}

		a.cmd.readarr[idx] = nil
		delete(a.cmd.readarr, idx)
	}

	for idx := range a.cmd.sonarr {
		for i := range a.cmd.sonarr[idx].Records {
			a.cmd.sonarr[idx].Records[i] = nil
		}

		a.cmd.sonarr[idx] = nil
		delete(a.cmd.sonarr, idx)
	}

	a.cmd.lidarr = nil
	a.cmd.radarr = nil
	a.cmd.readarr = nil
	a.cmd.sonarr = nil
}

// StuckItems sends the stuck queues items for all apps.
// Does not fetch fresh data first, uses cache.
func (a *Action) StuckItems(event website.EventType) {
	a.cmd.Exec(event, TrigStuckItems)
}

// listItem is data formatted for sending a json payload to the website.
type listItem struct {
	Name  string        `json:"name"`
	Queue []interface{} `json:"queue"`
}

// itemList stores an instance->queue map.
type itemList map[int]listItem

func (i itemList) Len() int {
	count := 0

	for _, v := range i {
		count += len(v.Queue)
	}

	return count
}

func (i itemList) Empty() bool {
	return i.Len() < 1
}

// sendStuckQueues gathers the stuck quues from cache and sends them.
func (c *cmd) sendStuckQueues(event website.EventType) {
	lidarr := c.getFinishedItemsLidarr()
	radarr := c.getFinishedItemsRadarr()
	readarr := c.getFinishedItemsReadarr()
	sonarr := c.getFinishedItemsSonarr()

	if lidarr.Empty() && radarr.Empty() && readarr.Empty() && sonarr.Empty() {
		return
	}

	type stuckPaylod struct {
		Lidarr  itemList `json:"lidarr"`
		Radarr  itemList `json:"radarr"`
		Readarr itemList `json:"readarr"`
		Sonarr  itemList `json:"sonarr"`
	}

	c.SendData(&website.Request{
		Route:      website.StuckRoute,
		Event:      event,
		LogPayload: true,
		LogMsg: fmt.Sprintf("Stuck Items; Lidarr: %d, Radarr: %d, Readarr: %d, Sonarr: %d",
			lidarr.Len(), radarr.Len(), readarr.Len(), sonarr.Len()),
		Payload: stuckPaylod{
			Lidarr:  lidarr,
			Radarr:  radarr,
			Readarr: readarr,
			Sonarr:  sonarr,
		},
	})
}
