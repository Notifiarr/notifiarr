package unmonitor

import (
	"context"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr"
)

type UnmonitorData struct {
	Action    website.EventType `json:"action"`            // unmonitor|delete
	App       string            `json:"app"`               // sonarr|radarr
	Instances []int             `json:"instances"`         // 1,2,3
	TmdbID    int64             `json:"tmdbId,omitempty"`  // tmdb id
	TvdbID    int64             `json:"tvdbId,omitempty"`  // tvdb id
	Season    int               `json:"season,omitempty"`  // season number
	Episode   int               `json:"episode,omitempty"` // episode number
}

type ResponseData struct {
	*UnmonitorData
	Codes    []int    `json:"codes"`    // one per instance
	Statuses []string `json:"statuses"` // one per instance
}

const TrigUnmonitorOrDelete common.TriggerName = "Unmonitoring or deleting played Starr content."

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
	c.Add(&common.Action{
		Key:  "TrigUnmonitorOrDelete",
		Name: TrigUnmonitorOrDelete,
		Fn:   c.unmonitorOrDeleteContent,
		C:    make(chan *common.ActionInput, 1),
	})
}

// Now unmonitors or deletes played Starr content.
func (a *Action) Now(input *common.ActionInput, data *UnmonitorData) {
	input.Raw = data
	a.cmd.Exec(input, TrigUnmonitorOrDelete)
}

// unmonitorOrDeleteContent is Exec'd by the Now() method above though the actions channel.
func (c *cmd) unmonitorOrDeleteContent(ctx context.Context, input *common.ActionInput) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: unmonitorOrDeleteContent", input.Type)
	defer mnd.Log.Trace(reqID, "end: unmonitorOrDeleteContent", input.Type)

	var data *UnmonitorData
	data, ok := input.Raw.(*UnmonitorData)
	if !ok {
		mnd.Log.Errorf(reqID, "Unmonitor data is wrong type (please report this bug): %v", input.Raw)
		return
	}

	event := website.EventEpisode
	if data.App == starr.Radarr.Lower() {
		event = website.EventMovie
	}

	response := c.unmonitorStarrContent(ctx, data)
	// Send a report with the response to the website.
	website.SendData(&website.Request{
		ReqID:   reqID,
		Route:   website.PlexRoute,
		Event:   event,
		Payload: response,
	})
}
