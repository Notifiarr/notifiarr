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
	"golift.io/starr/sonarr"
)

// sonarrHandlers is called once on startup to register the web API paths.
func (a *Apps) sonarrHandlers() {
	a.HandleAPIpath(starr.Sonarr, "/add", sonarrAddSeries, "POST")
	a.HandleAPIpath(starr.Sonarr, "/check/{tvdbid:[0-9]+}", sonarrCheckSeries, "GET")
	a.HandleAPIpath(starr.Sonarr, "/get/{seriesid:[0-9]+}", sonarrGetSeries, "GET")
	a.HandleAPIpath(starr.Sonarr, "/getEpisodes/{seriesid:[0-9]+}", sonarrGetEpisodes, "GET")
	a.HandleAPIpath(starr.Sonarr, "/unmonitor/{episodeid:[0-9]+}", sonarrUnmonitorEpisode, "GET")
	a.HandleAPIpath(starr.Sonarr, "/languageProfiles", sonarrLangProfiles, "GET")
	a.HandleAPIpath(starr.Sonarr, "/qualityProfiles", sonarrGetQualityProfiles, "GET")
	a.HandleAPIpath(starr.Sonarr, "/qualityProfile", sonarrGetQualityProfile, "GET")
	a.HandleAPIpath(starr.Sonarr, "/qualityProfile", sonarrAddQualityProfile, "POST")
	a.HandleAPIpath(starr.Sonarr, "/qualityProfile/{profileID:[0-9]+}", sonarrUpdateQualityProfile, "PUT")
	a.HandleAPIpath(starr.Sonarr, "/releaseProfiles", sonarrGetReleaseProfiles, "GET")
	a.HandleAPIpath(starr.Sonarr, "/releaseProfile", sonarrAddReleaseProfile, "POST")
	a.HandleAPIpath(starr.Sonarr, "/releaseProfile/{profileID:[0-9]+}", sonarrUpdateReleaseProfile, "PUT")
	a.HandleAPIpath(starr.Sonarr, "/rootFolder", sonarrRootFolders, "GET")
	a.HandleAPIpath(starr.Sonarr, "/search/{query}", sonarrSearchSeries, "GET")
	a.HandleAPIpath(starr.Sonarr, "/tag", sonarrGetTags, "GET")
	a.HandleAPIpath(starr.Sonarr, "/tag/{tid:[0-9]+}/{label}", sonarrUpdateTag, "PUT")
	a.HandleAPIpath(starr.Sonarr, "/tag/{label}", sonarrSetTag, "PUT")
	a.HandleAPIpath(starr.Sonarr, "/update", sonarrUpdateSeries, "PUT")
	a.HandleAPIpath(starr.Sonarr, "/command/{commandid:[0-9]+}", sonarrStatusCommand, "GET")
	a.HandleAPIpath(starr.Sonarr, "/command", sonarrTriggerCommand, "POST")
	a.HandleAPIpath(starr.Sonarr, "/command/search/{seriesid:[0-9]+}", sonarrTriggerSearchSeries, "GET")
}

// SonarrConfig represents the input data for a Sonarr server.
type SonarrConfig struct {
	Name      string        `toml:"name" xml:"name"`
	Interval  cnfg.Duration `toml:"interval" xml:"interval"`
	StuckItem bool          `toml:"stuck_items" xml:"stuck_items"`
	Corrupt   string        `toml:"corrupt" xml:"corrupt"`
	Backup    string        `toml:"backup" xml:"backup"`
	*starr.Config
	*sonarr.Sonarr
	Errorf func(string, ...interface{}) `toml:"-" xml:"-"`
}

func (a *Apps) setupSonarr(timeout time.Duration) error {
	for idx := range a.Sonarr {
		if a.Sonarr[idx].Config == nil || a.Sonarr[idx].Config.URL == "" {
			return fmt.Errorf("%w: missing url: Sonarr config %d", ErrInvalidApp, idx+1)
		}

		a.Sonarr[idx].Debugf = a.DebugLog.Printf
		a.Sonarr[idx].Errorf = a.ErrorLog.Printf
		a.Sonarr[idx].setup(timeout)
	}

	return nil
}

