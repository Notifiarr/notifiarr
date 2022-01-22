//nolint:dupl
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/gorilla/mux"
)

/* The site relies on release and quality profiles data from Radarr and Sonarr.
 * If someone has several instances, it causes slow page loads times.
 * So we made this file to aggregate responses from each of the app types.
 */

func (c *Client) aggregateTrash(req *http.Request) (int, interface{}) {
	var wait sync.WaitGroup
	defer wait.Wait()

	var input struct {
		Radarr struct { // used for "all"
			Instances notifiarr.IntList `json:"instances"`
		} `json:"radarr"`
		Sonarr struct { // used for "all"
			Instances notifiarr.IntList `json:"instances"`
		} `json:"sonarr"`
		Instances notifiarr.IntList `json:"instances"`
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

func (c *Client) aggregateTrashSonarr(ctx context.Context, wait *sync.WaitGroup,
	instances notifiarr.IntList) []*notifiarr.SonarrTrashPayload {
	output := []*notifiarr.SonarrTrashPayload{}
	// Create our known+requested instances, so we can write slice values in go routines.
	for i, app := range c.Config.Apps.Sonarr {
		if instance := i + 1; instances.Has(instance) {
			output = append(output, &notifiarr.SonarrTrashPayload{Instance: instance, Name: app.Name})
		}
	}

	var err error
	// Grab data for each requested instance in parallel/go routine.
	for idx := range output {
		wait.Add(1)

		go func(idx, instance int) {
			defer wait.Done()
			// Add the profiles, and/or error into our data structure/output data.
			app := c.Config.Apps.Sonarr[instance-1]
			if output[idx].QualityProfiles, err = app.GetQualityProfilesContext(ctx); err != nil {
				output[idx].Error = fmt.Sprintf("getting quality profiles: %v", err)
				c.Errorf("Handling Sonarr API request (%d): %s", instance, output[idx].Error)
			} else if output[idx].ReleaseProfiles, err = app.GetReleaseProfilesContext(ctx); err != nil {
				output[idx].Error = fmt.Sprintf("getting release profiles: %v", err)
				c.Errorf("Handling Sonarr API request (%d): %s", instance, output[idx].Error)
			}
		}(idx, output[idx].Instance)
	}

	return output
}

// This is basically a duplicate of the above code.
func (c *Client) aggregateTrashRadarr(ctx context.Context, wait *sync.WaitGroup,
	instances notifiarr.IntList) []*notifiarr.RadarrTrashPayload {
	output := []*notifiarr.RadarrTrashPayload{}
	// Create our known+requested instances, so we can write slice values in go routines.
	for i, app := range c.Config.Apps.Radarr {
		if instance := i + 1; instances.Has(instance) {
			output = append(output, &notifiarr.RadarrTrashPayload{Instance: instance, Name: app.Name})
		}
	}

	var err error
	// Grab data for each requested instance in parallel/go routine.
	for idx := range output {
		wait.Add(1)

		go func(idx, instance int) {
			defer wait.Done()
			// Add the profiles, and/or error into our data structure/output data.
			app := c.Config.Apps.Radarr[instance-1]
			if output[idx].QualityProfiles, err = app.GetQualityProfilesContext(ctx); err != nil {
				output[idx].Error = fmt.Sprintf("getting quality profiles: %v", err)
				c.Errorf("Handling Radarr API request (%d): %s", instance, output[idx].Error)
			} else if output[idx].CustomFormats, err = app.GetCustomFormatsContext(ctx); err != nil {
				output[idx].Error = fmt.Sprintf("getting custom formats: %v", err)
				c.Errorf("Handling Radarr API request (%d): %s", instance, output[idx].Error)
			}
		}(idx, output[idx].Instance)
	}

	return output
}
