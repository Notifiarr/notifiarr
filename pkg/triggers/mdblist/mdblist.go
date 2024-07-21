package mdblist

import (
	"context"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
	"golift.io/starr/radarr"
)

const TrigMDBListSync common.TriggerName = "Sending Library contents for MDBList."

const randomMilliseconds = 5000

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
}

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.create()
}

func (c *cmd) create() {
	var dur time.Duration

	info := clientinfo.Get()
	if info != nil && info.Actions.Mdblist.Interval.Duration > 0 &&
		(len(info.Actions.Mdblist.Radarr) > 0 || len(info.Actions.Mdblist.Sonarr) > 0) {
		randomTime := time.Duration(c.Config.Rand().Intn(randomMilliseconds)) * time.Millisecond
		dur = info.Actions.Mdblist.Interval.Duration + randomTime
		c.Printf("==> MDB List Timer Enabled, interval:%s, Radarr/Sonarr: %d/%d instances",
			info.Actions.Mdblist.Interval, len(info.Actions.Mdblist.Radarr), len(info.Actions.Mdblist.Sonarr))
	}

	c.Add(&common.Action{
		Name: TrigMDBListSync,
		Fn:   c.sendMDBList,
		C:    make(chan *common.ActionInput, 1),
		D:    cnfg.Duration{Duration: dur},
	})
}

// Send library contents to the website for MDBList.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(&common.ActionInput{Type: event}, TrigMDBListSync)
}

type mdbListPayload struct {
	Instance int            `json:"instance"`
	Name     string         `json:"name"`
	Library  []*libraryData `json:"library"`
	Error    string         `json:"error"`
}

type libraryData struct {
	Imdb   string `json:"imdb,omitempty"`
	Tmdb   int64  `json:"tmdb,omitempty"`
	Tvdb   int64  `json:"tvdb,omitempty"`
	Exists bool   `json:"exists"`
}

func (c *cmd) sendMDBList(ctx context.Context, input *common.ActionInput) {
	c.SendData(&website.Request{
		Route:      website.MdbListRoute,
		Event:      input.Type,
		LogPayload: true,
		LogMsg:     "MDBList Libraries Update",
		Payload: map[string][]*mdbListPayload{
			"radarr": c.getRadarrLibraries(ctx, input),
			"sonarr": c.getSonarrLibraries(ctx, input),
		},
	})
}

func (c *cmd) getRadarrLibraries(ctx context.Context, input *common.ActionInput) []*mdbListPayload {
	output := []*mdbListPayload{}
	ci := clientinfo.Get()

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if !app.Enabled() || !ci.Actions.Mdblist.Radarr.Has(instance) {
			c.Debugf("Skipping MDBList for Radarr %d:%s, not enabled.", instance, app.URL)
			continue
		}

		library := &mdbListPayload{Instance: instance, Name: app.Name}
		output = append(output, library)

		items, err := app.GetMovieContext(ctx, &radarr.GetMovie{ExcludeLocalCovers: true})
		if err != nil {
			library.Error = err.Error()
			c.Errorf("[%s requested] Radarr Library (MDBList) (%d:%s) failed: getting movies: %v",
				input.Type, instance, app.URL, library.Error)

			continue
		}

		library.Library = make([]*libraryData, len(items))
		for idx, item := range items {
			library.Library[idx] = &libraryData{
				Imdb:   item.ImdbID,
				Tmdb:   item.TmdbID,
				Exists: item.SizeOnDisk > 0,
			}
		}
	}

	return output
}

func (c *cmd) getSonarrLibraries(ctx context.Context, input *common.ActionInput) []*mdbListPayload {
	output := []*mdbListPayload{}
	ci := clientinfo.Get()

	for idx, app := range c.Apps.Sonarr {
		instance := idx + 1
		if !app.Enabled() || !ci.Actions.Mdblist.Sonarr.Has(instance) {
			c.Debugf("Skipping MDBList for Sonarr %d:%s, not enabled.", instance, app.URL)
			continue
		}

		library := &mdbListPayload{Instance: instance, Name: app.Name}
		output = append(output, library)

		items, err := app.GetSeriesContext(ctx, 0)
		if err != nil {
			library.Error = err.Error()
			c.Errorf("[%s requested] Sonarr Library (MDBList) (%d:%s) failed: getting series: %v",
				input.Type, instance, app.URL, library.Error)

			continue
		}

		library.Library = make([]*libraryData, len(items))
		for idx, item := range items {
			library.Library[idx] = &libraryData{
				Imdb:   item.ImdbID,
				Tvdb:   item.TvdbID,
				Exists: item.Statistics != nil && item.Statistics.SizeOnDisk > 0,
			}
		}
	}

	return output
}
