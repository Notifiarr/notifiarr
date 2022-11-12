package cfsync

import (
	"math/rand"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
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
	ci := clientinfo.Get()
	c.setupRadarr(ci)
	c.setupSonarr(ci)

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
			c.Printf("==> Radarr TRaSH Sync: interval: %s, %s ",
				ci.Actions.Sync.Interval, strings.Join(ci.Actions.Sync.RadarrSync, ", "))
		}

		if len(ci.Actions.Sync.SonarrInstances) > 0 {
			randomTime := time.Duration(rand.Intn(randomMilliseconds)) * time.Millisecond
			sonarrTicker = time.NewTicker(ci.Actions.Sync.Interval.Duration + randomTime)
			c.Printf("==> Sonarr TRaSH Sync: interval: %s, %s ",
				ci.Actions.Sync.Interval, strings.Join(ci.Actions.Sync.SonarrSync, ", "))
		}
	}

	c.Add(&common.Action{
		Name: TrigCFSyncRadarr,
		Fn:   c.syncRadarr,
		C:    make(chan *common.ActionInput, 1),
		T:    radarrTicker,
	}, &common.Action{
		Name: TrigRPSyncSonarr,
		Fn:   c.syncSonarr,
		C:    make(chan *common.ActionInput, 1),
		T:    sonarrTicker,
	})
}

type radarrApp struct {
	app *apps.RadarrConfig
	cmd *cmd
	idx int
}

func (c *cmd) setupRadarr(ci *clientinfo.ClientInfo) {
	if ci == nil {
		return
	}

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if !app.Enabled() || !ci.Actions.Sync.RadarrInstances.Has(instance) {
			continue
		}

		var ticker *time.Ticker
		//nolint:gosec
		if ci != nil && ci.Actions.Sync.Interval.Duration > 0 {
			randomTime := time.Duration(rand.Intn(randomMilliseconds)) * time.Millisecond
			ticker = time.NewTicker(ci.Actions.Sync.Interval.Duration + randomTime)
		}

		c.Add(&common.Action{
			Hide: true,
			Name: TrigCFSyncRadarrInt.WithInstance(instance),
			Fn:   (&radarrApp{app: app, cmd: c, idx: idx}).syncRadarr,
			C:    make(chan *common.ActionInput, 1),
			T:    ticker,
		})
	}
}

type sonarrApp struct {
	app *apps.SonarrConfig
	cmd *cmd
	idx int
}

func (c *cmd) setupSonarr(ci *clientinfo.ClientInfo) {
	if ci == nil {
		return
	}

	for idx, app := range c.Apps.Sonarr {
		instance := idx + 1
		if !app.Enabled() || !ci.Actions.Sync.SonarrInstances.Has(instance) {
			continue
		}

		var ticker *time.Ticker
		//nolint:gosec
		if ci != nil && ci.Actions.Sync.Interval.Duration > 0 {
			randomTime := time.Duration(rand.Intn(randomMilliseconds)) * time.Millisecond
			ticker = time.NewTicker(ci.Actions.Sync.Interval.Duration + randomTime)
		}

		c.Add(&common.Action{
			Hide: true,
			Name: TrigRPSyncSonarrInt.WithInstance(instance),
			Fn:   (&sonarrApp{app: app, cmd: c, idx: idx}).syncSonarr,
			C:    make(chan *common.ActionInput, 1),
			T:    ticker,
		})
	}
}