func (r *SonarrConfig) setup(timeout time.Duration) {
	r.Sonarr = sonarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}

	r.URL = strings.TrimRight(r.URL, "/")

	if u, err := r.GetURL(); err != nil {
		r.Errorf("Checking Sonarr Path: %v", err)
	} else if u = strings.TrimRight(u, "/"); u != r.URL {
		r.Errorf("Sonarr URL fixed: %s -> %s (continuing)", r.URL, u)
		r.URL = u
	}
}

func sonarrAddSeries(req *http.Request) (int, interface{}) {
	var payload sonarr.AddSeriesInput
	// Extract payload and check for TVDB ID.
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.TvdbID == 0 {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoTVDB)
	}

	app := getSonarr(req)
	// Check for existing series.
	m, err := app.GetSeries(payload.TvdbID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, sonarrData(m[0])
	}

	series, err := app.AddSeries(&payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding series: %w", err)
	}

	return http.StatusCreated, series
}

func sonarrData(series *sonarr.Series) map[string]interface{} {
	hasFile := false
	if series.Statistics != nil {
		hasFile = series.Statistics.SizeOnDisk > 0
	}

	return map[string]interface{}{
		"id":        series.ID,
		"hasFile":   hasFile,
		"monitored": series.Monitored,
	}
}

func sonarrCheckSeries(req *http.Request) (int, interface{}) {
	tvdbid, _ := strconv.ParseInt(mux.Vars(req)["tvdbid"], mnd.Base10, mnd.Bits64)
	// Check for existing series.
	m, err := getSonarr(req).GetSeries(tvdbid)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, sonarrData(m[0])
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func sonarrGetSeries(req *http.Request) (int, interface{}) {
	seriesID, _ := strconv.ParseInt(mux.Vars(req)["seriesid"], mnd.Base10, mnd.Bits64)

	series, err := getSonarr(req).GetSeriesByID(seriesID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	}

	return http.StatusOK, series
}

func sonarrGetEpisodes(req *http.Request) (int, interface{}) {
	seriesID, _ := strconv.ParseInt(mux.Vars(req)["seriesid"], mnd.Base10, mnd.Bits64)

	episodes, err := getSonarr(req).GetSeriesEpisodes(seriesID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	}

	return http.StatusOK, episodes
}

func sonarrUnmonitorEpisode(req *http.Request) (int, interface{}) {
	episodeID, _ := strconv.ParseInt(mux.Vars(req)["episodeid"], mnd.Base10, mnd.Bits64)

	episodes, err := getSonarr(req).MonitorEpisode([]int64{episodeID}, false)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking series: %w", err)
	} else if len(episodes) != 1 {
		return http.StatusServiceUnavailable, fmt.Errorf("%w (%d): %v", ErrWrongCount, len(episodes), episodes)
	}

	return http.StatusOK, episodes[0]
}

func sonarrTriggerSearchSeries(req *http.Request) (int, interface{}) {
	seriesID, _ := strconv.ParseInt(mux.Vars(req)["seriesid"], mnd.Base10, mnd.Bits64)

	output, err := getSonarr(req).SendCommand(&sonarr.CommandRequest{
		Name:     "SeriesSearch",
		SeriesID: seriesID,
	})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("triggering series search: %w", err)
	}

	return http.StatusOK, output.Status
}

func sonarrTriggerCommand(req *http.Request) (int, interface{}) {
	var command sonarr.CommandRequest

	err := json.NewDecoder(req.Body).Decode(&command)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding command payload: %w", err)
	}

	output, err := getSonarr(req).SendCommand(&command)
	if err != nil {
		return http.StatusServiceUnavailable,
			fmt.Errorf("triggering command '%s' on series %d: %w", command.Name, command.SeriesID, err)
	}

	return http.StatusOK, output
}

func sonarrStatusCommand(req *http.Request) (int, interface{}) {
	commandID, _ := strconv.ParseInt(mux.Vars(req)["commandid"], mnd.Base10, mnd.Bits64)

	output, err := getSonarr(req).GetCommandStatus(commandID)
	if err != nil {
		return http.StatusServiceUnavailable,
			fmt.Errorf("getting command status for ID %d: %w", commandID, err)
	}

	return http.StatusOK, output
}

