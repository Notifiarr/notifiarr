package apps

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
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
	a.HandleAPIpath(starr.Radarr, "/qualityProfile/{profileID:[0-9]+}", radarrDeleteQualityProfile, "DELETE")
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
	a.HandleAPIpath(starr.Radarr, "/customformats/{cfid:[0-9]+}", radarrDeleteCustomFormat, "DELETE")
	a.HandleAPIpath(starr.Radarr, "/importlist", radarrGetImportLists, "GET")
	a.HandleAPIpath(starr.Radarr, "/importlist", radarrAddImportList, "POST")
	a.HandleAPIpath(starr.Radarr, "/importlist/{ilid:[0-9]+}", radarrUpdateImportList, "PUT")
	a.HandleAPIpath(starr.Radarr, "/command/search/{movieid:[0-9]+}", radarrTriggerSearchMovie, "GET")
}

// RadarrConfig represents the input data for a Radarr server.
type RadarrConfig struct {
	starrConfig
	*starr.Config
	*radarr.Radarr `toml:"-" xml:"-" json:"-"`
	errorf         func(string, ...interface{}) `toml:"-" xml:"-" json:"-"`
}

// Enabled returns true if the Radarr instance is enabled and usable.
func (r *RadarrConfig) Enabled() bool {
	return r != nil && r.Config != nil && r.URL != "" && r.APIKey != "" && r.Timeout.Duration > 0
}

func (a *Apps) setupRadarr() error {
	for idx, app := range a.Radarr {
		if app.Config == nil || app.Config.URL == "" {
			return fmt.Errorf("%w: missing url: Radarr config %d", ErrInvalidApp, idx+1)
		}

		app.Config.Client = &http.Client{
			Timeout: app.Timeout.Duration,
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: exp.NewMetricsRoundTripper(string(starr.Radarr), &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: app.Config.ValidSSL}, //nolint:gosec
			}),
		}

		app.Debugf = a.Debugf
		app.errorf = a.Errorf
		app.URL = strings.TrimRight(app.URL, "/")
		app.Radarr = radarr.New(app.Config)
	}

	return nil
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

	// Check for existing movie.
	m, err := getRadarr(req).GetMovieContext(req.Context(), payload.TmdbID)
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
	movie, err := getRadarr(req).AddMovieContext(req.Context(), &payload)
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
		"tags":      movie.Tags,
	}
}

func radarrCheckMovie(req *http.Request) (int, interface{}) {
	tmdbID, _ := strconv.ParseInt(mux.Vars(req)["tmdbid"], mnd.Base10, mnd.Bits64)
	// Check for existing movie.
	m, err := getRadarr(req).GetMovieContext(req.Context(), tmdbID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, radarrData(m[0])
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func radarrGetMovie(req *http.Request) (int, interface{}) {
	movieID, _ := strconv.ParseInt(mux.Vars(req)["movieid"], mnd.Base10, mnd.Bits64)

	movie, err := getRadarr(req).GetMovieByIDContext(req.Context(), movieID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	}

	return http.StatusOK, movie
}

func radarrTriggerSearchMovie(req *http.Request) (int, interface{}) {
	movieID, _ := strconv.ParseInt(mux.Vars(req)["movieid"], mnd.Base10, mnd.Bits64)

	output, err := getRadarr(req).SendCommandContext(req.Context(), &radarr.CommandRequest{
		Name:     "MoviesSearch",
		MovieIDs: []int64{movieID},
	})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("triggering movie search: %w", err)
	}

	return http.StatusOK, output.Status
}

func radarrGetAllMovies(req *http.Request) (int, interface{}) {
	movies, err := getRadarr(req).GetMovieContext(req.Context(), 0)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking movie: %w", err)
	}

	return http.StatusOK, movies
}

func radarrQualityProfile(req *http.Request) (int, interface{}) {
	// Get the profiles from radarr.
	profiles, err := getRadarr(req).GetQualityProfilesContext(req.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	return http.StatusOK, profiles
}

func radarrQualityProfiles(req *http.Request) (int, interface{}) {
	// Get the profiles from radarr.
	profiles, err := getRadarr(req).GetQualityProfilesContext(req.Context())
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
	id, err := getRadarr(req).AddQualityProfileContext(req.Context(), &profile)
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
	err = getRadarr(req).UpdateQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating profile: %w", err)
	}

	return http.StatusOK, "OK"
}

func radarrDeleteQualityProfile(req *http.Request) (int, interface{}) {
	profileID, _ := strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profileID == 0 {
		return http.StatusBadRequest, ErrNonZeroID
	}

	// Delete the profile from radarr.
	err := getRadarr(req).DeleteQualityProfileContext(req.Context(), profileID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("deleting profile: %w", err)
	}

	return http.StatusOK, "OK"
}

