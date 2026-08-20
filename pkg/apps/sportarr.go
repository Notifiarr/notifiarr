//nolint:dupl
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
	wuzzy "github.com/paul-mannino/go-fuzzywuzzy"
	"golang.org/x/time/rate"
	"golift.io/starr"
	"golift.io/starr/debuglog"
	"golift.io/starr/sonarr"
)

// SportarrApp is the starr.App identity for Sportarr instances. Sportarr
// speaks the Sonarr v3 API, so instances ride the starr Sonarr client under
// their own app name for routes, logs, and metrics.
const SportarrApp starr.App = "Sportarr"

// sportarrHandlers is called once on startup to register the web API paths.
func (a *Apps) sportarrHandlers() { //nolint:funlen
	a.HandleAPIpath(SportarrApp, "/add", sportarrAddSeries, "POST")
	a.HandleAPIpath(SportarrApp, "/check/{tvdbid:[0-9]+}", sportarrCheckSeries, "GET")
	a.HandleAPIpath(SportarrApp, "/get/{seriesid:[0-9]+}", sportarrGetSeries, "GET")
	a.HandleAPIpath(SportarrApp, "/getEpisodes/{seriesid:[0-9]+}", sportarrGetEpisodes, "GET")
	a.HandleAPIpath(SportarrApp, "/unmonitor/{episodeid:[0-9]+}", sportarrUnmonitorEpisode, "GET")
	a.HandleAPIpath(SportarrApp, "/languageProfiles", sportarrLangProfiles, "GET")
	a.HandleAPIpath(SportarrApp, "/qualityProfiles", sportarrGetQualityProfiles, "GET")
	a.HandleAPIpath(SportarrApp, "/qualityProfile", sportarrGetQualityProfile, "GET")
	a.HandleAPIpath(SportarrApp, "/qualityProfile", sportarrAddQualityProfile, "POST")
	a.HandleAPIpath(SportarrApp, "/qualityProfile/{profileID:[0-9]+}", sportarrUpdateQualityProfile, "PUT")
	a.HandleAPIpath(SportarrApp, "/qualityProfile/{profileID:[0-9]+}", sportarrDeleteQualityProfile, "DELETE")
	a.HandleAPIpath(SportarrApp, "/qualityProfiles/all", sportarrDeleteAllQualityProfiles, "DELETE")
	a.HandleAPIpath(SportarrApp, "/releaseProfiles", sportarrGetReleaseProfiles, "GET")
	a.HandleAPIpath(SportarrApp, "/releaseProfile", sportarrAddReleaseProfile, "POST")
	a.HandleAPIpath(SportarrApp, "/releaseProfile/{profileID:[0-9]+}", sportarrUpdateReleaseProfile, "PUT")
	a.HandleAPIpath(SportarrApp, "/releaseProfile/{profileID:[0-9]+}", sportarrDeleteReleaseProfile, "DELETE")
	a.HandleAPIpath(SportarrApp, "/releaseProfiles/all", sportarrDeleteAllReleaseProfiles, "DELETE")
	a.HandleAPIpath(SportarrApp, "/customformats", sportarrGetCustomFormats, "GET")
	a.HandleAPIpath(SportarrApp, "/customformats", sportarrAddCustomFormat, "POST")
	a.HandleAPIpath(SportarrApp, "/customformats/{cfid:[0-9]+}", sportarrUpdateCustomFormat, "PUT")
	a.HandleAPIpath(SportarrApp, "/customformats/{cfid:[0-9]+}", sportarrDeleteCustomFormat, "DELETE")
	a.HandleAPIpath(SportarrApp, "/customformats/all", sportarrDeleteAllCustomFormats, "DELETE")
	a.HandleAPIpath(SportarrApp, "/importlist", sportarrGetImportLists, "GET")
	a.HandleAPIpath(SportarrApp, "/importlist", sportarrAddImportList, "POST")
	a.HandleAPIpath(SportarrApp, "/importlist/{ilid:[0-9]+}", sportarrUpdateImportList, "PUT")
	a.HandleAPIpath(SportarrApp, "/qualitydefinitions", sportarrGetQualityDefinitions, "GET")
	a.HandleAPIpath(SportarrApp, "/qualitydefinition", sportarrUpdateQualityDefinition, "PUT")
	a.HandleAPIpath(SportarrApp, "/rootFolder", sportarrRootFolders, "GET")
	a.HandleAPIpath(SportarrApp, "/naming", sportarrGetNaming, "GET")
	a.HandleAPIpath(SportarrApp, "/naming", sportarrUpdateNaming, "PUT")
	a.HandleAPIpath(SportarrApp, "/search/{query}", sportarrSearchSeries, "GET")
	a.HandleAPIpath(SportarrApp, "/tag", sportarrGetTags, "GET")
	a.HandleAPIpath(SportarrApp, "/tag/{tid:[0-9]+}/{label}", sportarrUpdateTag, "PUT")
	a.HandleAPIpath(SportarrApp, "/tag/{label}", sportarrSetTag, "PUT")
	a.HandleAPIpath(SportarrApp, "/update", sportarrUpdateSeries, "PUT")
	a.HandleAPIpath(SportarrApp, "/seasonPass", sportarrSeasonPass, "POST")
	a.HandleAPIpath(SportarrApp, "/command/{commandid:[0-9]+}", sportarrStatusCommand, "GET")
	a.HandleAPIpath(SportarrApp, "/command", sportarrTriggerCommand, "POST")
	a.HandleAPIpath(SportarrApp, "/command/search/{seriesid:[0-9]+}", sportarrTriggerSearchSeries, "GET")
	a.HandleAPIpath(SportarrApp, "/notification", sportarrGetNotifications, "GET")
	a.HandleAPIpath(SportarrApp, "/notification", sportarrUpdateNotification, "PUT")
	a.HandleAPIpath(SportarrApp, "/notification", sportarrAddNotification, "POST")
	a.HandleAPIpath(SportarrApp, "/queue/{queueID}", sportarrDeleteQueue, "DELETE")
	a.HandleAPIpath(SportarrApp, "/delete/{episodeFileID:[0-9]+}", sportarrDeleteEpisode, "DELETE")
}

type Sportarr struct {
	StarrApp       `json:"-" toml:"-" xml:"-"`
	*sonarr.Sonarr `json:"-" toml:"-" xml:"-"`
}

func getSportarr(r *http.Request) Sportarr {
	return r.Context().Value(SportarrApp).(Sportarr) //nolint:forcetypeassert
}

