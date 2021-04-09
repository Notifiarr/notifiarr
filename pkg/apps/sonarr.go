package apps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golift.io/cnfg"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

// sonarrHandlers is called once on startup to register the web API paths.
func (a *Apps) sonarrHandlers() {
	a.HandleAPIpath(Sonarr, "/add", sonarrAddSeries, "POST")
	a.HandleAPIpath(Sonarr, "/check/{tvdbid:[0-9]+}", sonarrCheckSeries, "GET")
	a.HandleAPIpath(Sonarr, "/get/{seriesid:[0-9]+}", sonarrGetSeries, "GET")
	a.HandleAPIpath(Sonarr, "/languageProfiles", sonarrLangProfiles, "GET")
	a.HandleAPIpath(Sonarr, "/qualityProfiles", sonarrProfiles, "GET")
	a.HandleAPIpath(Sonarr, "/rootFolder", sonarrRootFolders, "GET")
	a.HandleAPIpath(Sonarr, "/search/{query}", sonarrSearchSeries, "GET")
	a.HandleAPIpath(Sonarr, "/tag", sonarrGetTags, "GET")
	a.HandleAPIpath(Sonarr, "/tag/{tid:[0-9]+}/{label}", sonarrUpdateTag, "PUT")
	a.HandleAPIpath(Sonarr, "/tag/{label}", sonarrSetTag, "PUT")
	a.HandleAPIpath(Sonarr, "/update", sonarrUpdateSeries, "PUT")
}

// SonarrConfig represents the input data for a Sonarr server.
type SonarrConfig struct {
	Name     string
	Interval cnfg.Duration
	*starr.Config
	sonarr *sonarr.Sonarr
}

func (r *SonarrConfig) setup(timeout time.Duration) {
	r.sonarr = sonarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}
}

func sonarrAddSeries(r *http.Request) (int, interface{}) {
	var payload sonarr.AddSeriesInput
	// Extract payload and check for TVDB ID.
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TvdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTMDB)
	}

	app := getSonarr(r)
	// Check for existing series.
	m, err := app.GetSeries(payload.TvdbID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.TvdbID, ErrExists)
	}

	series, err := app.AddSeries(&payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding series: %w", err)
	}

	return http.StatusCreated, series
}

func sonarrCheckSeries(r *http.Request) (int, interface{}) {
	tvdbid, _ := strconv.ParseInt(mux.Vars(r)["tvdbid"], 10, 64)
	// Check for existing series.
	m, err := getSonarr(r).GetSeries(tvdbid)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%s: %w", m[0].ID, ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func sonarrGetSeries(r *http.Request) (int, interface{}) {
	seriesID, _ := strconv.ParseInt(mux.Vars(r)["seriesid"], 10, 64)

	series, err := getSonarr(r).GetSeriesByID(seriesID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	}

	return http.StatusOK, series
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
				"id":                s.ID,
				"title":             s.Title,
				"first":             s.FirstAired,
				"next":              s.NextAiring,
				"prev":              s.PreviousAiring,
				"added":             s.Added,
				"status":            s.Status,
				"path":              s.Path,
				"tvdbId":            s.TvdbID,
				"monitored":         s.Monitored,
				"qualityProfileId":  s.QualityProfileID,
				"seasonFolder":      s.SeasonFolder,
				"seriesType":        s.SeriesType,
				"languageProfileId": s.LanguageProfileID,
				"seasons":           s.Seasons,
				"exists":            false,
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

func sonarrGetTags(r *http.Request) (int, interface{}) {
	tags, err := getSonarr(r).GetTags()
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
	}

	return http.StatusOK, tags
}

func sonarrUpdateTag(r *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(r)["tid"])

	tagID, err := getSonarr(r).UpdateTag(id, mux.Vars(r)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
	}

	return http.StatusOK, tagID
}

func sonarrSetTag(r *http.Request) (int, interface{}) {
	tagID, err := getSonarr(r).AddTag(mux.Vars(r)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
	}

	return http.StatusOK, tagID
}

func sonarrUpdateSeries(r *http.Request) (int, interface{}) {
	var series sonarr.Series

	err := json.NewDecoder(r.Body).Decode(&series)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	err = getSonarr(r).UpdateSeries(series.ID, &series)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating series: %w", err)
	}

	return http.StatusOK, "sonarr seems to have worked"
}
