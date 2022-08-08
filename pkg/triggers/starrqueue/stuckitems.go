package starrqueue

import (
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
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
	// How often to check starr apps for queue list.
	queueDuration = 4*time.Minute + 45*time.Second
	// Between 0 and this number of seconds will be added to each queueDuration.
	randomSeconds = 30
	// This is the max number of queued items to inspect/send.
	queueItemsMax = 100
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

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.setupLidarr()
	a.cmd.setupRadarr()
	a.cmd.setupReadarr()
	a.cmd.setupSonarr()
}

func (i ItemList) Len() int {
	count := 0

	for _, v := range i {
		count += len(v.Queue)
	}

	return count
}

func (i ItemList) Empty() bool {
	return i.Len() < 1
}
