package starrqueue

import (
	"fmt"
	"time"

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
func (a *Action) Create() {
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
		c.Debugf("No stuck items found.")
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
