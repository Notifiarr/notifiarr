package cfsync

import (
	"math/rand"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

/* CF Sync means Custom Format Sync. This is a premium feature that allows syncing
   TRaSH's custom Radarr formats and Sonarr Release Profiles.
	 The code in this file deals with sending data and getting updates at an interval.
*/

const randomMilliseconds = 5000

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{
		cmd: &cmd{
			Config: config,
		},
	}
}

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.create()
}

func (c *cmd) create() {
	ci := c.ClientInfo

	var (
		radarrTicker *time.Ticker
		sonarrTicker *time.Ticker
	)

	// Check each instance and enable only if needed.
	//nolint:gosec
	if ci != nil && ci.Actions.Sync.Interval.Duration > 0 {
		if len(ci.Actions.Sync.RadarrInstances) > 0 {
			randomTime := time.Duration(rand.Intn(randomMilliseconds)) * time.Millisecond
			radarrTicker = time.NewTicker(ci.Actions.Sync.Interval.Duration + randomTime)
			c.Printf("==> Keeping %d Radarr Custom Formats synced, interval:%s",
				ci.Actions.Sync.Radarr, ci.Actions.Sync.Interval)
		}

		if len(ci.Actions.Sync.SonarrInstances) > 0 {
			randomTime := time.Duration(rand.Intn(randomMilliseconds)) * time.Millisecond
			sonarrTicker = time.NewTicker(ci.Actions.Sync.Interval.Duration + randomTime)
			c.Printf("==> Keeping %d Sonarr Release Profiles synced, interval:%s",
				ci.Actions.Sync.Sonarr, ci.Actions.Sync.Interval)
		}
	}

	c.Add(&common.Action{
		Name: TrigCFSyncRadarr,
		Fn:   c.syncRadarr,
		C:    make(chan website.EventType, 1),
		T:    radarrTicker,
	}, &common.Action{
		Name: TrigRPSyncSonarr,
		Fn:   c.syncSonarr,
		C:    make(chan website.EventType, 1),
		T:    sonarrTicker,
	})
}