func (a *AppsConfig) setupSportarr() ([]Sportarr, error) {
	output := make([]Sportarr, len(a.Sportarr))

	for idx := range a.Sportarr {
		app := &a.Sportarr[idx]
		if err := checkUrl(app.URL, SportarrApp.String(), idx); err != nil {
			return nil, err
		}

		if mnd.Log.DebugEnabled() {
			app.Config.Client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  func(format string, v ...any) { mnd.Log.Debugf("remote", format, v...) },
				Caller:  metricMakerCallback(string(SportarrApp)),
				Redact:  []string{app.APIKey, app.Password, app.HTTPPass},
			})
		} else {
			app.Config.Client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			app.Config.Client.Transport = NewMetricsRoundTripper(SportarrApp.String(), app.Config.Client.Transport)
		}

		app.URL = strings.TrimRight(app.URL, "/")
		output[idx] = Sportarr{
			StarrApp: StarrApp{
				StarrConfig: a.Sportarr[idx],
			},
			Sonarr: sonarr.New(&app.Config),
		}

		if app.Deletes > 0 {
			output[idx].delLimit = rate.NewLimiter(rate.Every(1*time.Hour/time.Duration(app.Deletes)), app.Deletes)
		}
	}

	return output, nil
}

// @Description	Adds a new Series to Sportarr.
// @Summary		Add Sportarr Series
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64									true	"instance ID"
// @Param			POST		body		sonarr.AddSeriesInput					true	"new item content"
// @Success		201			{object}	apps.ApiResponse{message=sonarr.Series}	"series content"
// @Failure		400			{object}	apps.ApiResponse{message=string}		"bad json payload"
// @Failure		409			{object}	apps.ApiResponse{message=string}		"item already exists"
// @Failure		422			{object}	apps.ApiResponse{message=string}		"no item ID provided"
// @Failure		503			{object}	apps.ApiResponse{message=string}		"instance error during check"
// @Failure		500			{object}	apps.ApiResponse{message=string}		"instance error during add"
// @Failure		404			{object}	string									"bad token or api key"
// @Router			/sportarr/{instance}/add [post]
// @Security		ApiKeyAuth
func sportarrAddSeries(req *http.Request) (int, any) {
	var payload sonarr.AddSeriesInput
	// Extract payload and check for TVDB ID.
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	} else if payload.TvdbID == 0 {
		return apiError(http.StatusUnprocessableEntity, "0", ErrNoTVDB)
	}

	// Check for existing series.
	m, err := getSportarr(req).GetSeriesContext(req.Context(), payload.TvdbID)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "checking series", err)
	} else if len(m) > 0 {
		return http.StatusConflict, sportarrData(m[0])
	}

	series, err := getSportarr(req).AddSeriesContext(req.Context(), &payload)
	if err != nil {
		return apiError(http.StatusInternalServerError, "adding series", err)
	}

	return http.StatusCreated, series
}

func sportarrData(series *sonarr.Series) map[string]any {
	hasFile := false
	if series.Statistics != nil {
		hasFile = series.Statistics.SizeOnDisk > 0
	}

	return map[string]any{
		"id":        series.ID,
		"hasFile":   hasFile,
		"monitored": series.Monitored,
		"tags":      series.Tags,
	}
}

// @Description	Checks if a Sportarr Series already exists.
// @Summary		Check Sportarr Series Existence
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64								true	"instance ID"
// @Param			tvdbid		path		int64								true	"TVDB ID"
// @Success		201			{object}	apps.ApiResponse{message=string}	"series does not exist"
// @Failure		409			{object}	apps.ApiResponse{message=string}	"item already exists"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/check/{tvdbid} [get]
// @Security		ApiKeyAuth
func sportarrCheckSeries(req *http.Request) (int, any) {
	tvdbid, _ := strconv.ParseInt(mux.Vars(req)["tvdbid"], mnd.Base10, mnd.Bits64)
	// Check for existing series.
	m, err := getSportarr(req).GetSeriesContext(req.Context(), tvdbid)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "checking series", err)
	} else if len(m) > 0 {
		return http.StatusConflict, sportarrData(m[0])
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

// @Description	Returns a Sportarr Series by ID.
// @Summary		Get Sportarr Series
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64									true	"instance ID"
// @Param			seriesID	path		int64									true	"Series ID"
// @Success		201			{object}	apps.ApiResponse{message=sonarr.Series}	"series content"
// @Failure		503			{object}	apps.ApiResponse{message=string}		"instance error"
// @Failure		404			{object}	string									"bad token or api key"
// @Router			/sportarr/{instance}/get/{seriesID} [get]
// @Security		ApiKeyAuth
func sportarrGetSeries(req *http.Request) (int, any) {
	seriesID, _ := strconv.ParseInt(mux.Vars(req)["seriesid"], mnd.Base10, mnd.Bits64)

	series, err := getSportarr(req).GetSeriesByIDContext(req.Context(), seriesID)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "checking series", err)
	}

	return http.StatusOK, series
}

// @Description	Returns a Sportarr Series Episodes by Series ID.
// @Summary		Get Sportarr Series Episodes
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64										true	"instance ID"
// @Param			seriesID	path		int64										true	"Series ID"
// @Success		201			{object}	apps.ApiResponse{message=[]sonarr.Episode}	"episodes content"
// @Failure		503			{object}	apps.ApiResponse{message=string}			"instance error"
// @Failure		404			{object}	string										"bad token or api key"
// @Router			/sportarr/{instance}/getEpisodes/{seriesID} [get]
// @Security		ApiKeyAuth
func sportarrGetEpisodes(req *http.Request) (int, any) {
	seriesID, _ := strconv.ParseInt(mux.Vars(req)["seriesid"], mnd.Base10, mnd.Bits64)

	episodes, err := getSportarr(req).GetSeriesEpisodesContext(req.Context(), &sonarr.GetEpisode{SeriesID: seriesID})
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "checking series", err)
	}

	return http.StatusOK, episodes
}