func sonarrLangProfiles(req *http.Request) (int, interface{}) {
	// Get the profiles from sonarr.
	profiles, err := getSonarr(req).GetLanguageProfiles()
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

func sonarrGetQualityProfile(req *http.Request) (int, interface{}) {
	// Get the profiles from sonarr.
	profiles, err := getSonarr(req).GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	return http.StatusOK, profiles
}

func sonarrGetQualityProfiles(req *http.Request) (int, interface{}) {
	// Get the profiles from sonarr.
	profiles, err := getSonarr(req).GetQualityProfiles()
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

func sonarrAddQualityProfile(req *http.Request) (int, interface{}) {
	var profile sonarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// Get the profiles from radarr.
	id, err := getSonarr(req).AddQualityProfile(&profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding profile: %w", err)
	}

	return http.StatusOK, id
}

func sonarrUpdateQualityProfile(req *http.Request) (int, interface{}) {
	var profile sonarr.QualityProfile

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
	err = getSonarr(req).UpdateQualityProfile(&profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating profile: %w", err)
	}

	return http.StatusOK, "OK"
}

func sonarrGetReleaseProfiles(req *http.Request) (int, interface{}) {
	// Get the profiles from sonarr.
	profiles, err := getSonarr(req).GetReleaseProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	return http.StatusOK, profiles
}

func sonarrAddReleaseProfile(req *http.Request) (int, interface{}) {
	var profile sonarr.ReleaseProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// Get the profiles from radarr.
	id, err := getSonarr(req).AddReleaseProfile(&profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding profile: %w", err)
	}

	return http.StatusOK, id
}

func sonarrUpdateReleaseProfile(req *http.Request) (int, interface{}) {
	var profile sonarr.ReleaseProfile

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
	err = getSonarr(req).UpdateReleaseProfile(&profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating profile: %w", err)
	}

	return http.StatusOK, "OK"
}

func sonarrRootFolders(req *http.Request) (int, interface{}) {
	// Get folder list from Sonarr.
	folders, err := getSonarr(req).GetRootFolders()
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

func sonarrSearchSeries(req *http.Request) (int, interface{}) {
	// Get all movies
	series, err := getSonarr(req).GetAllSeries()
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting series: %w", err)
	}

	query := strings.TrimSpace(mux.Vars(req)["query"]) // in
	resp := make([]map[string]interface{}, 0)          // out

	for _, item := range series {
		if seriesSearch(query, item.Title, item.AlternateTitles) {
			resp = append(resp, map[string]interface{}{
				"id":                item.ID,
				"title":             item.Title,
				"first":             item.FirstAired,
				"next":              item.NextAiring,
				"prev":              item.PreviousAiring,
				"added":             item.Added,
				"status":            item.Status,
				"path":              item.Path,
				"tvdbId":            item.TvdbID,
				"monitored":         item.Monitored,
				"qualityProfileId":  item.QualityProfileID,
				"seasonFolder":      item.SeasonFolder,
				"seriesType":        item.SeriesType,
				"languageProfileId": item.LanguageProfileID,
				"seasons":           item.Seasons,
				"exists":            item.Statistics != nil && item.Statistics.SizeOnDisk > 0,
			})
		}
	}

	return http.StatusOK, resp
}

func seriesSearch(query, title string, alts []*sonarr.AlternateTitle) bool {
	if strings.Contains(strings.ToLower(title), strings.ToLower(query)) {
		return true
	}

	for _, t := range alts {
		if strings.Contains(strings.ToLower(t.Title), strings.ToLower(query)) {
			return true
		}
	}

	return false
}

func sonarrGetTags(req *http.Request) (int, interface{}) {
	tags, err := getSonarr(req).GetTags()
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
	}

	return http.StatusOK, tags
}

func sonarrUpdateTag(req *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(req)["tid"])

	tagID, err := getSonarr(req).UpdateTag(id, mux.Vars(req)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
	}

	return http.StatusOK, tagID
}

func sonarrSetTag(req *http.Request) (int, interface{}) {
	tagID, err := getSonarr(req).AddTag(mux.Vars(req)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
	}

	return http.StatusOK, tagID
}

func sonarrUpdateSeries(req *http.Request) (int, interface{}) {
	var series sonarr.Series

	err := json.NewDecoder(req.Body).Decode(&series)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	err = getSonarr(req).UpdateSeries(series.ID, &series)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating series: %w", err)
	}

	return http.StatusOK, "sonarr seems to have worked"
}
