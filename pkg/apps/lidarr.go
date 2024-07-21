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
	"golang.org/x/time/rate"
	"golift.io/starr"
	"golift.io/starr/debuglog"
	"golift.io/starr/lidarr"
)

/*
mbid - music brainz is the source for lidarr (todo)
*/

// lidarrHandlers is called once on startup to register the web API paths.
func (a *Apps) lidarrHandlers() {
	a.HandleAPIpath(starr.Lidarr, "/add", lidarrAddAlbum, "POST")
	a.HandleAPIpath(starr.Lidarr, "/artist/{artistid:[0-9]+}", lidarrGetArtist, "GET")
	a.HandleAPIpath(starr.Lidarr, "/check/{mbid:[-a-z0-9]+}", lidarrCheckAlbum, "GET")
	a.HandleAPIpath(starr.Lidarr, "/get/{albumid:[0-9]+}", lidarrGetAlbum, "GET")
	a.HandleAPIpath(starr.Lidarr, "/metadataProfiles", lidarrMetadata, "GET")
	a.HandleAPIpath(starr.Lidarr, "/naming", lidarrGetNaming, "GET")
	a.HandleAPIpath(starr.Lidarr, "/naming", lidarrUpdateNaming, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/customformats", lidarrGetCustomFormats, "GET")
	a.HandleAPIpath(starr.Lidarr, "/customformats", lidarrAddCustomFormat, "POST")
	a.HandleAPIpath(starr.Lidarr, "/customformats", lidarrUpdateCustomFormat, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/customformats/{cfid:[0-9]+}", lidarrUpdateCustomFormat, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/customformats/{cfid:[0-9]+}", lidarrDeleteCustomFormat, "DELETE")
	a.HandleAPIpath(starr.Lidarr, "/customformats/all", lidarrDeleteAllCustomFormats, "DELETE")
	a.HandleAPIpath(starr.Lidarr, "/qualitydefinition", lidarrUpdateQualityDefinition, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/qualityDefinitions", lidarrGetQualityDefinitions, "GET")
	a.HandleAPIpath(starr.Lidarr, "/qualityProfiles", lidarrQualityProfiles, "GET")
	a.HandleAPIpath(starr.Lidarr, "/qualityProfile", lidarrGetQualityProfile, "GET")
	a.HandleAPIpath(starr.Lidarr, "/qualityProfile", lidarrAddQualityProfile, "POST")
	a.HandleAPIpath(starr.Lidarr, "/qualityProfile/{profileID:[0-9]+}", lidarrUpdateQualityProfile, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/rootFolder", lidarrRootFolders, "GET")
	a.HandleAPIpath(starr.Lidarr, "/search/{query}", lidarrSearchAlbum, "GET")
	a.HandleAPIpath(starr.Lidarr, "/tag", lidarrGetTags, "GET")
	a.HandleAPIpath(starr.Lidarr, "/tag/{tid:[0-9]+}/{label}", lidarrUpdateTag, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/tag/{label}", lidarrSetTag, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/update", lidarrUpdateAlbum, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/updateartist", lidarrUpdateArtist, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/command/search/{albumid:[0-9]+}", lidarrTriggerSearchAlbum, "GET")
	a.HandleAPIpath(starr.Lidarr, "/notification", lidarrGetNotifications, "GET")
	a.HandleAPIpath(starr.Lidarr, "/notification", lidarrUpdateNotification, "PUT")
	a.HandleAPIpath(starr.Lidarr, "/notification", lidarrAddNotification, "POST")
	a.HandleAPIpath(starr.Lidarr, "/queue/{queueID}", lidarrDeleteQueue, "DELETE")
}

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	ExtraConfig
	*starr.Config
	*lidarr.Lidarr `json:"-" toml:"-" xml:"-"`
	errorf         func(string, ...interface{}) `json:"-" toml:"-" xml:"-"`
}

func getLidarr(r *http.Request) *LidarrConfig {
	return r.Context().Value(starr.Lidarr).(*LidarrConfig) //nolint:forcetypeassert
}

// Enabled returns true if the Lidarr instance is enabled and usable.
func (l *LidarrConfig) Enabled() bool {
	return l != nil && l.Config != nil && l.URL != "" && l.APIKey != "" && l.Timeout.Duration >= 0
}

func (a *Apps) setupLidarr() error {
	for idx, app := range a.Lidarr {
		if app.Config == nil || app.Config.URL == "" {
			return fmt.Errorf("%w: missing url: Lidarr config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Lidarr config %d", ErrInvalidApp, idx+1)
		}

		if a.Logger.DebugEnabled() {
			app.Config.Client = starr.ClientWithDebug(app.Timeout.Duration, app.ValidSSL, debuglog.Config{
				MaxBody: a.MaxBody,
				Debugf:  a.Debugf,
				Caller:  metricMakerCallback(string(starr.Lidarr)),
				Redact:  []string{app.APIKey, app.Password, app.HTTPPass},
			})
		} else {
			app.Config.Client = starr.Client(app.Timeout.Duration, app.ValidSSL)
			app.Config.Client.Transport = NewMetricsRoundTripper(starr.Lidarr.String(), nil)
		}

		app.errorf = a.Errorf
		app.URL = strings.TrimRight(app.URL, "/")
		app.Lidarr = lidarr.New(app.Config)

		if app.Deletes > 0 {
			app.delLimit = rate.NewLimiter(rate.Every(1*time.Hour/time.Duration(app.Deletes)), app.Deletes)
		}
	}

	return nil
}

// @Description  Adds a new Album to Lidarr.
// @Summary      Add Lidarr Album
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        POST body lidarr.AddAlbumInput true "new item content"
// @Accept       json
// @Success      201  {object} apps.Respond.apiResponse{message=lidarr.Album} "created"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json payload"
// @Failure      409  {object} apps.Respond.apiResponse{message=string} "item already exists"
// @Failure      422  {object} apps.Respond.apiResponse{message=string} "no item ID provided"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error during check"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error during add"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/add [post]
// @Security     ApiKeyAuth
func lidarrAddAlbum(req *http.Request) (int, interface{}) {
	var payload lidarr.AddAlbumInput

	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	} else if payload.ForeignAlbumID == "" {
		return apiError(http.StatusUnprocessableEntity, "0", ErrNoMBID)
	}

	// Check for existing album.
	m, err := getLidarr(req).GetAlbumContext(req.Context(), payload.ForeignAlbumID)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "checking album", err)
	} else if len(m) > 0 {
		return http.StatusConflict, lidarrData(m[0])
	}

	album, err := getLidarr(req).AddAlbumContext(req.Context(), &payload)
	if err != nil {
		return apiError(http.StatusInternalServerError, "adding album", err)
	}

	return http.StatusCreated, album
}

