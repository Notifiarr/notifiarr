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

func (c *Client) radarrMethods(r *mux.Router) {
	for _, r := range c.Config.Radarr {
		r.Radarr = radarr.New(r.Config)
	}

	r.Handle("/api/radarr/add/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.radarrAddMovie))).Methods("POST")
	r.Handle("/api/radarr/check/{id:[0-9]+}/{tmdbid:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.radarrCheckMovie))).Methods("GET")
	r.Handle("/api/radarr/qualityProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.radarrProfiles))).Methods("GET")
	r.Handle("/api/radarr/rootFolder/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.radarrRootFolders))).Methods("GET")
}

func (c *Config) fixRadarrConfig() {
	for i := range c.Radarr {
		if c.Radarr[i].Timeout.Duration == 0 {
			c.Radarr[i].Timeout.Duration = c.Timeout.Duration
		}
	}
}

// RadarrConfig represents the input data for a Radarr server.
type RadarrConfig struct {
	*starr.Config
	*radarr.Radarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logRadarr() {
	if count := len(c.Radarr); count == 1 {
		c.Printf(" => Radarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Radarr[0].URL, c.Radarr[0].APIKey != "", c.Radarr[0].Timeout, c.Radarr[0].ValidSSL)
	} else {
		c.Print(" => Radarr Config:", count, "servers")

		for _, f := range c.Radarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// getRadarr finds a radar based on the passed-in ID.
// Every Radarr handler calls this.
func (c *Client) getRadarr(id string) *RadarrConfig {
	j, _ := strconv.Atoi(id)

	for i, radar := range c.Radarr {
		if i != j-1 { // discordnotifier wants 1-indexes
			continue
		}

		return radar
	}

	return nil
}

func (c *Client) radarrRootFolders(r *http.Request) (int, interface{}) {
	// Make sure the provided radarr id exists.
	radar := c.getRadarr(mux.Vars(r)["id"])
	if radar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoRadarr)
	}

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
	// Make sure the provided radarr id exists.
	radar := c.getRadarr(mux.Vars(r)["id"])
	if radar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoRadarr)
	}

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
	radar := c.getRadarr(mux.Vars(r)["id"])
	if radar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoRadarr)
	}

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
	// Make sure the provided radarr id exists.
	radar := c.getRadarr(mux.Vars(r)["id"])
	if radar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoRadarr)
	}

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
