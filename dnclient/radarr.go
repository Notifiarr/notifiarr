package dnclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

// RadarrConfig represents the input data for a Radarr server.
type RadarrConfig struct {
	*starr.Config
	*radarr.Radarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// radarrHandlers is called once on startup to register the web API paths.
func (c *Client) radarrHandlers() {
	c.serveAPIpath(Radarr, "/add/{id:[0-9]+}", "POST", c.radarrAddMovie)
	c.serveAPIpath(Radarr, "/check/{id:[0-9]+}/{tmdbid:[0-9]+}", "GET", c.radarrCheckMovie)
	c.serveAPIpath(Radarr, "/qualityProfiles/{id:[0-9]+}", "GET", c.radarrProfiles)
	c.serveAPIpath(Radarr, "/rootFolder/{id:[0-9]+}", "GET", c.radarrRootFolders)
}

func (c *Client) radarrRootFolders(r *http.Request) (int, interface{}) {
	radar := getRadarr(r)

	// Get folder list from Radarr.
	folders, err := radar.GetRootFolders()
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

func (c *Client) radarrProfiles(r *http.Request) (int, interface{}) {
	radar := getRadarr(r)

	// Get the profiles from radarr.
	profiles, err := radar.GetQualityProfiles()
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

func (c *Client) radarrCheckMovie(r *http.Request) (int, interface{}) {
	radar := getRadarr(r)
	tmdbID, _ := strconv.Atoi(mux.Vars(r)["tmdbid"])

	// Check for existing movie.
	if m, err := radar.GetMovie(tmdbID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", tmdbID, ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func (c *Client) radarrAddMovie(r *http.Request) (int, interface{}) {
	radar := getRadarr(r)

	// Extract payload and check for TMDB ID.
	payload := &radarr.AddMovieInput{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TmdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTMDB)
	}

	// Check for existing movie.
	if m, err := radar.GetMovie(payload.TmdbID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.TmdbID, ErrExists)
	}

	if payload.Title == "" {
		// Title must exist, even if it's wrong.
		payload.Title = strconv.Itoa(payload.TmdbID)
	}

	if payload.MinimumAvailability == "" {
		payload.MinimumAvailability = "released"
	}

	// Add movie using fixed payload.
	movie, err := radar.AddMovie(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding movie: %w", err)
	}

	return http.StatusCreated, movie
}
