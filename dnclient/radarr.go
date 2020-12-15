package dnclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"golift.io/starr"
)

// RadarrConfig represents the input data for a Radarr server.
type RadarrConfig struct {
	//	Name      string `json:"name" toml:"name" xml:"name" yaml:"name"`
	Root   string `json:"root_folder" toml:"root_folder" xml:"root_folder" yaml:"root_folder"`
	Search bool   `json:"search" toml:"search" xml:"search" yaml:"search"`
	*starr.Config
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Config) fixRadarrConfig() {
	for i := range c.Radarr {
		if c.Radarr[i].Timeout.Duration == 0 {
			c.Radarr[i].Timeout.Duration = c.Timeout.Duration
		}
	}
}

func (c *Client) logRadarr() {
	if count := len(c.Radarr); count == 1 {
		c.Printf(" => Radarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v, search:%v, root:%s",
			c.Radarr[0].URL, c.Radarr[0].APIKey != "", c.Radarr[0].Timeout,
			c.Radarr[0].ValidSSL, c.Radarr[0].Search, c.Radarr[0].Root)
	} else {
		c.Print(" => Radarr Config:", count, "servers")

		for _, f := range c.Radarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v, search:%v, root:%s",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.Search, f.Root)
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
	folders, err := radar.Radarr3RootFolders()
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
	profiles, err := radar.Radarr3QualityProfiles()
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

func (c *Client) radarrAddMovie(r *http.Request) (int, interface{}) {
	// Make sure the provided radarr id exists.
	radar := c.getRadarr(mux.Vars(r)["id"])
	if radar == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", mux.Vars(r)["id"], ErrNoRadarr)
	}

	// Extract payload and check for TMDB ID.
	payload := &starr.AddMovie{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TmdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTMDB)
	}

	// Check for existing movie.
	if m, err := radar.Radarr3Movie(payload.TmdbID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.TmdbID, ErrExists)
	}

	// Fix payload data.
	payload.Monitored = true
	payload.AddMovieOptions.SearchForMovie = radar.Search

	if payload.RootFolderPath == "" {
		payload.RootFolderPath = radar.Root
	}

	if payload.Title == "" {
		// Title must exist, even if it's wrong.
		payload.Title = strconv.Itoa(payload.TmdbID)
	}

	if payload.MinimumAvailability == "" {
		payload.MinimumAvailability = "released"
	}

	// Add movie using fixed payload.
	if err := radar.Radarr3AddMovie(payload); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding movie: %w", err)
	}

	return http.StatusCreated, fmt.Sprintf("added %d", payload.TmdbID)
}