// @Description  Fetches an Artist from Lidarr.
// @Summary      Get Lidarr Artist
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        artistID  path   int64  true  "artist ID"
// @Success      200  {object} apps.Respond.apiResponse{message=lidarr.Artist} "ok"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/artist/{artistID} [get]
// @Security     ApiKeyAuth
func lidarrGetArtist(req *http.Request) (int, interface{}) {
	artistID, _ := strconv.ParseInt(mux.Vars(req)["artistid"], mnd.Base10, mnd.Bits64)

	artist, err := getLidarr(req).GetArtistByIDContext(req.Context(), artistID)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "checking artist", err)
	}

	return http.StatusOK, artist
}

func lidarrData(album *lidarr.Album) map[string]interface{} {
	hasFile := false
	if album.Statistics != nil {
		hasFile = album.Statistics.SizeOnDisk > 0
	}

	return map[string]interface{}{
		"id":        album.ID,
		"hasFile":   hasFile,
		"monitored": album.Monitored,
		"tags":      album.Artist.Tags,
	}
}

// @Description  Checks if an album already exists in Lidarr.
// @Summary      Check for Lidarr Album existence
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        mbid  path   int64  true  "music brains ID"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "not found"
// @Failure      409  {object} apps.Respond.apiResponse{message=string} "already exists"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/check/{mbid} [get]
// @Security     ApiKeyAuth
func lidarrCheckAlbum(req *http.Request) (int, interface{}) {
	id := mux.Vars(req)["mbid"]

	m, err := getLidarr(req).GetAlbumContext(req.Context(), id)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "checking album", err)
	} else if len(m) > 0 {
		return http.StatusConflict, lidarrData(m[0])
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

// @Description  Fetches an Album from Lidarr.
// @Summary      Get Lidarr Album
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        albumID  path   int64  true  "album ID"
// @Success      200  {object} apps.Respond.apiResponse{message=lidarr.Album} "ok"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/get/{albumID} [get]
// @Security     ApiKeyAuth
func lidarrGetAlbum(req *http.Request) (int, interface{}) {
	albumID, _ := strconv.ParseInt(mux.Vars(req)["albumid"], mnd.Base10, mnd.Bits64)

	album, err := getLidarr(req).GetAlbumByIDContext(req.Context(), albumID)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "checking album", err)
	}

	return http.StatusOK, album
}