// @Description	Unmonnitors and returns a Sportarr Series Episode.
// @Summary		Unmonnitors Sportarr Series Episode
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64										true	"instance ID"
// @Param			episodeID	path		int64										true	"Episode ID"
// @Success		201			{object}	apps.ApiResponse{message=sonarr.Episode}	"episode content"
// @Failure		503			{object}	apps.ApiResponse{message=string}			"instance error"
// @Failure		404			{object}	string										"bad token or api key"
// @Router			/sportarr/{instance}/unmonitor/{episodeID} [get]
// @Security		ApiKeyAuth
func sportarrUnmonitorEpisode(req *http.Request) (int, any) {
	episodeID, _ := strconv.ParseInt(mux.Vars(req)["episodeid"], mnd.Base10, mnd.Bits64)

	episodes, err := getSportarr(req).MonitorEpisodeContext(req.Context(), []int64{episodeID}, false)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "checking series", err)
	} else if len(episodes) != 1 {
		return http.StatusServiceUnavailable, fmt.Errorf("%w (%d): %v", ErrWrongCount, len(episodes), episodes)
	}

	return http.StatusOK, episodes[0]
}

// @Description	Trigger an Internet search for a Sportarr Series.
// @Summary		Search for Sportarr Series
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64								true	"instance ID"
// @Param			seriesID	path		int64								true	"Series ID"
// @Success		201			{object}	apps.ApiResponse{message=string}	"search status"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/command/search/{seriesID} [get]
// @Security		ApiKeyAuth
func sportarrTriggerSearchSeries(req *http.Request) (int, any) {
	seriesID, _ := strconv.ParseInt(mux.Vars(req)["seriesid"], mnd.Base10, mnd.Bits64)

	output, err := getSportarr(req).SendCommandContext(req.Context(), &sonarr.CommandRequest{
		Name:     "SeriesSearch",
		SeriesID: seriesID,
	})
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "triggering series search", err)
	}

	return http.StatusOK, output.Status
}

// @Description	Execute any command in Sportarr.
// @Summary		Execute Sportarr Command
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64												true	"instance ID"
// @Param			POST		body		sonarr.CommandRequest								true	"command content, must include series ID"
// @Success		201			{object}	apps.ApiResponse{message=sonarr.CommandResponse}	"command response"
// @Failure		400			{object}	apps.ApiResponse{message=string}					"invalid json input"
// @Failure		503			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/command [post]
// @Security		ApiKeyAuth
func sportarrTriggerCommand(req *http.Request) (int, any) {
	var command sonarr.CommandRequest

	err := json.NewDecoder(req.Body).Decode(&command)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding command payload", err)
	}

	output, err := getSportarr(req).SendCommandContext(req.Context(), &command)
	if err != nil {
		return http.StatusServiceUnavailable,
			fmt.Errorf("triggering command '%s' on series %d: %w", command.Name, command.SeriesID, err)
	}

	return http.StatusOK, output
}

// @Description	Check the status of an executed Sportarr Command.
// @Summary		Sonar Command Status
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64												true	"instance ID"
// @Param			commandID	path		int64												true	"Command ID returned by executing a command"
// @Success		201			{object}	apps.ApiResponse{message=sonarr.CommandResponse}	"command status"
// @Failure		503			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/command/{commandID} [get]
// @Security		ApiKeyAuth
func sportarrStatusCommand(req *http.Request) (int, any) {
	commandID, _ := strconv.ParseInt(mux.Vars(req)["commandid"], mnd.Base10, mnd.Bits64)

	output, err := getSportarr(req).GetCommandStatusContext(req.Context(), commandID)
	if err != nil {
		return http.StatusServiceUnavailable,
			fmt.Errorf("getting command status for ID %d: %w", commandID, err)
	}

	return http.StatusOK, output
}

// @Description	Fetches all Language Profiles from Sportarr.
// @Summary		Get Sportarr Language Profiles
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64										true	"instance ID"
// @Success		201			{object}	apps.ApiResponse{message=map[int64]string}	"map of ID to name"
// @Failure		500			{object}	apps.ApiResponse{message=string}			"instance error"
// @Failure		404			{object}	string										"bad token or api key"
// @Router			/sportarr/{instance}/languageProfiles [get]
// @Security		ApiKeyAuth
func sportarrLangProfiles(req *http.Request) (int, any) {
	// Get the profiles from sportarr.
	profiles, err := getSportarr(req).GetLanguageProfilesContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting language profiles", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int64]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

// @Description	Fetches all Quality Profiles Data from Sportarr.
// @Summary		Get Sportarr Quality Profile Data
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64												true	"instance ID"
// @Success		201			{object}	apps.ApiResponse{message=[]sonarr.QualityProfile}	"all profiles"
// @Failure		500			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/qualityProfile [get]
// @Security		ApiKeyAuth
func sportarrGetQualityProfile(req *http.Request) (int, any) {
	// Get the profiles from sportarr.
	profiles, err := getSportarr(req).GetQualityProfilesContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting profiles", err)
	}

	return http.StatusOK, profiles
}

// @Description	Fetches all Quality Profiles from Sportarr.
// @Summary		Get Sportarr Quality Profiles
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64										true	"instance ID"
// @Success		201			{object}	apps.ApiResponse{message=map[int64]string}	"map of ID to name"
// @Failure		500			{object}	apps.ApiResponse{message=string}			"instance error"
// @Failure		404			{object}	string										"bad token or api key"
// @Router			/sportarr/{instance}/qualityProfiles [get]
// @Security		ApiKeyAuth
func sportarrGetQualityProfiles(req *http.Request) (int, any) {
	// Get the profiles from sportarr.
	profiles, err := getSportarr(req).GetQualityProfilesContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting profiles", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int64]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

// @Description	Creates a new Sportarr Quality Profile.
// @Summary		Add Sportarr Quality Profile
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64								true	"instance ID"
// @Param			POST		body		sonarr.QualityProfile				true	"new item content"
// @Success		200			{object}	apps.ApiResponse{message=int64}		"new profile ID"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"json input error"
// @Failure		500			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/qualityProfile [post]
// @Security		ApiKeyAuth
func sportarrAddQualityProfile(req *http.Request) (int, any) {
	var profile sonarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	// Get the profiles from sportarr.
	id, err := getSportarr(req).AddQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return apiError(http.StatusInternalServerError, "adding profile", err)
	}

	return http.StatusOK, id
}

