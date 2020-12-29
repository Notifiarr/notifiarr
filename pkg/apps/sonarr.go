package apps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

// sonarrHandlers is called once on startup to register the web API paths.
func (a *Apps) sonarrHandlers() {
	a.HandleAPIpath(Sonarr, "/add", sonarrAddSeries, "POST")
	a.HandleAPIpath(Sonarr, "/check/{tvdbid:[0-9]+}", sonarrCheckSeries, "GET")
	a.HandleAPIpath(Sonarr, "/search/{query}", sonarrSearchSeries, "GET")
	a.HandleAPIpath(Sonarr, "/qualityProfiles", sonarrProfiles, "GET")
	a.HandleAPIpath(Sonarr, "/languageProfiles", sonarrLangProfiles, "GET")
	a.HandleAPIpath(Sonarr, "/rootFolder", sonarrRootFolders, "GET")
}

// SonarrConfig represents the input data for a Sonarr server.
type SonarrConfig struct {
	*starr.Config
	sonarr *sonarr.Sonarr
}

func (r *SonarrConfig) setup(timeout time.Duration) {
	r.sonarr = sonarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}
}

func sonarrRootFolders(r *http.Request) (int, interface{}) {
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

func sonarrProfiles(r *http.Request) (int, interface{}) {
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

func sonarrLangProfiles(r *http.Request) (int, interface{}) {
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

func sonarrCheckSeries(r *http.Request) (int, interface{}) {
	tvdbid, _ := strconv.ParseInt(mux.Vars(r)["tvdbid"], 10, 64)
	// Check for existing series.
	if m, err := getSonarr(r).GetSeries(tvdbid); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", tvdbid, ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func sonarrSearchSeries(r *http.Request) (int, interface{}) {
	// Get all movies
	series, err := getSonarr(r).GetAllSeries()
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting series: %w", err)
	}

	query := strings.TrimSpace(strings.ToLower(mux.Vars(r)["query"])) // in
	returnSeries := make([]map[string]interface{}, 0)                 // out

	for _, s := range series {
		if seriesSearch(query, s.Title, s.AlternateTitles) {
			b := map[string]interface{}{
				"id":     s.ID,
				"title":  s.Title,
				"first":  s.FirstAired,
				"next":   s.NextAiring,
				"prev":   s.PreviousAiring,
				"added":  s.Added,
				"status": s.Status,
				"exists": false,
				"path":   s.Path,
			}

			if s.Statistics != nil {
				b["exists"] = s.Statistics.SizeOnDisk > 0
			}

			returnSeries = append(returnSeries, b)
		}
	}

	return http.StatusOK, returnSeries
}

func seriesSearch(query, title string, alts []*sonarr.AlternateTitle) bool {
	if strings.Contains(strings.ToLower(title), query) {
		return true
	}

	for _, t := range alts {
		if strings.Contains(strings.ToLower(t.Title), query) {
			return true
		}
	}

	return false
}

func sonarrAddSeries(r *http.Request) (int, interface{}) {
	payload := &sonarr.AddSeriesInput{}
	// Extract payload and check for TMDB ID.
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TvdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTMDB)
	}

	app := getSonarr(r)
	// Check for existing series.
	if m, err := app.GetSeries(payload.TvdbID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.TvdbID, ErrExists)
	}

	series, err := app.AddSeries(payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding series: %w", err)
	}

	return http.StatusCreated, series
}
