package cfsync

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/gorilla/mux"
)

/* The site relies on release and quality profiles data from Radarr and Sonarr.
 * If someone has several instances, it causes slow page loads times.
 * So we made this file to aggregate responses from each of the app types.
 */

// @Description  Returns custom format and related data for multiple Radarr and/or Sonarr instances at once. May be slow.
// @Summary      Retrieve custom format data from multiple instances.
// @Tags         trash
// @Produce      json
// @Accept       json
// @Param        request body bothInstanceLists true "list of instances"
// @Success      200  {object} TrashAggOutput "contains app info included appStatus"
// @Failure      400  {object} string "bad input payload or missing app"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trash/all [post]
//
//nolint:lll
func _() {}

// @Description   Returns custom format and related data for multiple Radarr instances at once. May be slow.
// @Summary      Retrieve custom format data from multiple Radarr instances.
// @Tags         trash
// @Produce      json
// @Accept       json
// @Param        request body InstanceList true "list of instances"
// @Success      200  {object} []RadarrTrashPayload "contains app info included appStatus"
// @Failure      400  {object} string "bad input payload or missing app"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trash/radarr [post]
func _() {}

// @Description  Returns custom format and related data for multiple Sonarr instances at once. May be slow.
// @Summary      Retrieve custom format data from multiple Sonarr instances.
// @Tags         trash
// @Produce      json
// @Accept       json
// @Param        request body InstanceList true "list of instances"
// @Success      200  {object} []SonarrTrashPayload "contains app info included appStatus"
// @Failure      400  {object} string "bad input payload or missing app"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trash/sonarr [post]
// Handler is passed into the webserver as an HTTP handler.
func (a *Action) Handler(req *http.Request) (int, interface{}) {
	return a.cmd.aggregateTrash(req)
}

type InstanceList struct {
	Instances clientinfo.IntList `json:"instances"`
}

type bothInstanceLists struct {
	// Radarr or Sonarr instance list is required when app = all.
	Radarr InstanceList `json:"radarr"`
	// Sonarr or Radarr instance list is required when app = all.
	Sonarr InstanceList `json:"sonarr"`
}

// TrashAggInput is the data input for a Trash aggregate request.
type TrashAggInput struct {
	bothInstanceLists
	// Instances is required when app != all.
	Instances clientinfo.IntList `json:"instances"`
}

// TrashAggOutput is the data returned by the trash aggregate API endpoint.
type TrashAggOutput struct {
	Radarr []*RadarrTrashPayload `json:"radarr"`
	Sonarr []*SonarrTrashPayload `json:"sonarr"`
}

func (c *cmd) aggregateTrash(req *http.Request) (int, interface{}) {
	var wait sync.WaitGroup
	defer wait.Wait()

	var input TrashAggInput

	// Extract POST payload.
	err := json.NewDecoder(req.Body).Decode(&input)

	switch app := mux.Vars(req)["app"]; {
	default:
		return http.StatusBadRequest, fmt.Errorf("%w: %s", apps.ErrInvalidApp, app)
	case err != nil:
		return http.StatusBadRequest, fmt.Errorf("decoding POST payload: (app: %s) %w", app, err)
	case app == "sonarr":
		//	return http.StatusOK, &TrashAggOutput{Sonarr: c.aggregateTrashSonarr(req.Context(), &wait, input.Instances)}
		return http.StatusOK, c.aggregateTrashSonarr(req.Context(), &wait, input.Instances)
	case app == "radarr":
		// return http.StatusOK, &TrashAggOutput{Radarr: c.aggregateTrashRadarr(req.Context(), &wait, input.Instances)}
		return http.StatusOK, c.aggregateTrashRadarr(req.Context(), &wait, input.Instances)
	case app == "all":
		return http.StatusOK, &TrashAggOutput{
			Radarr: c.aggregateTrashRadarr(req.Context(), &wait, input.Radarr.Instances),
			Sonarr: c.aggregateTrashSonarr(req.Context(), &wait, input.Sonarr.Instances),
		}
	}
}
