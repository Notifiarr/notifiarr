//nolint:dupl
package dnclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

// SonarrConfig represents the input data for a Sonarr server.
type SonarrConfig struct {
	*starr.Config
	*sonarr.Sonarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// sonarrHandlers is called once on startup to register the web API paths.
func (c *Client) sonarrHandlers() {
	c.serveAPIpath(Sonarr, "/add/{id:[0-9]+}", "POST", c.sonarrAddSeries)
	c.serveAPIpath(Sonarr, "/check/{id:[0-9]+}/{tvdbid:[0-9]+}", "GET", c.sonarrCheckSeries)
	c.serveAPIpath(Sonarr, "/qualityProfiles/{id:[0-9]+}", "GET", c.sonarrProfiles)
	c.serveAPIpath(Sonarr, "/languageProfiles/{id:[0-9]+}", "GET", c.sonarrLangProfiles)
	c.serveAPIpath(Sonarr, "/rootFolder/{id:[0-9]+}", "GET", c.sonarrRootFolders)
}

func (c *Client) sonarrRootFolders(r *http.Request) (int, interface{}) {
	sonar := getSonarr(r)

	// Get folder list from Sonarr.
	folders, err := sonar.GetRootFolders()
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
	sonar := getSonarr(r)

	// Get the profiles from sonarr.
	profiles, err := sonar.GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) sonarrLangProfiles(r *http.Request) (int, interface{}) {
	sonar := getSonarr(r)

	// Get the profiles from sonarr.
	profiles, err := sonar.GetLanguageProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting language profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) sonarrCheckSeries(r *http.Request) (int, interface{}) {
	sonar := getSonarr(r)
	tvdbid, _ := strconv.Atoi(mux.Vars(r)["tvdbid"])

	// Check for existing series.
	if m, err := sonar.GetSeries(tvdbid); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", tvdbid, ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func (c *Client) sonarrAddSeries(r *http.Request) (int, interface{}) {
	sonar := getSonarr(r)

	// Extract payload and check for TMDB ID.
	payload := &sonarr.AddSeriesInput{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TvdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTMDB)
	}

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
