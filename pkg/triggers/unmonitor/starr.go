package unmonitor

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"golift.io/starr"
)

// unmonitorStarrContent loops through the provided instance list and unmonitors or deletes the episode or movie requested.
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

// parseStarrError parses a starr package error and returns the status code and error message from the underlying api request.
func parseStarrError(err error) (int, string) {
	var starrErr *starr.ReqError
	if errors.As(err, &starrErr) {
		return starrErr.Code, string(starrErr.Body)
	}

	return http.StatusInternalServerError, err.Error()
}
