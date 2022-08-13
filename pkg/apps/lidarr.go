//nolint:dupl
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
}

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	starrConfig
	*starr.Config
	*lidarr.Lidarr `toml:"-" xml:"-" json:"-"`
	errorf         func(string, ...interface{}) `toml:"-" xml:"-" json:"-"`
}

// Enabled returns true if the Lidarr instance is enabled and usable.
func (l *LidarrConfig) Enabled() bool {
	return l != nil && l.Config != nil && l.URL != "" && l.APIKey != "" && l.Timeout.Duration > 0
}

func (a *Apps) setupLidarr() error {
	for idx, app := range a.Lidarr {
		if app.Config == nil || app.Config.URL == "" {
			return fmt.Errorf("%w: missing url: Lidarr config %d", ErrInvalidApp, idx+1)
		}

		app.Config.Client = &http.Client{
			Timeout: app.Timeout.Duration,
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: exp.NewMetricsRoundTripper(string(starr.Lidarr), &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: app.Config.ValidSSL}, //nolint:gosec
			}),
		}
		app.Debugf = a.Debugf
		app.errorf = a.Errorf
		app.URL = strings.TrimRight(app.URL, "/")
		app.Lidarr = lidarr.New(app.Config)
	}

	return nil
}

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

func lidarrGetAlbum(req *http.Request) (int, interface{}) {
	albumID, _ := strconv.ParseInt(mux.Vars(req)["albumid"], mnd.Base10, mnd.Bits64)

	album, err := getLidarr(req).GetAlbumByIDContext(req.Context(), albumID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
	}

	return http.StatusOK, album
}

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

func lidarrGetQualityProfile(req *http.Request) (int, interface{}) {
	// Get the profiles from lidarr.
	profiles, err := getLidarr(req).GetQualityProfilesContext(req.Context())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	return http.StatusOK, profiles
}

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

func lidarrUpdateQualityProfile(req *http.Request) (int, interface{}) {
	var profile lidarr.QualityProfile

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
	err = getLidarr(req).UpdateQualityProfileContext(req.Context(), &profile)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("updating profile: %w", err)
	}

	return http.StatusOK, "OK"
}

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

func lidarrSearchAlbum(req *http.Request) (int, interface{}) {
	albums, err := getLidarr(req).GetAlbumContext(req.Context(), "")
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting albums: %w", err)
	}

	query := strings.TrimSpace(mux.Vars(req)["query"]) // in
	output := make([]map[string]interface{}, 0)        // out

	for _, album := range albums {
		if albumSearch(query, album.Title, album.Releases) {
			output = append(output, map[string]interface{}{
				"id":         album.ID,
				"mbid":       album.ForeignAlbumID,
				"metadataId": album.Artist.MetadataProfileID,
				"qualityId":  album.Artist.QualityProfileID,
				"title":      album.Title,
				"release":    album.ReleaseDate,
				"artistId":   album.ArtistID,
				"artistName": album.Artist.ArtistName,
				"profileId":  album.ProfileID,
				"overview":   album.Overview,
				"ratings":    album.Ratings.Value,
				"type":       album.AlbumType,
				"exists":     album.Statistics != nil && album.Statistics.SizeOnDisk > 0,
				"files":      0,
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

func lidarrGetTags(req *http.Request) (int, interface{}) {
	tags, err := getLidarr(req).GetTagsContext(req.Context())
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
	}

	return http.StatusOK, tags
}

func lidarrUpdateTag(req *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(req)["tid"])

	tag, err := getLidarr(req).UpdateTagContext(req.Context(), &starr.Tag{ID: id, Label: mux.Vars(req)["label"]})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
	}

	return http.StatusOK, tag.ID
}

func lidarrSetTag(req *http.Request) (int, interface{}) {
	tag, err := getLidarr(req).AddTagContext(req.Context(), &starr.Tag{Label: mux.Vars(req)["label"]})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
	}

	return http.StatusOK, tag.ID
}

func lidarrUpdateAlbum(req *http.Request) (int, interface{}) {
	var album lidarr.Album

	err := json.NewDecoder(req.Body).Decode(&album)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	_, err = getLidarr(req).UpdateAlbumContext(req.Context(), album.ID, &album)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating album: %w", err)
	}

	return http.StatusOK, "success"
}

func lidarrUpdateArtist(req *http.Request) (int, interface{}) {
	var artist lidarr.Artist

	err := json.NewDecoder(req.Body).Decode(&artist)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	_, err = getLidarr(req).UpdateArtistContext(req.Context(), &artist)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating artist: %w", err)
	}

	return http.StatusOK, "success"
}
