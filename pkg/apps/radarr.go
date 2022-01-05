package apps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	"golift.io/cnfg"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

// radarrHandlers is called once on startup to register the web API paths.
func (a *Apps) radarrHandlers() {
	a.HandleAPIpath(starr.Radarr, "/add", radarrAddMovie, "POST")
	a.HandleAPIpath(starr.Radarr, "/check/{tmdbid:[0-9]+}", radarrCheckMovie, "GET")
	a.HandleAPIpath(starr.Radarr, "/get/{movieid:[0-9]+}", radarrGetMovie, "GET")
	a.HandleAPIpath(starr.Radarr, "/get", radarrGetAllMovies, "GET")
	a.HandleAPIpath(starr.Radarr, "/qualityProfiles", radarrQualityProfiles, "GET")
	a.HandleAPIpath(starr.Radarr, "/qualityProfile", radarrQualityProfile, "GET")
	a.HandleAPIpath(starr.Radarr, "/qualityProfile", radarrAddQualityProfile, "POST")
	a.HandleAPIpath(starr.Radarr, "/qualityProfile/{profileID:[0-9]+}", radarrUpdateQualityProfile, "PUT")
	a.HandleAPIpath(starr.Radarr, "/rootFolder", radarrRootFolders, "GET")
	a.HandleAPIpath(starr.Radarr, "/search/{query}", radarrSearchMovie, "GET")
	a.HandleAPIpath(starr.Radarr, "/tag", radarrGetTags, "GET")
	a.HandleAPIpath(starr.Radarr, "/tag/{tid:[0-9]+}/{label}", radarrUpdateTag, "PUT")
	a.HandleAPIpath(starr.Radarr, "/tag/{label}", radarrSetTag, "PUT")
	a.HandleAPIpath(starr.Radarr, "/update", radarrUpdateMovie, "PUT")
	a.HandleAPIpath(starr.Radarr, "/exclusions", radarrGetExclusions, "GET")
	a.HandleAPIpath(starr.Radarr, "/exclusions", radarrAddExclusions, "POST")
	a.HandleAPIpath(starr.Radarr, "/exclusions/{eid:(?:[0-9],?)+}", radarrDelExclusions, "DELETE")
	a.HandleAPIpath(starr.Radarr, "/customformats", radarrGetCustomFormats, "GET")
	a.HandleAPIpath(starr.Radarr, "/customformats", radarrAddCustomFormat, "POST")
	a.HandleAPIpath(starr.Radarr, "/customformats/{cfid:[0-9]+}", radarrUpdateCustomFormat, "PUT")
	a.HandleAPIpath(starr.Radarr, "/importlist", radarrGetImportLists, "GET")
	a.HandleAPIpath(starr.Radarr, "/importlist", radarrAddImportList, "POST")
	a.HandleAPIpath(starr.Radarr, "/importlist/{ilid:[0-9]+}", radarrUpdateImportList, "PUT")
	a.HandleAPIpath(starr.Radarr, "/command/search/{movieid:[0-9]+}", radarrTriggerSearchMovie, "GET")
}

// RadarrConfig represents the input data for a Radarr server.
type RadarrConfig struct {
	Name      string        `toml:"name" xml:"name"`
	Interval  cnfg.Duration `toml:"interval" xml:"interval"`
	StuckItem bool          `toml:"stuck_items" xml:"stuck_items"`
	Corrupt   string        `toml:"corrupt" xml:"corrupt"`
	Backup    string        `toml:"backup" xml:"backup"`
	*starr.Config
	*radarr.Radarr
	Errorf func(string, ...interface{}) `toml:"-" xml:"-"`
}

func (a *Apps) setupRadarr(timeout time.Duration) error {
	for idx := range a.Radarr {
		if a.Radarr[idx].Config == nil || a.Radarr[idx].Config.URL == "" {
			return fmt.Errorf("%w: missing url: Radarr config %d", ErrInvalidApp, idx+1)
		}

		a.Radarr[idx].Debugf = a.DebugLog.Printf
		a.Radarr[idx].Errorf = a.ErrorLog.Printf
		a.Radarr[idx].setup(timeout)
	}

	return nil
}

