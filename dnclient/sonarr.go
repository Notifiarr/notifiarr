//nolint:dupl
package dnclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golift.io/starr/sonarr"
)

// sonarrHandlers is called once on startup to register the web API paths.
func (c *Client) sonarrHandlers() {
	c.handleAPIpath(Sonarr, "/add/{id:[0-9]+}", c.sonarrAddSeries, "POST")
	c.handleAPIpath(Sonarr, "/check/{id:[0-9]+}/{tvdbid:[0-9]+}", c.sonarrCheckSeries, "GET")
	c.handleAPIpath(Sonarr, "/qualityProfiles/{id:[0-9]+}", c.sonarrProfiles, "GET")
	c.handleAPIpath(Sonarr, "/languageProfiles/{id:[0-9]+}", c.sonarrLangProfiles, "GET")
	c.handleAPIpath(Sonarr, "/rootFolder/{id:[0-9]+}", c.sonarrRootFolders, "GET")
}

func (c *Client) sonarrRootFolders(r *http.Request) (int, interface{}) {
	// Get folder list from Sonarr.
	folders, err := getSonarr(r).GetRootFolders()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting folders: %w", err)
	}

	// Format folder list into a nice path=>freesSpace map.
	p := make(map[string]int64)
	for i := range folders {
		p[folders[i].Path] = folders[i].FreeSpace
	}

	return http.StatusOK, p
}

func (c *Client) sonarrProfiles(r *http.Request) (int, interface{}) {
	// Get the profiles from sonarr.
	profiles, err := getSonarr(r).GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int64]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) sonarrLangProfiles(r *http.Request) (int, interface{}) {
	// Get the profiles from sonarr.
	profiles, err := getSonarr(r).GetLanguageProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting language profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int64]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) sonarrCheckSeries(r *http.Request) (int, interface{}) {
	tvdbid, _ := strconv.ParseInt(mux.Vars(r)["tvdbid"], 10, 64)
	// Check for existing series.
	if m, err := getSonarr(r).GetSeries(tvdbid); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", tvdbid, ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func (c *Client) sonarrAddSeries(r *http.Request) (int, interface{}) {
	payload := &sonarr.AddSeriesInput{}
	// Extract payload and check for TMDB ID.
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TvdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTMDB)
	}

	sonar := getSonarr(r)
	// Check for existing series.
	if m, err := sonar.GetSeries(payload.TvdbID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.TvdbID, ErrExists)
	}

	series, err := sonar.AddSeries(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding series: %w", err)
	}

	return http.StatusCreated, series
}