// @Description  Returns the search status of an album ID.
// @Summary      Search Lidarr Album ID
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        albumID  path   int64  true  "album ID"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "status"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/command/search/{albumID} [get]
// @Security     ApiKeyAuth
func lidarrTriggerSearchAlbum(req *http.Request) (int, interface{}) {
	albumID, _ := strconv.ParseInt(mux.Vars(req)["albumid"], mnd.Base10, mnd.Bits64)

	output, err := getLidarr(req).SendCommandContext(req.Context(), &lidarr.CommandRequest{
		Name:     "AlbumSearch",
		AlbumIDs: []int64{albumID},
	})
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "triggering album search", err)
	}

	return http.StatusOK, output.Status
}

// @Description  Fetches all Metadata Profiles from Lidarr.
// @Summary      Get Lidarr Metadata Profiles
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=map[int64]string} "map of ID to name"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/metadataProfiles [get]
// @Security     ApiKeyAuth
func lidarrMetadata(req *http.Request) (int, interface{}) {
	profiles, err := getLidarr(req).GetMetadataProfilesContext(req.Context())
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

// @Description  Returns Lidarr track naming conventions.
// @Summary      Retrieve Lidarr Track Naming
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=lidarr.Naming} "naming conventions"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/naming [get]
// @Security     ApiKeyAuth
func lidarrGetNaming(req *http.Request) (int, interface{}) {
	naming, err := getLidarr(req).GetNamingContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting naming", err)
	}

	return http.StatusOK, naming
}

// @Description  Updates the Lidarr track naming conventions.
// @Summary      Update Lidarr Track Naming
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        PUT body lidarr.Naming  true  "naming conventions"
// @Success      200  {object} apps.Respond.apiResponse{message=int64} "naming ID"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json input"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/naming [put]
// @Security     ApiKeyAuth
func lidarrUpdateNaming(req *http.Request) (int, interface{}) {
	var naming lidarr.Naming

	err := json.NewDecoder(req.Body).Decode(&naming)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	output, err := getLidarr(req).UpdateNamingContext(req.Context(), &naming)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating naming", err)
	}

	return http.StatusOK, output.ID
}

// @Description  Creates a new Custom Format in Lidarr.
// @Summary      Create Lidarr Custom Format
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        POST body lidarr.CustomFormatInput  true  "New Custom Format content"
// @Success      200  {object} apps.Respond.apiResponse{message=lidarr.CustomFormatOutput}  "custom format"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "invalid json provided"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/customformats [post]
// @Security     ApiKeyAuth
func lidarrAddCustomFormat(req *http.Request) (int, interface{}) {
	var cusform lidarr.CustomFormatInput

	err := json.NewDecoder(req.Body).Decode(&cusform)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	resp, err := getLidarr(req).AddCustomFormatContext(req.Context(), &cusform)
	if err != nil {
		return apiError(http.StatusInternalServerError, "adding custom format", err)
	}

	return http.StatusOK, resp
}

// @Description  Returns all Custom Formats Data from Lidarr.
// @Summary      Get Lidarr Custom Formats Data
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]lidarr.CustomFormatOutput}  "custom formats"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/customformats [get]
// @Security     ApiKeyAuth
func lidarrGetCustomFormats(req *http.Request) (int, interface{}) {
	cusform, err := getLidarr(req).GetCustomFormatsContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting custom formats", err)
	}

	return http.StatusOK, cusform
}

