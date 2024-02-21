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

/*
 * If someone has several instances, it causes slow page loads times.
 * So we made this file to aggregate responses from each of the app types.
 */

// @Description  Returns custom format and related data for multiple Radarr instances at once. May be slow.
// @Summary      Retrieve custom format data from multiple Radarr instances.
// @Tags         TRaSH,Radarr
// @Produce      json
// @Accept       json
// @Param        request body TrashAggInput true "list of instances"
// @Success      200  {object} apps.Respond.apiResponse{message=[]RadarrTrashPayload} "contains app info included appStatus"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad input payload or missing app"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trash/radarr [post]
// @Security     ApiKeyAuth
//
//nolint:lll
func _() {}

// @Description  Returns custom format and related data for multiple Lidarr instances at once. May be slow.
// @Summary      Retrieve custom format data from multiple Lidarr instances.
// @Tags         TRaSH,Lidarr
// @Produce      json
// @Accept       json
// @Param        request body TrashAggInput true "list of instances"
// @Success      200  {object} apps.Respond.apiResponse{message=[]LidarrTrashPayload} "contains app info included appStatus"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad input payload or missing app"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trash/lidarr [post]
// @Security     ApiKeyAuth
//
//nolint:lll
func _() {}

// Handler is passed into the webserver as an HTTP handler.
// @Description  Returns custom format and related data for multiple Sonarr instances at once. May be slow.
// @Summary      Retrieve custom format data from multiple Sonarr instances.
// @Tags         TRaSH,Sonarr
// @Produce      json
// @Accept       json
// @Param        request body TrashAggInput true "list of instances"
// @Success      200  {object} apps.Respond.apiResponse{message=[]SonarrTrashPayload} "contains app info included appStatus"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad input payload or missing app"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/trash/sonarr [post]
// @Security     ApiKeyAuth
//
//nolint:lll
func (a *Action) Handler(req *http.Request) (int, interface{}) {
	return a.cmd.aggregateTrash(req)
}

// TrashAggInput is the data input for a Trash aggregate request.
type TrashAggInput struct {
	Instances clientinfo.IntList `json:"instances"`
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
		return http.StatusOK, c.aggregateTrashSonarr(req.Context(), &wait, input.Instances)
	case app == "radarr":
		return http.StatusOK, c.aggregateTrashRadarr(req.Context(), &wait, input.Instances)
	case app == "lidarr":
		return http.StatusOK, c.aggregateTrashLidarr(req.Context(), &wait, input.Instances)
	}
}
