package dnclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"golift.io/starr/radarr"
)

// radarrHandlers is called once on startup to register the web API paths.
func (c *Client) radarrHandlers() {
	c.handleAPIpath(Radarr, "/add", c.radarrAddMovie, "POST")
	c.handleAPIpath(Radarr, "/search/{query}", c.radarrSearchMovie, "GET")
	c.handleAPIpath(Radarr, "/check/{tmdbid:[0-9]+}", c.radarrCheckMovie, "GET")
	c.handleAPIpath(Radarr, "/qualityProfiles", c.radarrProfiles, "GET")
	c.handleAPIpath(Radarr, "/rootFolder", c.radarrRootFolders, "GET")
}

func (c *Client) radarrRootFolders(r *http.Request) (int, interface{}) {
	// Get folder list from Radarr.
	folders, err := getRadarr(r).GetRootFolders()
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
	// Get the profiles from radarr.
	profiles, err := getRadarr(r).GetQualityProfiles()
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

func (c *Client) radarrCheckMovie(r *http.Request) (int, interface{}) {
	tmdbID, _ := strconv.ParseInt(mux.Vars(r)["tmdbid"], 10, 64)
	// Check for existing movie.
	if m, err := getRadarr(r).GetMovie(tmdbID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", tmdbID, ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func (c *Client) radarrSearchMovie(r *http.Request) (int, interface{}) {
	// Get all movies
	movies, err := getRadarr(r).GetMovie(0)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting movies: %w", err)
	}

	query := strings.TrimSpace(strings.ToLower(mux.Vars(r)["query"])) // in
	returnMovies := make([]map[string]interface{}, 0)                 // out

	for _, movie := range movies {
		if movieSearch(query, []string{movie.Title, movie.OriginalTitle}, movie.AlternateTitles) {
			returnMovies = append(returnMovies, map[string]interface{}{
				"id":      movie.ID,
				"title":   movie.Title,
				"cinemas": movie.InCinemas,
				"status":  movie.Status,
				"exists":  movie.HasFile,
				"added":   movie.Added,
				"year":    movie.Year,
				"path":    movie.Path,
			})
		}
	}

	return http.StatusOK, returnMovies
}

func movieSearch(query string, titles []string, alts []*radarr.AlternativeTitle) bool {
	for _, t := range titles {
		if t != "" && strings.Contains(strings.ToLower(t), query) {
			return true
		}
	}

	for _, t := range alts {
		if strings.Contains(strings.ToLower(t.Title), query) {
			return true
		}
	}

	return false
}

func (c *Client) radarrAddMovie(r *http.Request) (int, interface{}) {
	payload := &radarr.AddMovieInput{}
	// Extract payload and check for TMDB ID.
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TmdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTMDB)
	}

	radar := getRadarr(r)
	// Check for existing movie.
	if m, err := radar.GetMovie(payload.TmdbID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.TmdbID, ErrExists)
	}

	if payload.Title == "" {
		// Title must exist, even if it's wrong.
		payload.Title = strconv.FormatInt(payload.TmdbID, 10)
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