// @Description  Updates a Custom Format in Lidarr.
// @Summary      Update Lidarr Custom Format
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        PUT body lidarr.CustomFormatInput  true  "Updated Custom Format content"
// @Success      200  {object} apps.Respond.apiResponse{message=lidarr.CustomFormatOutput}  "custom format"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "invalid json provided"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/customformats/{formatID} [put]
// @Security     ApiKeyAuth
func lidarrUpdateCustomFormat(req *http.Request) (int, interface{}) {
	var cusform lidarr.CustomFormatInput
	if err := json.NewDecoder(req.Body).Decode(&cusform); err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	output, err := getLidarr(req).UpdateCustomFormatContext(req.Context(), &cusform)
	if err != nil {
		return apiError(http.StatusInternalServerError, "updating custom format", err)
	}

	return http.StatusOK, output
}

// @Description  Delete a Custom Format from Lidarr.
// @Summary      Delete Lidarr Custom Format
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        formatID  path   int64  true  "Custom Format ID"
// @Success      200  {object} apps.Respond.apiResponse{message=string}  "ok"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/customformats/{formatID} [delete]
// @Security     ApiKeyAuth
func lidarrDeleteCustomFormat(req *http.Request) (int, interface{}) {
	cfID, _ := strconv.ParseInt(mux.Vars(req)["cfid"], mnd.Base10, mnd.Bits64)

	err := getLidarr(req).DeleteCustomFormatContext(req.Context(), cfID)
	if err != nil {
		return apiError(http.StatusInternalServerError, "deleting custom format", err)
	}

	return http.StatusOK, "OK"
}

// @Description  Delete all Custom Formats from Lidarr.
// @Summary      Delete all Lidarr Custom Formats
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=apps.deleteResponse}  "item delete counters"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/customformats/all [delete]
// @Security     ApiKeyAuth
func lidarrDeleteAllCustomFormats(req *http.Request) (int, interface{}) {
	formats, err := getLidarr(req).GetCustomFormatsContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting custom formats", err)
	}

	var (
		deleted int
		errs    []string
	)

	for _, format := range formats {
		err := getLidarr(req).DeleteCustomFormatContext(req.Context(), format.ID)
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

// @Description  Updates all Quality Definitions in Lidarr.
// @Summary      Update Lidarr Quality Definitions
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        PUT body []lidarr.QualityDefinition  true  "Updated quality definitions"
// @Success      200  {object} apps.Respond.apiResponse{message=[]lidarr.QualityDefinition}  "quality definitions return"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "invalid json provided"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/qualitydefinition [put]
// @Security     ApiKeyAuth
//
//nolint:lll
func lidarrUpdateQualityDefinition(req *http.Request) (int, interface{}) {
	var input []*lidarr.QualityDefinition
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	output, err := getLidarr(req).UpdateQualityDefinitionsContext(req.Context(), input)
	if err != nil {
		return apiError(http.StatusInternalServerError, "updating quality definition", err)
	}

	return http.StatusOK, output
}

// @Description  Fetches all Quality Definitions from Lidarr.
// @Summary      Get Lidarr Quality Definitions
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=map[int64]string} "map of ID to name"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/qualityDefinitions [get]
// @Security     ApiKeyAuth
func lidarrGetQualityDefinitions(req *http.Request) (int, interface{}) {
	// Get the profiles from lidarr.
	definitions, err := getLidarr(req).GetQualityDefinitionsContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting profiles", err)
	}

	// Format definitions ID=>Title into a nice map.
	p := make(map[int64]string)
	for i := range definitions {
		p[definitions[i].ID] = definitions[i].Title
	}

	return http.StatusOK, p
}

