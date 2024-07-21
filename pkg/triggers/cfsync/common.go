package cfsync

import (
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
)

/* CF Sync means Custom Format Sync. This is a premium feature that allows syncing
   TRaSH's custom Radarr formats and Sonarr Release Profiles.
	 The code in this file deals with sending data and getting updates at an interval.
*/

const (
	randomMilliseconds = 5000
)

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
	info := clientinfo.Get()
	c.setupRadarr(info)
	c.setupSonarr(info)
	c.setupLidarr(info)

	// Check each instance and enable only if needed.
	if info != nil && info.Actions.Sync.Interval.Duration > 0 {
		if len(info.Actions.Sync.RadarrInstances) > 0 {
			c.Printf("==> Radarr TRaSH Sync: interval: %s, %s ",
				info.Actions.Sync.Interval, strings.Join(info.Actions.Sync.RadarrSync, ", "))
		}

		if len(info.Actions.Sync.SonarrInstances) > 0 {
			c.Printf("==> Sonarr TRaSH Sync: interval: %s, %s ",
				info.Actions.Sync.Interval, strings.Join(info.Actions.Sync.SonarrSync, ", "))
		}

		if len(info.Actions.Sync.LidarrInstances) > 0 {
			c.Printf("==> Lidarr profile and format sync interval: %s, %s",
				info.Actions.Sync.Interval, strings.Join(info.Actions.Sync.SonarrSync, ", "))
		}
	}

	// These aggregate triggers have no timers. Used to sync "all the things" at once.
	c.Add(&common.Action{
		Name: TrigCFSyncRadarr,
		Fn:   c.syncRadarr,
		C:    make(chan *common.ActionInput, 1),
	}, &common.Action{
		Name: TrigRPSyncSonarr,
		Fn:   c.syncSonarr,
		C:    make(chan *common.ActionInput, 1),
	}, &common.Action{
		Name: TrigCFSyncLidarr,
		Fn:   c.syncLidarr,
		C:    make(chan *common.ActionInput, 1),
	})
}

type lidarrApp struct {
	app *apps.LidarrConfig
	cmd *cmd
	idx int
}

func (c *cmd) setupLidarr(info *clientinfo.ClientInfo) {
	if info == nil {
		return
	}

	for idx, app := range c.Apps.Lidarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Sync.LidarrInstances.Has(instance) {
			continue
		}

		var dur cnfg.Duration

		if info != nil && info.Actions.Sync.Interval.Duration > 0 {
			randomTime := time.Duration(c.Config.Rand().Intn(randomMilliseconds)) * time.Millisecond
			dur = cnfg.Duration{Duration: info.Actions.Sync.Interval.Duration + randomTime}
		}

		c.Add(&common.Action{
			Hide: true,
			D:    dur,
			Name: TrigCFSyncLidarrInt.WithInstance(instance),
			Fn:   (&lidarrApp{app: app, cmd: c, idx: idx}).syncLidarr,
			C:    make(chan *common.ActionInput, 1),
		})
	}
}

type radarrApp struct {
	app *apps.RadarrConfig
	cmd *cmd
	idx int
}

func (c *cmd) setupRadarr(info *clientinfo.ClientInfo) {
	if info == nil {
		return
	}

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Sync.RadarrInstances.Has(instance) {
			continue
		}

		var dur cnfg.Duration

		if info != nil && info.Actions.Sync.Interval.Duration > 0 {
			randomTime := time.Duration(c.Config.Rand().Intn(randomMilliseconds)) * time.Millisecond
			dur = cnfg.Duration{Duration: info.Actions.Sync.Interval.Duration + randomTime}
		}

		c.Add(&common.Action{
			Hide: true,
			D:    dur,
			Name: TrigCFSyncRadarrInt.WithInstance(instance),
			Fn:   (&radarrApp{app: app, cmd: c, idx: idx}).syncRadarr,
			C:    make(chan *common.ActionInput, 1),
		})
	}
}

type sonarrApp struct {
	app *apps.SonarrConfig
	cmd *cmd
	idx int
}

func (c *cmd) setupSonarr(info *clientinfo.ClientInfo) {
	if info == nil {
		return
	}

	for idx, app := range c.Apps.Sonarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Sync.SonarrInstances.Has(instance) {
			continue
		}

		var dur cnfg.Duration

		if info != nil && info.Actions.Sync.Interval.Duration > 0 {
			randomTime := time.Duration(c.Config.Rand().Intn(randomMilliseconds)) * time.Millisecond
			dur = cnfg.Duration{Duration: info.Actions.Sync.Interval.Duration + randomTime}
		}

		c.Add(&common.Action{
			Hide: true,
			D:    dur,
			Name: TrigRPSyncSonarrInt.WithInstance(instance),
			Fn:   (&sonarrApp{app: app, cmd: c, idx: idx}).syncSonarr,
			C:    make(chan *common.ActionInput, 1),
		})
	}
}
