package unmonitor

import (
	"context"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr"
)

type UnmonitorData struct {
	Action    website.EventType `json:"action"`    // unmonitor|delete
	App       string            `json:"app"`       // sonarr|radarr
	Instances []int             `json:"instances"` // 1,2,3
	TvdbID    int64             `json:"tvdbid"`    // tvdb id
	TmdbID    int64             `json:"tmdbid"`    // tmdb id
	Season    int               `json:"season"`    // season number
	Episode   int               `json:"episode"`   // episode number
}

type ResponseData struct {
	*UnmonitorData
	Codes    []int    `json:"codes"`    // one per instance
	Statuses []string `json:"statuses"` // one per instance
}

const TrigPlexUnmonitor common.TriggerName = "Unmonitoring or deleting Starr item played by Plex."

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
		Key:  "TrigPlexUnmonitor",
		Name: TrigPlexUnmonitor,
		Fn:   c.unmonitorPlexPlayedItems,
		C:    make(chan *common.ActionInput, 1),
	})
}

// Now unmonitors Plex-Played Items in Sonarr or Radarr.
func (a *Action) Now(input *common.ActionInput, data *UnmonitorData) {
	input.Raw = data
	a.cmd.Exec(input, TrigPlexUnmonitor)
}

// unmonitorPlexPlayedItems is Exec'd by the Now() method above though the actions channel.
func (c *cmd) unmonitorPlexPlayedItems(ctx context.Context, input *common.ActionInput) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: unmonitorPlexPlayedItems", input.Type)
	defer mnd.Log.Trace(reqID, "end: unmonitorPlexPlayedItems", input.Type)

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
