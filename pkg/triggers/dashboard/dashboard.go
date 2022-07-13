package dashboard

import (
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/plexcron"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/cnfg"
)

/* This file sends state of affairs to notifiarr.com */
// That is, it collects library data and downloader data.

const TrigDashboard common.TriggerName = "Initiating State Collection for Dashboard."

type Cmd struct {
	*common.Config
	PlexCron *plexcron.Action
}

// Action contains the exported methods for this package.
type Action struct {
	cmd *Cmd
}

// How many "upcoming" or "newest" items to send.
const (
	showNext   = 10
	showLatest = 5
)

// Sortable holds data about any Starr item. Kind of a generic data store.
type Sortable struct {
	id      int64
	Name    string    `json:"name"`
	Sub     string    `json:"subName,omitempty"`
	Date    time.Time `json:"date"`
	Season  int64     `json:"season,omitempty"`
	Episode int64     `json:"episode,omitempty"`
}

// SortableList allows sorting a list.
type SortableList []*Sortable

// State is partially filled out once for each app instance.
type State struct {
	// Shared
	Error    string        `json:"error"`
	Instance int           `json:"instance"`
	Missing  int64         `json:"missing,omitempty"`
	Size     int64         `json:"size"`
	Percent  float64       `json:"percent,omitempty"`
	Upcoming int64         `json:"upcoming,omitempty"`
	Next     SortableList  `json:"next,omitempty"`
	Latest   SortableList  `json:"latest,omitempty"`
	OnDisk   int64         `json:"onDisk,omitempty"`
	Elapsed  cnfg.Duration `json:"elapsed"` // How long it took.
	Name     string        `json:"name"`
	// Radarr
	Movies int64 `json:"movies,omitempty"`
	// Sonarr
	Shows    int64 `json:"shows,omitempty"`
	Episodes int64 `json:"episodes,omitempty"`
	// Readarr
	Authors  int   `json:"authors,omitempty"`
	Books    int64 `json:"books,omitempty"`
	Editions int   `json:"editions,omitempty"`
	// Lidarr
	Artists int   `json:"artists,omitempty"`
	Albums  int64 `json:"albums,omitempty"`
	Tracks  int64 `json:"tracks,omitempty"`
	// Downloader
	Downloads   int   `json:"downloads,omitempty"`
	Uploaded    int64 `json:"uploaded,omitempty"`
	Incomplete  int64 `json:"incomplete,omitempty"`
	Downloaded  int64 `json:"downloaded,omitempty"`
	Uploading   int64 `json:"uploading,omitempty"`
	Downloading int64 `json:"downloading,omitempty"`
	Seeding     int64 `json:"seeding,omitempty"`
	Paused      int64 `json:"paused,omitempty"`
	Errors      int64 `json:"errors,omitempty"`
	Month       int64 `json:"month,omitempty"`
	Week        int64 `json:"week,omitempty"`
}

// States is our compiled states for the dashboard.
type States struct {
	Lidarr   []*State       `json:"lidarr"`
	Radarr   []*State       `json:"radarr"`
	Readarr  []*State       `json:"readarr"`
	Sonarr   []*State       `json:"sonarr"`
	NZBGet   []*State       `json:"nzbget"`
	RTorrent []*State       `json:"rtorrent"`
	Qbit     []*State       `json:"qbit"`
	Deluge   []*State       `json:"deluge"`
	SabNZB   []*State       `json:"sabnzbd"`
	Plex     *plex.Sessions `json:"plexSessions"`
}

// New configures the library.
func New(config *common.Config, plex *plexcron.Action) *Action {
	return &Action{
		cmd: &Cmd{
			Config:   config,
			PlexCron: plex,
		},
	}
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.create()
}

func (c *Cmd) create() {
	var ticker *time.Ticker

	ci := c.ClientInfo

	if ci != nil && ci.Actions.Dashboard.Interval.Duration > 0 {
		ticker = time.NewTicker(ci.Actions.Dashboard.Interval.Duration)
		c.Printf("==> Dashboard State timer started, interval:%s, serial:%v",
			ci.Actions.Dashboard.Interval, c.Config.Serial)
	}

	c.Add(&common.Action{
		Name: TrigDashboard,
		Fn:   c.sendDashboardState,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

// Send the current states for the dashboard to the website.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(event, TrigDashboard)
}

func (c *Cmd) sendDashboardState(event website.EventType) {
	cmd := c.getStatesParallel
	if c.Serial {
		cmd = c.getStatesSerial
	}

	var (
		start  = time.Now()
		states = cmd()
		apps   = time.Since(start).Round(time.Millisecond)
	)

	c.SendData(&website.Request{
		Route:      website.DashRoute,
		Event:      event,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("Dashboard State (elapsed: %v)", apps),
		Payload:    states,
	})
}

// getStatesSerial grabs data for each app serially.
func (c *Cmd) getStatesSerial() *States {
	sessions, _ := c.PlexCron.GetSessions(false)

	return &States{
		Deluge:   c.getDelugeStates(),
		Lidarr:   c.getLidarrStates(),
		Qbit:     c.getQbitStates(),
		NZBGet:   c.getNZBGetStates(),
		RTorrent: c.getRtorrentStates(),
		Radarr:   c.getRadarrStates(),
		Readarr:  c.getReadarrStates(),
		Sonarr:   c.getSonarrStates(),
		SabNZB:   c.getSabNZBStates(),
		Plex:     sessions,
	}
}

// getStatesParallel fires a routine for each app type and tries to get a lot of data fast!
func (c *Cmd) getStatesParallel() *States { //nolint:funlen
	states := &States{}

	var wg sync.WaitGroup

	wg.Add(10) //nolint:gomnd // we are polling 10 apps.

	go func() {
		defer c.CapturePanic()
		states.Deluge = c.getDelugeStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Lidarr = c.getLidarrStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.NZBGet = c.getNZBGetStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.RTorrent = c.getRtorrentStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Qbit = c.getQbitStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Radarr = c.getRadarrStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Readarr = c.getReadarrStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Sonarr = c.getSonarrStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.SabNZB = c.getSabNZBStates()
		wg.Done() //nolint:wsl
	}()
	go func() {
		defer c.CapturePanic()
		states.Plex, _ = c.PlexCron.GetSessions(false)
		wg.Done() //nolint:wsl
	}()
	wg.Wait()

	return states
}

type dateSorter []*Sortable

func (s dateSorter) Len() int {
	return len(s)
}

func (s dateSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s dateSorter) Less(i, j int) bool {
	return s[i].Date.Before(s[j].Date)
}

// Shrink a sortable list.
func (s *SortableList) Shrink(size int) {
	if s == nil {
		return
	}

	if len(*s) > size {
		*s = (*s)[:size]
	}
}