// @Description	Updates a Sportarr Quality Profile.
// @Summary		Update Sportarr Quality Profile
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64								true	"instance ID"
// @Param			profileID	path		int64								true	"profile ID to update"
// @Param			PUT			body		sonarr.QualityProfile				true	"updated item content"
// @Success		200			{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"json input error"
// @Failure		422			{object}	apps.ApiResponse{message=string}	"no profile ID"
// @Failure		500			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/qualityProfile/{profileID} [put]
// @Security		ApiKeyAuth
func sportarrUpdateQualityProfile(req *http.Request) (int, any) {
	var profile sonarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	profile.ID, _ = strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profile.ID == 0 {
		return http.StatusUnprocessableEntity, ErrNonZeroID
	}

	// Get the profiles from sportarr.
	_, err = getSportarr(req).UpdateQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return apiError(http.StatusInternalServerError, "updating profile", err)
	}

	return http.StatusOK, "OK"
}

// @Description	Removes a Sportarr Quality Profile.
// @Summary		Remove Sportarr Quality Profile
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64								true	"instance ID"
// @Param			profileID	path		int64								true	"profile ID to update"
// @Success		200			{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"no profile ID"
// @Failure		500			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/qualityProfile/{profileID} [delete]
// @Security		ApiKeyAuth
func sportarrDeleteQualityProfile(req *http.Request) (int, any) {
	profileID, _ := strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profileID == 0 {
		return http.StatusBadRequest, ErrNonZeroID
	}

	// Delete the profile from sportarr.
	err := getSportarr(req).DeleteQualityProfileContext(req.Context(), profileID)
	if err != nil {
		return apiError(http.StatusInternalServerError, "deleting profile", err)
	}

	return http.StatusOK, "OK"
}

// @Description	Removes all Sportarr Quality Profiles.
// @Summary		Remove Sportarr Quality Profiles
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64											true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=apps.deleteResponse}	"delete status"
// @Failure		500			{object}	apps.ApiResponse{message=string}				"instance error getting profiles"
// @Failure		404			{object}	string											"bad token or api key"
// @Router			/sportarr/{instance}/qualityProfiles/all [delete]
// @Security		ApiKeyAuth
func sportarrDeleteAllQualityProfiles(req *http.Request) (int, any) {
	// Get all the profiles from sportarr.
	profiles, err := getSportarr(req).GetQualityProfilesContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting profiles", err)
	}

	var (
		deleted int
		errs    []string
	)

	// Delete each profile from sportarr.
	for _, profile := range profiles {
		if err := getSportarr(req).DeleteQualityProfileContext(req.Context(), profile.ID); err != nil {
			errs = append(errs, err.Error())
			continue
		}

		deleted++
	}

	return http.StatusOK, &deleteResponse{
		Found:   len(profiles),
		Deleted: deleted,
		Errors:  errs,
	}
}

// @Description	Fetches all Release Profile Data from Sportarr.
// @Summary		Get Sportarr Release Profile Data
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64												true	"instance ID"
// @Success		201			{object}	apps.ApiResponse{message=[]sonarr.ReleaseProfile}	"all profiles"
// @Failure		500			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/releaseProfiles [get]
// @Security		ApiKeyAuth
func sportarrGetReleaseProfiles(req *http.Request) (int, any) {
	// Get the profiles from sportarr.
	profiles, err := getSportarr(req).GetReleaseProfilesContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting profiles", err)
	}

	return http.StatusOK, profiles
}

// @Description	Creates a new Sportarr Release Profile.
// @Summary		Add Sportarr Release Profile
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64								true	"instance ID"
// @Param			POST		body		sonarr.ReleaseProfile				true	"new item content"
// @Success		200			{object}	apps.ApiResponse{message=int64}		"new profile ID"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"json input error"
// @Failure		500			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/releaseProfile [post]
// @Security		ApiKeyAuth
func sportarrAddReleaseProfile(req *http.Request) (int, any) {
	var profile sonarr.ReleaseProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	// Get the profiles from sportarr.
	id, err := getSportarr(req).AddReleaseProfileContext(req.Context(), &profile)
	if err != nil {
		return apiError(http.StatusInternalServerError, "adding profile", err)
	}

	return http.StatusOK, id
}

// @Description	Updates a Sportarr Release Profile.
// @Summary		Update Sportarr Release Profile
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64								true	"instance ID"
// @Param			profileID	path		int64								true	"profile ID to update"
// @Param			PUT			body		sonarr.ReleaseProfile				true	"updated item content"
// @Success		200			{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"json input error"
// @Failure		422			{object}	apps.ApiResponse{message=string}	"no profile ID"
// @Failure		500			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/releaseProfile/{profileID} [put]
// @Security		ApiKeyAuth
func sportarrUpdateReleaseProfile(req *http.Request) (int, any) {
	var profile sonarr.ReleaseProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	profile.ID, _ = strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profile.ID == 0 {
		return http.StatusUnprocessableEntity, ErrNonZeroID
	}

	// Get the profiles from sportarr.
	_, err = getSportarr(req).UpdateReleaseProfileContext(req.Context(), &profile)
	if err != nil {
		return apiError(http.StatusInternalServerError, "updating profile", err)
	}

	return http.StatusOK, "OK"
}

// @Description	Removes a Sportarr Release Profile.
// @Summary		Remove Sportarr Release Profile
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64								true	"instance ID"
// @Param			profileID	path		int64								true	"profile ID to update"
// @Success		200			{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"no profile ID"
// @Failure		500			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/releaseProfile/{profileID} [delete]
// @Security		ApiKeyAuth
func sportarrDeleteReleaseProfile(req *http.Request) (int, any) {
	profileID, _ := strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profileID == 0 {
		return http.StatusBadRequest, ErrNonZeroID
	}

	// Delete the profile from sportarr.
	err := getSportarr(req).DeleteReleaseProfileContext(req.Context(), profileID)
	if err != nil {
		return apiError(http.StatusInternalServerError, "deleting profile", err)
	}

	return http.StatusOK, "OK"
}

// @Description	Removes all Sportarr Release Profiles.
// @Summary		Remove Sportarr Release Profiles
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64											true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=apps.deleteResponse}	"delete status"
// @Failure		500			{object}	apps.ApiResponse{message=string}				"instance error getting profiles"
// @Failure		404			{object}	string											"bad token or api key"
// @Router			/sportarr/{instance}/releaseProfiles/all [delete]
// @Security		ApiKeyAuth
func sportarrDeleteAllReleaseProfiles(req *http.Request) (int, any) {
	profiles, err := getSportarr(req).GetReleaseProfilesContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting profiles", err)
	}

	var (
		deleted int
		errs    []string
	)

	for _, profile := range profiles {
		// Delete the profile from sportarr.
		err := getSportarr(req).DeleteReleaseProfileContext(req.Context(), profile.ID)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		deleted++
	}

	return http.StatusOK, &deleteResponse{
		Found:   len(profiles),
		Deleted: deleted,
		Errors:  errs,
	}
}