func (r *RadarrConfig) setup(timeout time.Duration) {
	r.Radarr = radarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}

	r.URL = strings.TrimRight(r.URL, "/")

	if u, err := r.GetURL(); err != nil {
		r.Errorf("Checking Radarr Path: %v", err)
	} else if u = strings.TrimRight(u, "/"); u != r.URL {
		r.Errorf("Radarr URL fixed: %s -> %s (continuing)", r.URL, u)
		r.URL = u
	}
}

func radarrAddMovie(req *http.Request) (int, interface{}) {
	var payload radarr.AddMovieInput
	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TmdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTMDB)
	}

	app := getRadarr(req)
	// Check for existing movie.
	m, err := app.GetMovie(payload.TmdbID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, radarrData(m[0])
	}

	if payload.Title == "" {
		// Title must exist, even if it's wrong.
		payload.Title = strconv.FormatInt(payload.TmdbID, mnd.Base10)
	}

	if payload.MinimumAvailability == "" {
		payload.MinimumAvailability = "released"
	}

	// Add movie using fixed payload.
	movie, err := app.AddMovie(&payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding movie: %w", err)
	}

	return http.StatusCreated, movie
}

func radarrData(movie *radarr.Movie) map[string]interface{} {
	return map[string]interface{}{
		"id":        movie.ID,
		"hasFile":   movie.HasFile,
		"monitored": movie.Monitored,
	}
}

func radarrCheckMovie(req *http.Request) (int, interface{}) {
	tmdbID, _ := strconv.ParseInt(mux.Vars(req)["tmdbid"], mnd.Base10, mnd.Bits64)
	// Check for existing movie.
	m, err := getRadarr(req).GetMovie(tmdbID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, radarrData(m[0])
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func radarrGetMovie(req *http.Request) (int, interface{}) {
	movieID, _ := strconv.ParseInt(mux.Vars(req)["movieid"], mnd.Base10, mnd.Bits64)

	movie, err := getRadarr(req).GetMovieByID(movieID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	}

	return http.StatusOK, movie
}

func radarrTriggerSearchMovie(req *http.Request) (int, interface{}) {
	movieID, _ := strconv.ParseInt(mux.Vars(req)["movieid"], mnd.Base10, mnd.Bits64)

	output, err := getRadarr(req).SendCommand(&radarr.CommandRequest{
		Name:     "MoviesSearch",
		MovieIDs: []int64{movieID},
	})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("triggering movie search: %w", err)
	}

	return http.StatusOK, output.Status
}

func radarrGetAllMovies(req *http.Request) (int, interface{}) {
	movies, err := getRadarr(req).GetMovie(0)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	}

	return http.StatusOK, movies
}

func radarrQualityProfile(req *http.Request) (int, interface{}) {
	// Get the profiles from radarr.
	profiles, err := getRadarr(req).GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	return http.StatusOK, profiles
}

