package unmonitor

import (
	"context"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr/radarr"
)

func (c *cmd) unmonitorRadarrMovie(ctx context.Context, response *ResponseData, idx int) (int, string) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: unmonitorRadarrMovie", response.Instances[idx], response.TmdbID)
	defer mnd.Log.Trace(reqID, "end: unmonitorRadarrMovie", response.Instances[idx], response.TmdbID)

	radarrInstance := c.Apps.Radarr[response.Instances[idx]]

	// Get the internal RadarrMovie ID from Radarr using the TMDb ID.
	movie, err := radarrInstance.GetMovieContext(ctx, &radarr.GetMovie{TMDBID: response.TmdbID})
	if err != nil {
		return parseStarrError(err)
	}

	// Make sure we got the correct movie.
	if len(movie) == 0 || response.TmdbID != movie[0].TmdbID || response.TmdbID == 0 {
		return http.StatusNotFound, "Movie not found in this Radarr instance."
	}

	// Check if the instance is rate limited.
	if !radarrInstance.DelOK() {
		return http.StatusLocked, "This Radarr instance is rate limited. " +
			"Too many deletes through the Notifiarr client in the last hour."
	}

	mnd.Log.Trace(reqID, response.Action, "movie: unmonitorRadarrMovie", movie[0].ID)

	if response.Action == "delete" {
		// Delete the Movie File if the action is delete.
		err = radarrInstance.DeleteMovieFilesContext(ctx, movie[0].MovieFile.ID)
	} else { // Otherwise, only unmonitor the movie.
		movie[0].Monitored = false
		_, err = radarrInstance.UpdateMovieContext(ctx, movie[0].ID, movie[0], false)
	}

	if err != nil {
		return parseStarrError(err)
	}

	return http.StatusOK, "OK"
}