// @Description	Returns all Sportarr Root Folders paths and free space.
// @Summary		Retrieve Sportarr Root Folders
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64										true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=map[string]int64}	"map of path->space free"
// @Failure		500			{object}	apps.ApiResponse{message=string}			"instance error"
// @Failure		404			{object}	string										"bad token or api key"
// @Router			/sportarr/{instance}/rootFolder [get]
// @Security		ApiKeyAuth
func sportarrRootFolders(req *http.Request) (int, any) {
	// Get folder list from Sportarr.
	folders, err := getSportarr(req).GetRootFoldersContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting folders", err)
	}

	// Format folder list into a nice path=>freesSpace map.
	p := make(map[string]int64)
	for i := range folders {
		p[folders[i].Path] = folders[i].FreeSpace
	}

	return http.StatusOK, p
}

// @Description	Returns Sportarr series naming conventions.
// @Summary		Retrieve Sportarr Series Naming
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64									true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=sonarr.Naming}	"naming conventions"
// @Failure		500			{object}	apps.ApiResponse{message=string}		"instance error"
// @Failure		404			{object}	string									"bad token or api key"
// @Router			/sportarr/{instance}/naming [get]
// @Security		ApiKeyAuth
func sportarrGetNaming(req *http.Request) (int, any) {
	naming, err := getSportarr(req).GetNamingContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting naming", err)
	}

	return http.StatusOK, naming
}

// @Description	Updates the Sportarr series naming conventions.
// @Summary		Update Sportarr Series Naming
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			PUT	body		sonarr.Naming						true	"naming conventions"
// @Success		200	{object}	apps.ApiResponse{message=int64}		"naming ID"
// @Failure		400	{object}	apps.ApiResponse{message=string}	"bad json input"
// @Failure		500	{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404	{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/naming [put]
// @Security		ApiKeyAuth
func sportarrUpdateNaming(req *http.Request) (int, any) {
	var naming sonarr.Naming

	err := json.NewDecoder(req.Body).Decode(&naming)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	output, err := getSportarr(req).UpdateNamingContext(req.Context(), &naming)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating naming", err)
	}

	return http.StatusOK, output.ID
}

type SportarrSeriesSearchItem struct {
	ID                int64     `json:"id"`
	Title             string    `json:"title"`
	First             time.Time `json:"first"`
	Next              time.Time `json:"next"`
	Previous          time.Time `json:"prev"`
	Added             time.Time `json:"added"`
	Status            string    `json:"status"`
	Path              string    `json:"path"`
	TvDBID            int64     `json:"tvdbId"`
	Monitored         bool      `json:"monitored"`
	QualityProfileID  int64     `json:"qualityId"`
	SeasonFolder      bool      `json:"seasonFolder"`
	SeriesType        string    `json:"seriesType"`
	LanguageProfileID int64     `json:"languageProfileId"`
	Exists            bool      `json:"exists"`
	Year              int       `json:"year"`
	Score             int       `json:"score"`
}

func sportarrConvertToSeriesSearchOutput(item *sonarr.Series, score int) SportarrSeriesSearchItem {
	return SportarrSeriesSearchItem{
		ID:                item.ID,
		Title:             item.Title,
		First:             item.FirstAired,
		Next:              item.NextAiring,
		Previous:          item.PreviousAiring,
		Added:             item.Added,
		Status:            item.Status,
		Path:              item.Path,
		TvDBID:            item.TvdbID,
		Monitored:         item.Monitored,
		QualityProfileID:  item.QualityProfileID,
		SeasonFolder:      item.SeasonFolder,
		SeriesType:        item.SeriesType,
		LanguageProfileID: item.LanguageProfileID,
		Exists:            item.Statistics != nil && item.Statistics.SizeOnDisk > 0,
		Year:              item.Year,
		Score:             score,
	}
}

// @Description	Searches all Sportarr Series Titles for the search term provided. Returns a minimal amount of data for each found item.
// @Summary		Search for Sportarr Series
// @Tags			Sportarr
// @Produce		json
// @Param			query		path		string												true	"title search string"
// @Param			instance	path		int64												true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=[]apps.SportarrSeriesSearchItem}	"minimal series data"
// @Failure		503			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/search/{query} [get]
// @Security		ApiKeyAuth
//
//nolint:lll
func sportarrSearchSeries(req *http.Request) (int, any) {
	// Get all series
	series, err := getSportarr(req).GetAllSeriesContext(req.Context())
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "getting series", err)
	}

	return http.StatusOK, sportarrSeriesSearch(mnd.GetID(req.Context()), series, mux.Vars(req)["query"])
}

func sportarrSeriesSearch(reqID string, series []*sonarr.Series, query string) []SportarrSeriesSearchItem {
	const (
		minLength  = 2 // Too short to search.
		maxResults = 10
		minScore   = 56
	)

	cleanedQuery := wuzzy.Cleanse(query, false)
	if len(cleanedQuery) < minLength {
		return []SportarrSeriesSearchItem{}
	}

	titles, resp := sportarrBuildTitlesList(series)
	mnd.Log.Printf(reqID, "[sportarr search] Found %d Sportarr titles from %d series for query %q (cleaned %q)",
		len(titles), len(series), query, cleanedQuery)

	// Find fuzzy matches.
	matches, err := wuzzy.Extract(query, titles, -1, sportarrMatcher, minScore)
	if err != nil {
		mnd.Log.Errorf(reqID, "[sportarr search] Finding fuzzy matches: %s", err)
	}

	have := map[int]bool{}
	// Now go back through the matches, and find the series that matches by name.
	for _, match := range matches {
		if len(resp) >= maxResults {
			break
		}

		for _, idx := range sportarrSeriesMatches(match, series) {
			if !have[idx] {
				mnd.Log.Printf(reqID, "[sportarr search] Fuzzy match (score: %d): %q (matched: %q)",
					match.Score, series[idx].Title, match.Match)
				resp = append(resp, sportarrConvertToSeriesSearchOutput(series[idx], match.Score))
				have[idx] = true
			}
		}
	}

	return resp
}

const (
	sportarrSuffixScore = 81
	sportarrPrefixScore = 91
	sportarrExactScore  = 100
)

