package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/triggers/plexcron"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
)

/* This file sends state of affairs to notifiarr.com */
// That is, it collects library data and downloader data.

const TrigDashboard common.TriggerName = "Initiating State Collection for Dashboard."

const randomMilliseconds = 2500

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
	Day         int64 `json:"day,omitempty"`
}

// States is our compiled states for the dashboard.
type States struct {
	Lidarr   []*State `json:"lidarr"`
	Radarr   []*State `json:"radarr"`
	Readarr  []*State `json:"readarr"`
	Sonarr   []*State `json:"sonarr"`
	NZBGet   []*State `json:"nzbget"`
	RTorrent []*State `json:"rtorrent"`
	Qbit     []*State `json:"qbit"`
	Deluge   []*State `json:"deluge"`
	SabNZB   []*State `json:"sabnzbd"`
	Xmission []*State `json:"transmission"`
	Plex     any      `json:"plexSessions"`
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
	var dur time.Duration

	if ci := clientinfo.Get(); ci != nil && ci.Actions.Dashboard.Interval.Duration > 0 {
		dur = (time.Duration(c.Config.Rand().Intn(randomMilliseconds)) * time.Millisecond) +
			ci.Actions.Dashboard.Interval.Duration

		c.Printf("==> Dashboard State timer started, interval:%s", ci.Actions.Dashboard.Interval)
	}

	c.Add(&common.Action{
		Name: TrigDashboard,
		Fn:   c.sendDashboardState,
		C:    make(chan *common.ActionInput, 1),
		D:    cnfg.Duration{Duration: dur},
	})
}

// Send the current states for the dashboard to the website.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(&common.ActionInput{Type: event}, TrigDashboard)
}

func (c *Cmd) sendDashboardState(ctx context.Context, input *common.ActionInput) {
	var (
		start  = time.Now()
		states = c.getStates(ctx)
		apps   = time.Since(start).Round(time.Millisecond)
	)

	data.Save("dashboard", states)
	c.SendData(&website.Request{
		Route:      website.DashRoute,
		Event:      input.Type,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("Dashboard State (elapsed: %v)", apps),
		Payload:    states,
	})
}

// getStates grabs data for each app.
func (c *Cmd) getStates(ctx context.Context) *States {
	sessions, _ := c.PlexCron.GetSessions(ctx)

	return &States{
		Deluge:   c.getDelugeStates(ctx),
		Lidarr:   c.getLidarrStates(ctx),
		Qbit:     c.getQbitStates(ctx),
		NZBGet:   c.getNZBGetStates(ctx),
		RTorrent: c.getRtorrentStates(),
		Radarr:   c.getRadarrStates(ctx),
		Readarr:  c.getReadarrStates(ctx),
		Sonarr:   c.getSonarrStates(ctx),
		SabNZB:   c.getSabNZBStates(ctx),
		Xmission: c.getTransmissionStates(ctx),
		Plex:     sessions,
	}
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