// @Description  Fetches all Quality Profiles from Lidarr.
// @Summary      Get Lidarr Quality Profiles
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=map[int64]string} "map of ID to name"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/qualityProfiles [get]
// @Security     ApiKeyAuth
func lidarrQualityProfiles(req *http.Request) (int, interface{}) {
	// Get the profiles from lidarr.
	profiles, err := getLidarr(req).GetQualityProfilesContext(req.Context())
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

// @Description  Fetches all Quality Profiles Data from Lidarr.
// @Summary      Get Lidarr Quality Profile Data
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]lidarr.QualityProfile} "all profiles"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/qualityProfile [get]
// @Security     ApiKeyAuth
func lidarrGetQualityProfile(req *http.Request) (int, interface{}) {
	// Get the profiles from lidarr.
	profiles, err := getLidarr(req).GetQualityProfilesContext(req.Context())
	if err != nil {
		return apiError(http.StatusInternalServerError, "getting profiles", err)
	}

	return http.StatusOK, profiles
}

// @Description  Creates a new Lidarr Quality Profile.
// @Summary      Add Lidarr Quality Profile
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        POST body lidarr.QualityProfile true "new item content"
// @Success      200  {object} apps.Respond.apiResponse{message=int64} "new profile ID"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "json input error"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/qualityProfile [post]
// @Security     ApiKeyAuth
func lidarrAddQualityProfile(req *http.Request) (int, interface{}) {
	var profile lidarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	// Get the profiles from lidarr.
	id, err := getLidarr(req).AddQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return apiError(http.StatusInternalServerError, "adding profile", err)
	}

	return http.StatusOK, id
}

// @Description  Updates a Lidarr Quality Profile.
// @Summary      Update Lidarr Quality Profile
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        profileID  path   int64  true  "profile ID to update"
// @Param        PUT body lidarr.QualityProfile true "updated item content"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "ok"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "json input error"
// @Failure      422  {object} apps.Respond.apiResponse{message=string} "no profile ID"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/qualityProfile/{profileID} [put]
// @Security     ApiKeyAuth
func lidarrUpdateQualityProfile(req *http.Request) (int, interface{}) {
	var profile lidarr.QualityProfile

	// Extract payload and check for TMDB ID.
	err := json.NewDecoder(req.Body).Decode(&profile)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	profile.ID, _ = strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profile.ID == 0 {
		return http.StatusUnprocessableEntity, ErrNonZeroID
	}

	// Get the profiles from lidarr.
	_, err = getLidarr(req).UpdateQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return apiError(http.StatusInternalServerError, "updating profile", err)
	}

	return http.StatusOK, "OK"
}

// @Description  Returns all Lidarr Root Folders paths and free space.
// @Summary      Retrieve Lidarr Root Folders
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=map[string]int64} "map of path->space free"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/rootFolder [get]
// @Security     ApiKeyAuth
func lidarrRootFolders(req *http.Request) (int, interface{}) {
	// Get folder list from Lidarr.
	folders, err := getLidarr(req).GetRootFoldersContext(req.Context())
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

// @Description  Searches all Album Titles for the search term provided.
// @Summary      Search for Lidarr Albums
// @Tags         Lidarr
// @Produce      json
// @Param        query     path   string  true  "title search string"
// @Param        instance  path   int64   true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]apps.lidarrSearchAlbum.albumData}  "minimal album data"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/search/{query} [get]
// @Security     ApiKeyAuth
//
//nolint:lll
func lidarrSearchAlbum(req *http.Request) (int, interface{}) {
	albums, err := getLidarr(req).GetAlbumContext(req.Context(), "")
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "getting albums", err)
	}

	type albumData struct {
		ID         int64     `json:"id"`
		MBID       string    `json:"mbid"`
		MetadataID int64     `json:"metadataId"`
		QualityID  int64     `json:"qualityId"`
		Title      string    `json:"title"`
		Release    time.Time `json:"release"`
		ArtistID   int64     `json:"artistId"`
		ArtistName string    `json:"artistName"`
		ProfileID  int64     `json:"profileId"`
		Overview   string    `json:"overview"`
		Ratings    float64   `json:"ratings"`
		Type       string    `json:"type"`
		Exists     bool      `json:"exists"`
		Files      int64     `json:"files"`
	}

	query := strings.TrimSpace(mux.Vars(req)["query"]) // in
	output := make([]*albumData, 0)                    // out

	for _, album := range albums {
		if albumSearch(query, album.Title, album.Releases) {
			output = append(output, &albumData{
				ID:         album.ID,
				MBID:       album.ForeignAlbumID,
				MetadataID: album.Artist.MetadataProfileID,
				QualityID:  album.Artist.QualityProfileID,
				Title:      album.Title,
				Release:    album.ReleaseDate,
				ArtistID:   album.ArtistID,
				ArtistName: album.Artist.ArtistName,
				ProfileID:  album.ProfileID,
				Overview:   album.Overview,
				Ratings:    album.Ratings.Value,
				Type:       album.AlbumType,
				Exists:     album.Statistics != nil && album.Statistics.SizeOnDisk > 0,
				Files:      0,
			})
		}
	}

	return http.StatusOK, output
}