func sportarrMatcher(query, title string) int {
	trimmedQuery := wuzzy.Cleanse(query, false)
	trimmedTitle := wuzzy.Cleanse(title, false)
	wuzz := wuzzy.Ratio(trimmedQuery, trimmedTitle)
	cleanedQuery := strings.TrimSuffix(strings.TrimPrefix(trimmedQuery, "the "), "s")
	cleanedTitle := strings.TrimPrefix(trimmedTitle, "the ")

	if cleanedTitle == query || cleanedTitle == cleanedQuery+"s" || cleanedTitle == cleanedQuery {
		return max(wuzz, sportarrExactScore)
	}

	if strings.HasPrefix(cleanedTitle, cleanedQuery) {
		return max(wuzz, sportarrPrefixScore)
	}

	if strings.HasSuffix(cleanedTitle, cleanedQuery) ||
		strings.HasSuffix(cleanedTitle, cleanedQuery+"s") ||
		strings.Contains(cleanedTitle, " "+cleanedQuery) {
		return max(wuzz, sportarrSuffixScore)
	}

	return wuzz
}

func sportarrBuildTitlesList(series []*sonarr.Series) ([]string, []SportarrSeriesSearchItem) {
	titles := make([]string, 0)
	// Build the titles list.
	for idx := range series {
		titles = append(titles, series[idx].Title)
		for _, alt := range series[idx].AlternateTitles {
			titles = append(titles, alt.Title)
		}
	}

	return titles, make([]SportarrSeriesSearchItem, 0)
}

func sportarrSeriesMatches(wuzz *wuzzy.MatchPair, series []*sonarr.Series) []int {
	matches := []int{}
	for idx, show := range series {
		if wuzz.Match == show.Title {
			matches = append(matches, idx)
		}

		for _, alt := range show.AlternateTitles {
			if wuzz.Match == alt.Title {
				matches = append(matches, idx)
			}
		}
	}

	return matches
}

// @Description	Returns all Sportarr Tags.
// @Summary		Retrieve Sportarr Tags
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64									true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=[]starr.Tag}	"tags"
// @Failure		503			{object}	apps.ApiResponse{message=string}		"instance error"
// @Failure		404			{object}	string									"bad token or api key"
// @Router			/sportarr/{instance}/tag [get]
// @Security		ApiKeyAuth
func sportarrGetTags(req *http.Request) (int, any) {
	tags, err := getSportarr(req).GetTagsContext(req.Context())
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "getting tags", err)
	}

	return http.StatusOK, tags
}

// @Description	Updates the label for an existing Sportarr tag.
// @Summary		Update Sportarr Tag Label
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64								true	"instance ID"
// @Param			tagID		path		int64								true	"tag ID to update"
// @Param			label		path		string								true	"new label"
// @Success		200			{object}	apps.ApiResponse{message=int64}		"tag ID"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/tag/{tagID}/{label} [put]
// @Security		ApiKeyAuth
func sportarrUpdateTag(req *http.Request) (int, any) {
	id, _ := strconv.Atoi(mux.Vars(req)["tid"])

	tag, err := getSportarr(req).UpdateTagContext(req.Context(), &starr.Tag{ID: id, Label: mux.Vars(req)["label"]})
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating tag", err)
	}

	return http.StatusOK, tag.ID
}

// @Description	Create a brand new tag in Sportarr.
// @Summary		Create Sportarr Tag
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64								true	"instance ID"
// @Param			label		path		string								true	"tag label"
// @Success		200			{object}	apps.ApiResponse{message=int64}		"tag ID"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/tag/{label} [put]
// @Security		ApiKeyAuth
func sportarrSetTag(req *http.Request) (int, any) {
	tag, err := getSportarr(req).AddTagContext(req.Context(), &starr.Tag{Label: mux.Vars(req)["label"]})
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "setting tag", err)
	}

	return http.StatusOK, tag.ID
}

// @Description	Updates a series in Sportarr.
// @Summary		Update Sportarr Series
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64								true	"instance ID"
// @Param			moveFiles	query		int64								true	"move files? true/false"
// @Param			PUT			body		sonarr.Series						true	"series content"
// @Success		200			{object}	apps.ApiResponse{message=string}	"OK"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/update [put]
// @Security		ApiKeyAuth
func sportarrUpdateSeries(req *http.Request) (int, any) {
	var series sonarr.AddSeriesInput

	err := json.NewDecoder(req.Body).Decode(&series)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	moveFiles := req.URL.Query().Get("moveFiles") == mnd.True

	_, err = getSportarr(req).UpdateSeriesContext(req.Context(), &series, moveFiles)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating series", err)
	}

	return http.StatusOK, "sportarr seems to have worked"
}

// @Description	Season Pass allows you to mass-edit items in Sportarr.
// @Summary		Publish Sportarr Season Pass
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64								true	"instance ID"
// @Param			POST		body		sonarr.SeasonPass					true	"Season pass content"
// @Success		200			{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"invalid json provided"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/seasonPass [post]
// @Security		ApiKeyAuth
func sportarrSeasonPass(req *http.Request) (int, any) {
	var seasonPass sonarr.SeasonPass

	err := json.NewDecoder(req.Body).Decode(&seasonPass)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	err = getSportarr(req).UpdateSeasonPassContext(req.Context(), &seasonPass)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating seasonPass", err)
	}

	return http.StatusOK, "ok"
}

// @Description	Creates a new Custom Format in Sportarr.
// @Summary		Create Sportarr Custom Format
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64												true	"instance ID"
// @Param			POST		body		sonarr.CustomFormatInput							true	"New Custom Format content"
// @Success		200			{object}	apps.ApiResponse{message=sonarr.CustomFormatOutput}	"custom format"
// @Failure		400			{object}	apps.ApiResponse{message=string}					"invalid json provided"
// @Failure		500			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/customformats [post]
// @Security		ApiKeyAuth
func sportarrAddCustomFormat(req *http.Request) (int, any) {
	var cusform sonarr.CustomFormatInput

	err := json.NewDecoder(req.Body).Decode(&cusform)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	resp, err := getSportarr(req).AddCustomFormatContext(req.Context(), &cusform)
	if err != nil {
		return apiError(http.StatusInternalServerError, "adding custom format", err)
	}

	return http.StatusOK, resp
}

