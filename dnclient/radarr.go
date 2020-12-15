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

// RadarrAddMovie is the data we expect to get from discord notifier when adding a Radarr Movie.
type RadarrAddMovie struct {
	Root string `json:"rootFolderPath"`   // optional
	QID  int    `json:"qualityProfileId"` // required
	TMDB int    `json:"tmdbId"`           // required if App = radarr
}

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

func (c *Client) radarrProfiles(r *http.Request) (int, string) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"]) // defaults to 0

	for i, radar := range c.Radarr {
		if i != id {
			continue
		}

		profiles, err := radar.Radarr3QualityProfiles()
		if err != nil {
			return http.StatusInternalServerError, err.Error()
		}

		b, err := json.Marshal(profiles)
		if err != nil {
			return http.StatusInternalServerError, err.Error()
		}

		return http.StatusOK, string(b)
	}

	return http.StatusUnprocessableEntity, ErrNoRadarr.Error()
}

func (c *Client) radarrAddMovie(r *http.Request) (int, string) {
	payload := &RadarrAddMovie{}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	if r.Method != "POST" {
		return http.StatusMethodNotAllowed, ErrOnlyPOST.Error()
	} else if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, err.Error()
	} else if payload.TMDB == 0 {
		return http.StatusUnprocessableEntity, ErrNoTMDB.Error()
	}

	for i, radar := range c.Radarr {
		if i != id-1 { // discordnotifier wants 1-indexes.
			continue
		}

		if payload.Root == "" {
			payload.Root = radar.Root
		}

		status, err := radar.addMovie(payload)
		if err != nil {
			return status, err.Error()
		}

		return status, fmt.Sprintf("added %d", payload.TMDB)
	}

	return http.StatusUnprocessableEntity, ErrNoRadarr.Error()
}

func (r *RadarrConfig) addMovie(p *RadarrAddMovie) (int, error) {
	m, err := r.Radarr3Movie(p.TMDB)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	}

	if len(m) > 0 {
		return http.StatusConflict, ErrExists
	}

	err = r.Radarr3AddMovie(&starr.AddMovie{
		TmdbID:              p.TMDB,
		Monitored:           true,
		QualityProfileID:    p.QID,
		MinimumAvailability: "released",
		AddMovieOptions:     starr.AddMovieOptions{SearchForMovie: r.Search},
		RootFolderPath:      p.Root,
	})
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding movie: %w", err)
	}

	return http.StatusOK, nil
}