func albumSearch(query, title string, releases []*lidarr.Release) bool {
	if strings.Contains(strings.ToLower(title), strings.ToLower(query)) {
		return true
	}

	for _, t := range releases {
		if strings.Contains(strings.ToLower(t.Title), strings.ToLower(query)) {
			return true
		}
	}

	return false
}

// @Description  Returns all Lidarr Tags.
// @Summary      Retrieve Lidarr Tags
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]starr.Tag} "tags"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/tag [get]
// @Security     ApiKeyAuth
func lidarrGetTags(req *http.Request) (int, interface{}) {
	tags, err := getLidarr(req).GetTagsContext(req.Context())
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "getting tags", err)
	}

	return http.StatusOK, tags
}

// @Description  Updates the label for a an existing tag.
// @Summary      Update Lidarr Tag Label
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        tagID     path   int64  true  "tag ID to update"
// @Param        label     path   string  true  "new label"
// @Success      200  {object} apps.Respond.apiResponse{message=int64}  "tag ID"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/tag/{tagID}/{label} [put]
// @Security     ApiKeyAuth
func lidarrUpdateTag(req *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(req)["tid"])

	tag, err := getLidarr(req).UpdateTagContext(req.Context(), &starr.Tag{ID: id, Label: mux.Vars(req)["label"]})
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating tag", err)
	}

	return http.StatusOK, tag.ID
}

// @Description  Creates a new tag with the provided label.
// @Summary      Create Lidarr Tag
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        label     path   string true  "new tag's label"
// @Success      200  {object} apps.Respond.apiResponse{message=int64}  "tag ID"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/tag/{label} [put]
// @Security     ApiKeyAuth
func lidarrSetTag(req *http.Request) (int, interface{}) {
	tag, err := getLidarr(req).AddTagContext(req.Context(), &starr.Tag{Label: mux.Vars(req)["label"]})
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "setting tag", err)
	}

	return http.StatusOK, tag.ID
}

// @Description  Updates an Album in Lidarr.
// @Summary      Update Lidarr Album
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        instance  path  int64  true  "instance ID"
// @Param        moveFiles query int64  true  "move files? true/false"
// @Param        PUT body lidarr.Album  true  "album content"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "ok"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json input"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/update [put]
// @Security     ApiKeyAuth
func lidarrUpdateAlbum(req *http.Request) (int, interface{}) {
	var album lidarr.Album

	err := json.NewDecoder(req.Body).Decode(&album)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	moveFiles := mux.Vars(req)["moveFiles"] == strconv.FormatBool(true)

	_, err = getLidarr(req).UpdateAlbumContext(req.Context(), album.ID, &album, moveFiles)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating album", err)
	}

	return http.StatusOK, mnd.Success
}

// @Description  Updates an Artist in Lidarr.
// @Summary      Update Lidarr Artist
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        instance  path  int64  true  "instance ID"
// @Param        PUT body lidarr.Artist  true  "album content"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "ok"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json input"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/updateartist [put]
// @Security     ApiKeyAuth
func lidarrUpdateArtist(req *http.Request) (int, interface{}) {
	var artist lidarr.Artist

	err := json.NewDecoder(req.Body).Decode(&artist)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	_, err = getLidarr(req).UpdateArtistContext(req.Context(), &artist, true)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating artist", err)
	}

	return http.StatusOK, mnd.Success
}