// @Description	Returns all Custom Format from Sportarr.
// @Summary		Get Sportarr Custom Formats
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64													true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=[]sonarr.CustomFormatOutput}	"custom formats"
// @Failure		500			{object}	apps.ApiResponse{message=string}						"instance error"
// @Failure		404			{object}	string													"bad token or api key"
// @Router			/sportarr/{instance}/customformats [get]
// @Security		ApiKeyAuth
func sportarrGetCustomFormats(req *http.Request) (int, any) {
	cusform, err := getSportarr(req).GetCustomFormatsContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting custom formats", err)
	}

	return http.StatusOK, cusform
}

// @Description	Updates a Custom Format in Sportarr.
// @Summary		Update Sportarr Custom Format
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64												true	"instance ID"
// @Param			formatID	path		int64												true	"Custom Format ID"
// @Param			PUT			body		sonarr.CustomFormatInput							true	"Updated Custom Format content"
// @Success		200			{object}	apps.ApiResponse{message=sonarr.CustomFormatOutput}	"custom format"
// @Failure		400			{object}	apps.ApiResponse{message=string}					"invalid json provided"
// @Failure		500			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/customformats/{formatID} [put]
// @Security		ApiKeyAuth
func sportarrUpdateCustomFormat(req *http.Request) (int, any) {
	var cusform sonarr.CustomFormatInput
	if err := json.NewDecoder(req.Body).Decode(&cusform); err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	output, err := getSportarr(req).UpdateCustomFormatContext(req.Context(), &cusform)
	if err != nil {
		return apiError(http.StatusInternalServerError, "updating custom format", err)
	}

	return http.StatusOK, output
}

// @Description	Delete a Custom Format from Sportarr.
// @Summary		Delete Sportarr Custom Format
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64								true	"instance ID"
// @Param			formatID	path		int64								true	"Custom Format ID"
// @Success		200			{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		500			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/customformats/{formatID} [delete]
// @Security		ApiKeyAuth
func sportarrDeleteCustomFormat(req *http.Request) (int, any) {
	cfID, _ := strconv.ParseInt(mux.Vars(req)["cfid"], mnd.Base10, mnd.Bits64)

	err := getSportarr(req).DeleteCustomFormatContext(req.Context(), cfID)
	if err != nil {
		return apiError(http.StatusInternalServerError, "deleting custom format", err)
	}

	return http.StatusOK, "OK"
}

// @Description	Delete all Custom Formats from Sportarr.
// @Summary		Delete all Sportarr Custom Formats
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64											true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=apps.deleteResponse}	"item delete counters"
// @Failure		500			{object}	apps.ApiResponse{message=string}				"instance error"
// @Failure		404			{object}	string											"bad token or api key"
// @Router			/sportarr/{instance}/customformats/all [delete]
// @Security		ApiKeyAuth
func sportarrDeleteAllCustomFormats(req *http.Request) (int, any) {
	formats, err := getSportarr(req).GetCustomFormatsContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting custom formats", err)
	}

	var (
		deleted int
		errs    []string
	)

	for _, format := range formats {
		err := getSportarr(req).DeleteCustomFormatContext(req.Context(), format.ID)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		deleted++
	}

	return http.StatusOK, &deleteResponse{
		Found:   len(formats),
		Deleted: deleted,
		Errors:  errs,
	}
}

// @Description	Returns all Import Lists from Sportarr.
// @Summary		Get Sportarr Import Lists
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64												true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=[]sonarr.ImportListOutput}	"list of import lists"
// @Failure		500			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/importlist [get]
// @Security		ApiKeyAuth
func sportarrGetImportLists(req *http.Request) (int, any) {
	ilist, err := getSportarr(req).GetImportListsContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting import lists", err)
	}

	return http.StatusOK, ilist
}

// @Description	Updates an Import List in Sportarr.
// @Summary		Update Sportarr Import List
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64												true	"instance ID"
// @Param			listID		path		int64												true	"Import List ID"
// @Param			PUT			body		sonarr.ImportListInput								true	"Updated Import List Content"
// @Success		200			{object}	apps.ApiResponse{message=sonarr.ImportListOutput}	"import list returns"
// @Failure		400			{object}	apps.ApiResponse{message=string}					"invalid json provided"
// @Failure		500			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/importlist/{listID} [put]
// @Security		ApiKeyAuth
func sportarrUpdateImportList(req *http.Request) (int, any) {
	var ilist sonarr.ImportListInput
	if err := json.NewDecoder(req.Body).Decode(&ilist); err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	ilist.ID, _ = strconv.ParseInt(mux.Vars(req)["ilid"], mnd.Base10, mnd.Bits64)

	output, err := getSportarr(req).UpdateImportListContext(req.Context(), &ilist, false)
	if err != nil {
		return apiError(http.StatusInternalServerError, "updating import list", err)
	}

	return http.StatusOK, output
}

// @Description	Creates a new Import List in Sportarr.
// @Summary		Create Sportarr Import List
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64												true	"instance ID"
// @Param			POST		body		sonarr.ImportListInput								true	"New Import List"
// @Success		200			{object}	apps.ApiResponse{message=sonarr.ImportListOutput}	"import list returns"
// @Failure		400			{object}	apps.ApiResponse{message=string}					"invalid json provided"
// @Failure		500			{object}	apps.ApiResponse{message=string}					"instance error"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/sportarr/{instance}/importlist [post]
// @Security		ApiKeyAuth
func sportarrAddImportList(req *http.Request) (int, any) {
	var ilist sonarr.ImportListInput
	if err := json.NewDecoder(req.Body).Decode(&ilist); err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	output, err := getSportarr(req).AddImportListContext(req.Context(), &ilist)
	if err != nil {
		return apiError(http.StatusInternalServerError, "creating import list", err)
	}

	return http.StatusOK, output
}

// @Description	Returns all Quality Definitions from Sportarr.
// @Summary		Get Sportarr Quality Definitions
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64													true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=[]sonarr.QualityDefinition}	"quality definitions list"
// @Failure		500			{object}	apps.ApiResponse{message=string}						"instance error"
// @Failure		404			{object}	string													"bad token or api key"
// @Router			/sportarr/{instance}/qualitydefinitions [get]
// @Security		ApiKeyAuth
func sportarrGetQualityDefinitions(req *http.Request) (int, any) {
	output, err := getSportarr(req).GetQualityDefinitionsContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting quality definitions", err)
	}

	return http.StatusOK, output
}

