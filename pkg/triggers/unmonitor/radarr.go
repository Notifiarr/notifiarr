package unmonitor

import (
	"context"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr/radarr"
)

func (c *cmd) unmonitorRadarrMovie(ctx context.Context, response *ResponseData, idx int) (int, string) { //nolint:cyclop
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: unmonitorRadarrMovie", response.Instances[idx], response.TmdbID)
	defer mnd.Log.Trace(reqID, "end: unmonitorRadarrMovie", response.Instances[idx], response.TmdbID)

	radarrInstance := c.Apps.Radarr[response.Instances[idx]]

	// Get the internal RadarrMovie ID from Radarr using the TMDb ID.
	movie, err := radarrInstance.GetMovieContext(ctx, &radarr.GetMovie{TMDBID: response.TmdbID})
	if err != nil {
		return parseStarrError("getting movie", err)
	}

	// Make sure we got the correct movie.
	if len(movie) == 0 || response.TmdbID != movie[0].TmdbID || response.TmdbID == 0 {
		return http.StatusNotFound, "Movie not found in this Radarr instance."
	}

	mnd.Log.Trace(reqID, response.Action, "movie: unmonitorRadarrMovie", movie[0].ID)

	_, err = radarrInstance.UpdateMovieContext(ctx, movie[0].ID, movie[0], false)
	if err != nil {
		return parseStarrError("unmonitoring movie", err)
	}

	if response.Action != "delete" {
		return http.StatusOK, "OK"
	}

	// Check if the instance is rate limited.
	if !radarrInstance.DelOK() {
		return http.StatusLocked, "This Radarr instance is rate limited. " +
			"Too many deletes through the Notifiarr client in the last hour."
	}

	// Delete the Movie File if the action is delete.
	err = radarrInstance.DeleteMovieFilesContext(ctx, movie[0].MovieFile.ID)
	if err != nil {
		return parseStarrError("deleting movie file", err)
	}

	if response.MovieToo {
		// Delete the Movie also if the action is delete.
		err = radarrInstance.DeleteMovieContext(ctx, movie[0].ID, true, true)
		if err != nil {
			return parseStarrError("deleting movie file", err)
		}
	}

	return http.StatusOK, "OK"
}
