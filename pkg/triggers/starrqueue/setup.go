package starrqueue

import (
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
)

/* This file contains the procedures to send stuck download queue items to notifiarr. */

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	// We set empty to true after we send 1 "empty downloads" payload.
	empty bool
}

const (
	// How often to check starr apps for queue list when stuck items is enabled.
	stuckDuration = 5 * time.Minute
	// How often to check starr apps for queue list when finished items is enabled.
	finishedDuration = time.Minute
	// This is the max number of queued items to inspect/send.
	queueItemsMax = 1000
)

const (
	errorstr    = "error"
	failed      = "failed"
	warning     = "warning"
	completed   = "completed"
	downloading = "downloading"
	delay       = "delay"
)

const (
	TrigStuckItems       common.TriggerName = "Sending cached stuck items to website."
	TrigDownloadingItems common.TriggerName = "Sending cached downloading items to website."
)

// QueuesPaylod is what we send to the website.
type QueuesPaylod struct {
	Lidarr  itemList `json:"lidarr"`
	Radarr  itemList `json:"radarr"`
	Readarr itemList `json:"readarr"`
	Sonarr  itemList `json:"sonarr"`
}

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
			C:    make(chan *common.ActionInput, 1),
			D:    cnfg.Duration{Duration: stuckDuration},
		})

		// Only enable this timer if the user is a patron.
		if ci := clientinfo.Get(); ci != nil && ci.IsPatron() {
			a.cmd.Add(&common.Action{
				Hide: true,
				Name: TrigDownloadingItems,
				Fn:   a.cmd.sendDownloadingQueues,
				C:    make(chan *common.ActionInput, 1),
				D:    cnfg.Duration{Duration: finishedDuration},
			})
		}
	}
}

// listItem is data formatted for sending a json payload to the website.
type listItem struct {
	Name  string        `json:"name"`
	Queue []interface{} `json:"queue"`
	Total int           `json:"total"`
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
