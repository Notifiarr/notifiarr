package unmonitor

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
)

type UnmonitorData struct {
	Action    website.EventType `json:"action"`    // unmonitor|delete
	App       string            `json:"app"`       // sonarr|radarr
	Instances []int             `json:"instances"` // 1,2,3
	TvDBID    int64             `json:"tvdbid"`    // tvdb id
	TMDbID    int64             `json:"tmdbid"`    // tmdb id
	Season    int               `json:"season"`    // season number
	Episode   int64             `json:"episode"`   // episode number
}

type ResponseData struct {
	*UnmonitorData
	Codes    []int    `json:"codes"`    // one per instance
	Statuses []string `json:"statuses"` // one per instance
}

const TrigPlexUnmonitor common.TriggerName = "Unmonitoring or deleting Plex-played item."

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

// Unmonitor unmonitors Plex-Played Items in Sonarr or Radarr.
func (a *Action) Unmonitor(input *common.ActionInput, data *UnmonitorData) {
	input.Raw = data
	a.cmd.Exec(input, TrigPlexUnmonitor)
}

func (c *cmd) unmonitorPlexPlayedItems(ctx context.Context, input *common.ActionInput) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: unmonitorPlexPlayedItems", input.Type)
	defer mnd.Log.Trace(reqID, "end: unmonitorPlexPlayedItems", input.Type)

	var data *UnmonitorData
	data, ok := input.Raw.(*UnmonitorData)
	if !ok {
		mnd.Log.Errorf("{trace:%s} Unmonitor data is wrong type (this is a bug): %v", reqID, input.Raw)
		return
	}

	event := website.EventEpisode
	if data.App == starr.Radarr.Lower() {
		event = website.EventMovie
	}

	response := c.unmonitorStarrContent(ctx, data)
	// Send a report with the response to the website.
	website.SendData(&website.Request{
		ReqID:   mnd.GetID(ctx),
		Route:   website.PlexRoute,
		Event:   event,
		Payload: response,
	})
}

func (c *cmd) unmonitorStarrContent(ctx context.Context, data *UnmonitorData) *ResponseData { //nolint:cyclop
	response := &ResponseData{
		UnmonitorData: data,
		Codes:         make([]int, len(data.Instances)),
		Statuses:      make([]string, len(data.Instances)),
	}

	for idx := range response.Instances {
		if data.App == starr.Sonarr.Lower() && response.Instances[idx] >= len(c.Apps.Sonarr) ||
			data.App == starr.Radarr.Lower() && response.Instances[idx] >= len(c.Apps.Radarr) {
			response.Codes[idx] = http.StatusNotFound
			response.Statuses[idx] = data.App + " instance not found: " + strconv.Itoa(response.Instances[idx])
			continue
		}

		if data.App == starr.Sonarr.Lower() && !c.Apps.Sonarr[response.Instances[idx]].Enabled() ||
			data.App == starr.Radarr.Lower() && !c.Apps.Radarr[response.Instances[idx]].Enabled() {
			response.Codes[idx] = http.StatusNotFound
			response.Statuses[idx] = data.App + " instance not enabled: " + strconv.Itoa(response.Instances[idx])
			continue
		}

		// Try to unmonitor or delete the episode in each instance.
		switch data.App {
		case starr.Sonarr.Lower():
			response.Codes[idx], response.Statuses[idx] = c.unmonitorSonarrEpisode(ctx, response, idx)
		case starr.Radarr.Lower():
			response.Codes[idx], response.Statuses[idx] = c.unmonitorRadarrMovie(ctx, response, idx)
		}
	}

	return response
}

// unmonitorSonarrEpisode takes in a TVDB ID, Season, and Episode number
// and unmonitors or deletes the episode from the given Sonarr instance.
func (c *cmd) unmonitorSonarrEpisode(ctx context.Context, response *ResponseData, idx int) (int, string) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: unmonitorSonarrEpisode", response.Instances[idx])
	defer mnd.Log.Trace(reqID, "end: unmonitorSonarrEpisode", response.Instances[idx], response.TvDBID)

	sonarrInstance := c.Apps.Sonarr[response.Instances[idx]]

	// Get the internal SonarrSeries ID from Sonarr using the TVDB ID.
	series, err := sonarrInstance.GetSeriesContext(ctx, response.TvDBID)
	if err != nil {
		return parseStarrError(err)
	}

	if len(series) == 0 || response.TvDBID != series[0].TvdbID || response.TvDBID == 0 {
		return http.StatusNotFound, "Series not found"
	}

	mnd.Log.Trace(reqID, "getting episode: unmonitorSonarrEpisode", series[0].ID, response.Season, response.Episode)

	// Get the Episode ID from the Series ID and Season and Episode numbers.
	episode, err := sonarrInstance.GetSeriesEpisodesContext(ctx, &sonarr.GetEpisode{
		SeriesID:     series[0].ID,
		SeasonNumber: response.Season,
		EpisodeIDs:   []int64{response.Episode},
	})
	if err != nil {
		return parseStarrError(err)
	}

	if len(episode) == 0 {
		return http.StatusNotFound, "Episode not found"
	}

	// Check if the instance is rate limited.
	if !sonarrInstance.DelOK() {
		return http.StatusLocked, "Rate limit reached"
	}

	mnd.Log.Trace(reqID, response.Action, "episode: unmonitorSonarrEpisode", episode[0].ID)

	if response.Action == "delete" {
		// Delete the Episode File if the action is delete.
		err = sonarrInstance.DeleteEpisodeFileContext(ctx, episode[0].ID)
	} else {
		// Unmonitor the Episode if the action is unmonitor.
		_, err = sonarrInstance.MonitorEpisodeContext(ctx, []int64{episode[0].ID}, false)
	}

	if err != nil {
		return parseStarrError(err)
	}

	return http.StatusOK, "OK"
}

func (c *cmd) unmonitorRadarrMovie(ctx context.Context, response *ResponseData, idx int) (int, string) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: unmonitorRadarrMovie", response.Instances[idx], response.TMDbID)
	defer mnd.Log.Trace(reqID, "end: unmonitorRadarrMovie", response.Instances[idx], response.TMDbID)

	radarrInstance := c.Apps.Radarr[response.Instances[idx]]

	// Get the internal RadarrMovie ID from Radarr using the TMDb ID.
	movie, err := radarrInstance.GetMovieContext(ctx, &radarr.GetMovie{TMDBID: response.TMDbID})
	if err != nil {
		return parseStarrError(err)
	}

	if len(movie) == 0 || response.TMDbID != movie[0].TmdbID || response.TMDbID == 0 {
		return http.StatusNotFound, "Movie not found"
	}

	// Check if the instance is rate limited.
	if !radarrInstance.DelOK() {
		return http.StatusLocked, "Rate limit reached"
	}

	mnd.Log.Trace(reqID, response.Action, "movie: unmonitorRadarrMovie", movie[0].ID)

	if response.Action == "delete" {
		// Delete the Movie File if the action is delete.
		err = radarrInstance.DeleteMovieFilesContext(ctx, movie[0].ID)
	} else {
		movie[0].Monitored = false
		_, err = radarrInstance.UpdateMovieContext(ctx, movie[0].ID, movie[0], false)
	}

	if err != nil {
		return parseStarrError(err)
	}

	return http.StatusOK, "OK"
}

func parseStarrError(err error) (int, string) {
	var starrErr *starr.ReqError
	if errors.As(err, &starrErr) {
		return starrErr.Code, string(starrErr.Body)
	}

	return http.StatusInternalServerError, err.Error()
}