func radarrRootFolders(req *http.Request) (int, interface{}) {
	// Get folder list from Radarr.
	folders, err := getRadarr(req).GetRootFoldersContext(req.Context())
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
	movies, err := getRadarr(req).GetMovieContext(req.Context(), 0)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting movies: %w", err)
	}

	query := strings.TrimSpace(strings.ToLower(mux.Vars(req)["query"])) // in
	output := make([]map[string]interface{}, 0)                         // out

	for _, movie := range movies {
		if movieSearch(query, []string{movie.Title, movie.OriginalTitle}, movie.AlternateTitles) {
			output = append(output, map[string]interface{}{
				"id":                  movie.ID,
				"title":               movie.Title,
				"cinemas":             movie.InCinemas,
				"digital":             movie.DigitalRelease,
				"physical":            movie.PhysicalRelease,
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

	return http.StatusOK, output
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
	tags, err := getRadarr(req).GetTagsContext(req.Context())
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
	}

	return http.StatusOK, tags
}

func radarrUpdateTag(req *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(req)["tid"])

	tag, err := getRadarr(req).UpdateTagContext(req.Context(), &starr.Tag{ID: id, Label: mux.Vars(req)["label"]})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
	}

	return http.StatusOK, tag.ID
}

func radarrSetTag(req *http.Request) (int, interface{}) {
	tag, err := getRadarr(req).AddTagContext(req.Context(), &starr.Tag{Label: mux.Vars(req)["label"]})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
	}

	return http.StatusOK, tag.ID
}

func radarrUpdateMovie(req *http.Request) (int, interface{}) {
	var movie radarr.Movie
	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&movie)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// Check for existing movie.
	err = getRadarr(req).UpdateMovieContext(req.Context(), movie.ID, &movie)
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
	err = getRadarr(req).AddExclusionsContext(req.Context(), exclusions)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding exclusions: %w", err)
	}

	return http.StatusOK, "added " + strconv.Itoa(len(exclusions)) + " exclusions"
}

func radarrGetExclusions(req *http.Request) (int, interface{}) {
	exclusions, err := getRadarr(req).GetExclusionsContext(req.Context())
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

	err := getRadarr(req).DeleteExclusionsContext(req.Context(), exclusions)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("deleting exclusions: %w", err)
	}

	return http.StatusOK, "deleted: " + strings.Join(strings.Split(ids, ","), ", ")
}

func radarrAddCustomFormat(req *http.Request) (int, interface{}) {
	var cusform radarr.CustomFormat

	err := json.NewDecoder(req.Body).Decode(&cusform)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	resp, err := getRadarr(req).AddCustomFormatContext(req.Context(), &cusform)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding custom format: %w", err)
	}

	return http.StatusOK, resp
}

func radarrGetCustomFormats(req *http.Request) (int, interface{}) {
	cusform, err := getRadarr(req).GetCustomFormatsContext(req.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting custom formats: %w", err)
	}

	return http.StatusOK, cusform
}

func radarrUpdateCustomFormat(req *http.Request) (int, interface{}) {
	var cusform radarr.CustomFormat
	if err := json.NewDecoder(req.Body).Decode(&cusform); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	cfID, _ := strconv.Atoi(mux.Vars(req)["cfid"])

	output, err := getRadarr(req).UpdateCustomFormatContext(req.Context(), &cusform, cfID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating custom format: %w", err)
	}

	return http.StatusOK, output
}

func radarrDeleteCustomFormat(req *http.Request) (int, interface{}) {
	cfID, _ := strconv.Atoi(mux.Vars(req)["cfid"])

	err := getRadarr(req).DeleteCustomFormatContext(req.Context(), cfID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("deleting custom format: %w", err)
	}

	return http.StatusOK, "OK"
}

func radarrGetImportLists(req *http.Request) (int, interface{}) {
	ilist, err := getRadarr(req).GetImportListsContext(req.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting import lists: %w", err)
	}

	return http.StatusOK, ilist
}

func radarrUpdateImportList(req *http.Request) (int, interface{}) {
	var ilist radarr.ImportList
	if err := json.NewDecoder(req.Body).Decode(&ilist); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	ilist.ID, _ = strconv.ParseInt(mux.Vars(req)["ilid"], mnd.Base10, mnd.Bits64)

	output, err := getRadarr(req).UpdateImportListContext(req.Context(), &ilist)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating import list: %w", err)
	}

	return http.StatusOK, output
}

func radarrAddImportList(req *http.Request) (int, interface{}) {
	var ilist radarr.ImportList
	if err := json.NewDecoder(req.Body).Decode(&ilist); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	output, err := getRadarr(req).CreateImportListContext(req.Context(), &ilist)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("creating import list: %w", err)
	}

	return http.StatusOK, output
}