func radarrQualityProfiles(req *http.Request) (int, interface{}) {
	// Get the profiles from radarr.
	profiles, err := getRadarr(req).GetQualityProfiles()
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

func radarrAddQualityProfile(req *http.Request) (int, interface{}) {
	var profile radarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// Get the profiles from radarr.
	id, err := getRadarr(req).AddQualityProfile(&profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding profile: %w", err)
	}

	return http.StatusOK, id
}

func radarrUpdateQualityProfile(req *http.Request) (int, interface{}) {
	var profile radarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	profile.ID, _ = strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profile.ID == 0 {
		return http.StatusBadRequest, ErrNonZeroID
	}

	// Get the profiles from radarr.
	err = getRadarr(req).UpdateQualityProfile(&profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating profile: %w", err)
	}

	return http.StatusOK, "OK"
}

func radarrRootFolders(req *http.Request) (int, interface{}) {
	// Get folder list from Radarr.
	folders, err := getRadarr(req).GetRootFolders()
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

func radarrSearchMovie(req *http.Request) (int, interface{}) {
	// Get all movies
	movies, err := getRadarr(req).GetMovie(0)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting movies: %w", err)
	}

	query := strings.TrimSpace(strings.ToLower(mux.Vars(req)["query"])) // in
	returnMovies := make([]map[string]interface{}, 0)                   // out

	for _, movie := range movies {
		if movieSearch(query, []string{movie.Title, movie.OriginalTitle}, movie.AlternateTitles) {
			returnMovies = append(returnMovies, map[string]interface{}{
				"id":                  movie.ID,
				"title":               movie.Title,
				"cinemas":             movie.InCinemas,
				"status":              movie.Status,
				"exists":              movie.HasFile,
				"added":               movie.Added,
				"year":                movie.Year,
				"path":                movie.Path,
				"tmdbId":              movie.TmdbID,
				"qualityProfileId":    movie.QualityProfileID,
				"monitored":           movie.Monitored,
				"minimumAvailability": movie.MinimumAvailability,
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

func radarrGetTags(req *http.Request) (int, interface{}) {
	tags, err := getRadarr(req).GetTags()
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
	}

	return http.StatusOK, tags
}

func radarrUpdateTag(req *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(req)["tid"])

	tagID, err := getRadarr(req).UpdateTag(id, mux.Vars(req)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
	}

	return http.StatusOK, tagID
}

func radarrSetTag(req *http.Request) (int, interface{}) {
	tagID, err := getRadarr(req).AddTag(mux.Vars(req)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
	}

	return http.StatusOK, tagID
}

func radarrUpdateMovie(req *http.Request) (int, interface{}) {
	var movie radarr.Movie
	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&movie)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// Check for existing movie.
	err = getRadarr(req).UpdateMovie(movie.ID, &movie)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating movie: %w", err)
	}

	return http.StatusOK, "radarr seems to have worked"
}

func radarrAddExclusions(req *http.Request) (int, interface{}) {
	var exclusions []*radarr.Exclusion

	err := json.NewDecoder(req.Body).Decode(&exclusions)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// Get the profiles from radarr.
	err = getRadarr(req).AddExclusions(exclusions)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding exclusions: %w", err)
	}

	return http.StatusOK, "added " + strconv.Itoa(len(exclusions)) + " exclusions"
}

func radarrGetExclusions(req *http.Request) (int, interface{}) {
	exclusions, err := getRadarr(req).GetExclusions()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting exclusions: %w", err)
	}

	return http.StatusOK, exclusions
}

func radarrDelExclusions(req *http.Request) (int, interface{}) {
	ids := mux.Vars(req)["eid"]
	exclusions := []int64{}

	for _, s := range strings.Split(ids, ",") {
		if i, err := strconv.ParseInt(s, mnd.Base10, mnd.Bits64); err == nil {
			exclusions = append(exclusions, i)
		}
	}

	err := getRadarr(req).DeleteExclusions(exclusions)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("deleting exclusions: %w", err)
	}

	return http.StatusOK, "deleted: " + strings.Join(strings.Split(ids, ","), ", ")
}

func radarrAddCustomFormat(req *http.Request) (int, interface{}) {
	var cf radarr.CustomFormat

	err := json.NewDecoder(req.Body).Decode(&cf)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	resp, err := getRadarr(req).AddCustomFormat(&cf)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding custom format: %w", err)
	}

	return http.StatusOK, resp
}

func radarrGetCustomFormats(req *http.Request) (int, interface{}) {
	cf, err := getRadarr(req).GetCustomFormats()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting custom formats: %w", err)
	}

	return http.StatusOK, cf
}

func radarrUpdateCustomFormat(req *http.Request) (int, interface{}) {
	var cf radarr.CustomFormat
	if err := json.NewDecoder(req.Body).Decode(&cf); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	cfID, _ := strconv.Atoi(mux.Vars(req)["cfid"])

	output, err := getRadarr(req).UpdateCustomFormat(&cf, cfID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating custom format: %w", err)
	}

	return http.StatusOK, output
}

func radarrGetImportLists(req *http.Request) (int, interface{}) {
	il, err := getRadarr(req).GetImportLists()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting import lists: %w", err)
	}

	return http.StatusOK, il
}

func radarrUpdateImportList(req *http.Request) (int, interface{}) {
	var il radarr.ImportList
	if err := json.NewDecoder(req.Body).Decode(&il); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	il.ID, _ = strconv.ParseInt(mux.Vars(req)["ilid"], mnd.Base10, mnd.Bits64)

	output, err := getRadarr(req).UpdateImportList(&il)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating import list: %w", err)
	}

	return http.StatusOK, output
}

func radarrAddImportList(req *http.Request) (int, interface{}) {
	var il radarr.ImportList
	if err := json.NewDecoder(req.Body).Decode(&il); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	output, err := getRadarr(req).CreateImportList(&il)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("creating import list: %w", err)
	}

	return http.StatusOK, output
}