// @Description	Updates all Quality Definitions in Sportarr.
// @Summary		Update Sportarr Quality Definitions
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64													true	"instance ID"
// @Param			PUT			body		[]sonarr.QualityDefinition								true	"Updated Import Listcontent"
// @Success		200			{object}	apps.ApiResponse{message=[]sonarr.QualityDefinition}	"quality definitions return"
// @Failure		400			{object}	apps.ApiResponse{message=string}						"invalid json provided"
// @Failure		500			{object}	apps.ApiResponse{message=string}						"instance error"
// @Failure		404			{object}	string													"bad token or api key"
// @Router			/sportarr/{instance}/qualitydefinition [put]
// @Security		ApiKeyAuth
func sportarrUpdateQualityDefinition(req *http.Request) (int, any) {
	var input []*sonarr.QualityDefinition
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	output, err := getSportarr(req).UpdateQualityDefinitionsContext(req.Context(), input)
	if err != nil {
		return apiError(http.StatusInternalServerError, "updating quality definition", err)
	}

	return http.StatusOK, output
}

// @Description	Returns Sportarr Notifications with a name that matches 'notifiar'.
// @Summary		Retrieve Sportarr Notifications
// @Tags			Sportarr
// @Produce		json
// @Param			instance	path		int64													true	"instance ID"
// @Success		200			{object}	apps.ApiResponse{message=[]sonarr.NotificationOutput}	"notifications"
// @Failure		503			{object}	apps.ApiResponse{message=string}						"instance error"
// @Failure		404			{object}	string													"bad token or api key"
// @Router			/sportarr/{instance}/notification [get]
// @Security		ApiKeyAuth
func sportarrGetNotifications(req *http.Request) (int, any) {
	notifs, err := getSportarr(req).GetNotificationsContext(req.Context())
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "getting notifications", err)
	}

	output := []*sonarr.NotificationOutput{}

	for _, notif := range notifs {
		if strings.Contains(strings.ToLower(notif.Name), "notifiar") {
			output = append(output, notif)
		}
	}

	return http.StatusOK, output
}

// @Description	Updates a Notification in Sportarr.
// @Summary		Update Sportarr Notification
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64								true	"instance ID"
// @Param			PUT			body		sonarr.NotificationInput			true	"notification content"
// @Success		200			{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"bad json input"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/notification [put]
// @Security		ApiKeyAuth
func sportarrUpdateNotification(req *http.Request) (int, any) {
	var notif sonarr.NotificationInput

	err := json.NewDecoder(req.Body).Decode(&notif)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	_, err = getSportarr(req).UpdateNotificationContext(req.Context(), &notif)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating notification", err)
	}

	return http.StatusOK, mnd.Success
}

// @Description	Creates a new Sportarr Notification.
// @Summary		Add Sportarr Notification
// @Tags			Sportarr
// @Produce		json
// @Accept			json
// @Param			instance	path		int64								true	"instance ID"
// @Param			POST		body		sonarr.NotificationInput			true	"new item content"
// @Success		200			{object}	apps.ApiResponse{message=int64}		"new notification ID"
// @Failure		400			{object}	apps.ApiResponse{message=string}	"json input error"
// @Failure		503			{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404			{object}	string								"bad token or api key"
// @Router			/sportarr/{instance}/notification [post]
// @Security		ApiKeyAuth
func sportarrAddNotification(req *http.Request) (int, any) {
	var notif sonarr.NotificationInput

	err := json.NewDecoder(req.Body).Decode(&notif)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	id, err := getSportarr(req).AddNotificationContext(req.Context(), &notif)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "adding notification", err)
	}

	return http.StatusOK, id
}

// @Description	Delete items from the activity queue.
// @Summary		Delete Queue Items
// @Tags			Sportarr
// @Produce		json
// @Param			instance			path		int64								true	"instance ID"
// @Param			queueID				path		int64								true	"queue ID to delete"
// @Param			removeFromClient	query		bool								false	"remove download from download client?"
// @Param			blocklist			query		bool								false	"add item to blocklist?"
// @Param			skipRedownload		query		bool								false	"skip downloading this again?"
// @Param			changeCategory		query		bool								false	"tell download client to change categories?"
// @Success		200					{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		500					{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404					{object}	string								"bad token or api key"
// @Failure		423					{object}	string								"rate limit reached"
// @Router			/sportarr/{instance}/queue/{queueID} [delete]
// @Security		ApiKeyAuth
func sportarrDeleteQueue(req *http.Request) (int, any) {
	idString := mux.Vars(req)["queueID"]
	queueID, _ := strconv.ParseInt(idString, mnd.Base10, mnd.Bits64)
	removeFromClient := req.URL.Query().Get("removeFromClient") == mnd.True
	opts := &starr.QueueDeleteOpts{
		RemoveFromClient: &removeFromClient,
		BlockList:        req.URL.Query().Get("blocklist") == mnd.True,
		SkipRedownload:   req.URL.Query().Get("skipRedownload") == mnd.True,
		ChangeCategory:   req.URL.Query().Get("changeCategory") == mnd.True,
	}

	err := getSportarr(req).DeleteQueueContext(req.Context(), queueID, opts)
	if err != nil {
		return apiError(http.StatusInternalServerError, "deleting queue", err)
	}

	return http.StatusOK, mnd.Deleted + idString
}

// @Description	Delete episode files from Sportarr.
// @Summary		Remove Sportarr episode files
// @Tags			Sportarr
// @Produce		json
// @Param			instance		path		int64								true	"instance ID"
// @Param			episodeFileID	path		int64								true	"episode file ID to delete, not episode ID"
// @Success		200				{object}	apps.ApiResponse{message=string}	"ok"
// @Failure		500				{object}	apps.ApiResponse{message=string}	"instance error"
// @Failure		404				{object}	string								"bad token or api key"
// @Failure		423				{object}	string								"rate limit reached"
// @Router			/sportarr/{instance}/delete/{episodeFileID} [delete]
// @Security		ApiKeyAuth
func sportarrDeleteEpisode(req *http.Request) (int, any) {
	idString := mux.Vars(req)["episodeFileID"]
	episodeFileID, _ := strconv.ParseInt(idString, mnd.Base10, mnd.Bits64)

	if !getSportarr(req).StarrApp.DelOK() {
		return http.StatusLocked, ErrRateLimit
	}

	err := getSportarr(req).DeleteEpisodeFileContext(req.Context(), episodeFileID)
	if err != nil {
		return apiError(http.StatusInternalServerError, "deleting episode file", err)
	}

	return http.StatusOK, mnd.Deleted + idString
}
