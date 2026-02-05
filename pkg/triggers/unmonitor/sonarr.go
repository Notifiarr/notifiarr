package unmonitor

import (
	"context"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr/sonarr"
)

// unmonitorSonarrEpisode takes in a TVDB ID, Season, and Episode number
// and unmonitors or deletes the episode from the given Sonarr instance.
func (c *cmd) unmonitorSonarrEpisode( //nolint:cyclop
	ctx context.Context, response *ResponseData, idx int,
) (int, string) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: unmonitorSonarrEpisode", response.Instances[idx])
	defer mnd.Log.Trace(reqID, "end: unmonitorSonarrEpisode", response.Instances[idx], response.TvdbID)

	sonarrInstance := c.Apps.Sonarr[response.Instances[idx]]

	// Get the internal SonarrSeries ID from Sonarr using the TVDB ID.
	series, err := sonarrInstance.GetSeriesContext(ctx, response.TvdbID)
	if err != nil {
		return parseStarrError(err)
	}

	// Make sure we got the correct series.
	if len(series) == 0 || response.TvdbID != series[0].TvdbID || response.TvdbID == 0 {
		return http.StatusNotFound, "Series not found in this Sonarr instance."
	}

	mnd.Log.Trace(reqID, "getting episode: unmonitorSonarrEpisode", series[0].ID, response.Season, response.Episode)

	// Get all the episodes for the series and season.
	episodes, err := sonarrInstance.GetSeriesEpisodesContext(ctx, &sonarr.GetEpisode{
		SeriesID:     series[0].ID,
		SeasonNumber: response.Season,
	})
	if err != nil {
		return parseStarrError(err)
	}

	var (
		episodeID     int64 = 0 // for unmonitoring
		episodeFileID int64 = 0 // for deleting
	)
	// Loop through the season's episodes and find the correct episode by number.
	for _, episode := range episodes {
		if episode.EpisodeNumber == response.Episode {
			episodeID = episode.ID                // for unmonitoring
			episodeFileID = episode.EpisodeFileID // for deleting
			break
		}
	}

	// Make sure we got the correct episode.
	if episodeID == 0 {
		return http.StatusNotFound, "Episode not found in this Sonarr instance."
	}

	// Check if the instance is rate limited.
	if !sonarrInstance.DelOK() {
		return http.StatusLocked, "This Sonarr instance is rate limited." +
			"Too many deletes through the Notifiarr client in the last hour."
	}

	mnd.Log.Trace(reqID, response.Action, "episode: unmonitorSonarrEpisode", episodeID, episodeFileID)

	if response.Action == "delete" {
		// Delete the Episode File if the action is delete.
		err = sonarrInstance.DeleteEpisodeFileContext(ctx, episodeFileID)
	} else {
		// Unmonitor the Episode if the action is unmonitor.
		_, err = sonarrInstance.MonitorEpisodeContext(ctx, []int64{episodeID}, false)
	}

	if err != nil {
		return parseStarrError(err)
	}

	return http.StatusOK, "OK"
}
