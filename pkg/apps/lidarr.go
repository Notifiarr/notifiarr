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
	a.HandleAPIpath(starr.Lidarr, "/qualityDefinitions", lidarrQualityDefs, "GET")
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
}

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	ExtraConfig
	*starr.Config
	*lidarr.Lidarr `toml:"-" xml:"-" json:"-"`
	errorf         func(string, ...interface{}) `toml:"-" xml:"-" json:"-"`
}

func getLidarr(r *http.Request) *lidarr.Lidarr {
	app, _ := r.Context().Value(starr.Lidarr).(*LidarrConfig)
	return app.Lidarr
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
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.ForeignAlbumID == "" {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoMBID)
	}

	// Check for existing album.
	m, err := getLidarr(req).GetAlbumContext(req.Context(), payload.ForeignAlbumID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, lidarrData(m[0])
	}

	album, err := getLidarr(req).AddAlbumContext(req.Context(), &payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding album: %w", err)
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
		return http.StatusServiceUnavailable, fmt.Errorf("checking artist: %w", err)
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
// @Param        mbid  path   int64  true  "movie brains ID"
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
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
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
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
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
		return http.StatusServiceUnavailable, fmt.Errorf("triggering album search: %w", err)
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
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int64]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
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
func lidarrQualityDefs(req *http.Request) (int, interface{}) {
	// Get the profiles from lidarr.
	definitions, err := getLidarr(req).GetQualityDefinitionContext(req.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
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
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
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
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
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
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// Get the profiles from radarr.
	id, err := getLidarr(req).AddQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding profile: %w", err)
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
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	profile.ID, _ = strconv.ParseInt(mux.Vars(req)["profileID"], mnd.Base10, mnd.Bits64)
	if profile.ID == 0 {
		return http.StatusUnprocessableEntity, ErrNonZeroID
	}

	// Get the profiles from radarr.
	_, err = getLidarr(req).UpdateQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating profile: %w", err)
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
		return http.StatusInternalServerError, fmt.Errorf("getting folders: %w", err)
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
		return http.StatusServiceUnavailable, fmt.Errorf("getting albums: %w", err)
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
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
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
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
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
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
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
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	moveFiles := mux.Vars(req)["moveFiles"] == fmt.Sprint(true)

	_, err = getLidarr(req).UpdateAlbumContext(req.Context(), album.ID, &album, moveFiles)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating album: %w", err)
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
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	_, err = getLidarr(req).UpdateArtistContext(req.Context(), &artist, true)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating artist: %w", err)
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
		return http.StatusServiceUnavailable, fmt.Errorf("getting notifications: %w", err)
	}

	output := []*lidarr.NotificationOutput{}

	for _, notif := range notifs {
		if strings.Contains(strings.ToLower(notif.Name), "notifiar") {
			output = append(output, notif)
		}
	}

	return http.StatusOK, output
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
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	_, err = getLidarr(req).UpdateNotificationContext(req.Context(), &notif)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating notification: %w", err)
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
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	id, err := getLidarr(req).AddNotificationContext(req.Context(), &notif)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("adding notification: %w", err)
	}

	return http.StatusOK, id
}
