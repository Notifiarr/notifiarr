package cfsync

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
)

/* The site relies on release and quality profiles data from Radarr and Sonarr.
 * If someone has several instances, it causes slow page loads times.
 * So we made this file to aggregate responses from each of the app types.
 */

// Handler is passed into the webserver as an HTTP handler.
func (a *Action) Handler(req *http.Request) (int, interface{}) {
	return a.cmd.aggregateTrash(req)
}

func (c *cmd) aggregateTrash(req *http.Request) (int, interface{}) {
	var wait sync.WaitGroup
	defer wait.Wait()

	var input struct {
		Radarr struct { // used for "all"
			Instances website.IntList `json:"instances"`
		} `json:"radarr"`
		Sonarr struct { // used for "all"
			Instances website.IntList `json:"instances"`
		} `json:"sonarr"`
		Instances website.IntList `json:"instances"`
	}
	// Extract POST payload.
	err := json.NewDecoder(req.Body).Decode(&input)

	switch app := mux.Vars(req)["app"]; {
	default:
		return http.StatusBadRequest, fmt.Errorf("%w: %s", apps.ErrInvalidApp, app)
	case err != nil:
		return http.StatusBadRequest, fmt.Errorf("decoding POST payload: (app: %s) %w", app, err)
	case app == "sonarr":
		return http.StatusOK, c.aggregateTrashSonarr(req.Context(), &wait, input.Instances)
	case app == "radarr":
		return http.StatusOK, c.aggregateTrashRadarr(req.Context(), &wait, input.Instances)
	case app == "all":
		return http.StatusOK, map[string]interface{}{
			"radarr": c.aggregateTrashRadarr(req.Context(), &wait, input.Radarr.Instances),
			"sonarr": c.aggregateTrashSonarr(req.Context(), &wait, input.Sonarr.Instances),
		}
	}
}