// @Description  Returns Lidarr Notifications with a name that matches 'notifiar'.
// @Summary      Retrieve Lidarr Notifications
// @Tags         Lidarr
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]lidarr.NotificationOutput} "notifications"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/notifications [get]
// @Security     ApiKeyAuth
func lidarrGetNotifications(req *http.Request) (int, interface{}) {
	notifs, err := getLidarr(req).GetNotificationsContext(req.Context())
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "getting notifications", err)
	}

	output := []*lidarr.NotificationOutput{}

	for _, notif := range notifs {
		if strings.Contains(strings.ToLower(notif.Name), "notifiar") {
			output = append(output, notif)
		}
	}

	return http.StatusOK, output
}

// @Description  Delete items from the activity queue.
// @Summary      Delete Queue Items
// @Tags         Lidarr
// @Produce      json
// @Param        instance         path    int64  true  "instance ID"
// @Param        queueID          path    int64  true  "queue ID to delete"
// @Param        removeFromClient query   bool  false  "remove download from download client?"
// @Param        blocklist        query   bool  false  "add item to blocklist?"
// @Param        skipRedownload   query   bool  false  "skip downloading this again?"
// @Param        changeCategory   query   bool  false  "tell download client to change categories?"
// @Success      200  {object} apps.Respond.apiResponse{message=string}  "ok"
// @Failure      500  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Failure      423  {object} string "rate limit reached"
// @Router       /api/lidarr/{instance}/queue/{queueID} [delete]
// @Security     ApiKeyAuth
func lidarrDeleteQueue(req *http.Request) (int, interface{}) {
	idString := mux.Vars(req)["queueID"]
	queueID, _ := strconv.ParseInt(idString, mnd.Base10, mnd.Bits64)
	removeFromClient := req.URL.Query().Get("removeFromClient") == mnd.True
	opts := &starr.QueueDeleteOpts{
		RemoveFromClient: &removeFromClient,
		BlockList:        req.URL.Query().Get("blocklist") == mnd.True,
		SkipRedownload:   req.URL.Query().Get("skipRedownload") == mnd.True,
		ChangeCategory:   req.URL.Query().Get("changeCategory") == mnd.True,
	}

	err := getLidarr(req).DeleteQueueContext(req.Context(), queueID, opts)
	if err != nil {
		return apiError(http.StatusInternalServerError, "deleting queue", err)
	}

	return http.StatusOK, mnd.Deleted + idString
}

// @Description  Updates a Notification in Lidarr.
// @Summary      Update Lidarr Notification
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        PUT body lidarr.NotificationInput  true  "notification content"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "ok"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "bad json input"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/notification [put]
// @Security     ApiKeyAuth
func lidarrUpdateNotification(req *http.Request) (int, interface{}) {
	var notif lidarr.NotificationInput

	err := json.NewDecoder(req.Body).Decode(&notif)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	_, err = getLidarr(req).UpdateNotificationContext(req.Context(), &notif)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "updating notification", err)
	}

	return http.StatusOK, mnd.Success
}

// @Description  Creates a new Lidarr Notification.
// @Summary      Add Lidarr Notification
// @Tags         Lidarr
// @Produce      json
// @Accept       json
// @Param        instance  path   int64  true  "instance ID"
// @Param        POST body lidarr.NotificationInput true "new item content"
// @Success      200  {object} apps.Respond.apiResponse{message=int64} "new notification ID"
// @Failure      400  {object} apps.Respond.apiResponse{message=string} "json input error"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/lidarr/{instance}/notification [post]
// @Security     ApiKeyAuth
func lidarrAddNotification(req *http.Request) (int, interface{}) {
	var notif lidarr.NotificationInput

	err := json.NewDecoder(req.Body).Decode(&notif)
	if err != nil {
		return apiError(http.StatusBadRequest, "decoding payload", err)
	}

	id, err := getLidarr(req).AddNotificationContext(req.Context(), &notif)
	if err != nil {
		return apiError(http.StatusServiceUnavailable, "adding notification", err)
	}

	return http.StatusOK, id
}
