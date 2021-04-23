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
	"golift.io/starr/lidarr"
)

/*
mbid - music brainz is the source for lidarr (todo)
*/

// lidarrHandlers is called once on startup to register the web API paths.
func (a *Apps) lidarrHandlers() {
	a.HandleAPIpath(Lidarr, "/add", lidarrAddAlbum, "POST")
	a.HandleAPIpath(Lidarr, "/artist/{artistid:[0-9]+}", lidarrGetArtist, "GET")
	a.HandleAPIpath(Lidarr, "/check/{mbid:[-a-z0-9]+}", lidarrCheckAlbum, "GET")
	a.HandleAPIpath(Lidarr, "/get/{albumid:[0-9]+}", lidarrGetAlbum, "GET")
	a.HandleAPIpath(Lidarr, "/metadataProfiles", lidarrMetadata, "GET")
	a.HandleAPIpath(Lidarr, "/qualityDefinitions", lidarrQualityDefs, "GET")
	a.HandleAPIpath(Lidarr, "/qualityProfiles", lidarrProfiles, "GET")
	a.HandleAPIpath(Lidarr, "/rootFolder", lidarrRootFolders, "GET")
	a.HandleAPIpath(Lidarr, "/search/{query}", lidarrSearchAlbum, "GET")
	a.HandleAPIpath(Lidarr, "/tag", lidarrGetTags, "GET")
	a.HandleAPIpath(Lidarr, "/tag/{tid:[0-9]+}/{label}", lidarrUpdateTag, "PUT")
	a.HandleAPIpath(Lidarr, "/tag/{label}", lidarrSetTag, "PUT")
	a.HandleAPIpath(Lidarr, "/update", lidarrUpdateAlbum, "PUT")
	a.HandleAPIpath(Lidarr, "/updateartist", lidarrUpdateArtist, "PUT")
	a.HandleAPIpath(Lidarr, "/command/search/{albumid:[0-9]+}", lidarrTriggerSearchAlbum, "GET")
}

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	Name     string
	Interval cnfg.Duration
	*starr.Config
	lidarr *lidarr.Lidarr
}

func (r *LidarrConfig) setup(timeout time.Duration) {
	r.lidarr = lidarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}
}

func lidarrAddAlbum(r *http.Request) (int, interface{}) {
	var payload lidarr.AddAlbumInput

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload.ForeignAlbumID == "" {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoMBID)
	}

	app := getLidarr(r)
	// Check for existing album.
	m, err := app.GetAlbum(payload.ForeignAlbumID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, lidarrData(m[0])
	}

	album, err := app.AddAlbum(&payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding album: %w", err)
	}

	return http.StatusCreated, album
}

func lidarrGetArtist(r *http.Request) (int, interface{}) {
	artistID, _ := strconv.ParseInt(mux.Vars(r)["artistid"], 10, 64)

	artist, err := getLidarr(r).GetArtistByID(artistID)
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
	}
}

func lidarrCheckAlbum(r *http.Request) (int, interface{}) {
	id := mux.Vars(r)["mbid"]

	m, err := getLidarr(r).GetAlbum(id)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, lidarrData(m[0])
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func lidarrGetAlbum(r *http.Request) (int, interface{}) {
	albumID, _ := strconv.ParseInt(mux.Vars(r)["albumid"], 10, 64)

	album, err := getLidarr(r).GetAlbumByID(albumID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
	}

	return http.StatusOK, album
}

func lidarrTriggerSearchAlbum(r *http.Request) (int, interface{}) {
	albumID, _ := strconv.ParseInt(mux.Vars(r)["albumid"], 10, 64)

	output, err := getLidarr(r).SendCommand(&lidarr.CommandRequest{
		Name:     "AlbumSearch",
		AlbumIDs: []int64{albumID},
	})
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("triggering album search: %w", err)
	}

	return http.StatusOK, output.Status
}

func lidarrMetadata(r *http.Request) (int, interface{}) {
	profiles, err := getLidarr(r).GetMetadataProfiles()
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

func lidarrQualityDefs(r *http.Request) (int, interface{}) {
	// Get the profiles from lidarr.
	definitions, err := getLidarr(r).GetQualityDefinition()
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

func lidarrProfiles(r *http.Request) (int, interface{}) {
	// Get the profiles from lidarr.
	profiles, err := getLidarr(r).GetQualityProfiles()
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

func lidarrRootFolders(r *http.Request) (int, interface{}) {
	// Get folder list from Lidarr.
	folders, err := getLidarr(r).GetRootFolders()
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

func lidarrSearchAlbum(r *http.Request) (int, interface{}) {
	albums, err := getLidarr(r).GetAlbum("")
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting albums: %w", err)
	}

	query := strings.TrimSpace(strings.ToLower(mux.Vars(r)["query"])) // in
	output := make([]map[string]interface{}, 0)                       // out

	for _, album := range albums {
		if albumSearch(query, album.Title, album.Releases) {
			a := map[string]interface{}{
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
				"exists":     false,
				"files":      0,
			}

			if album.Statistics != nil {
				a["exists"] = album.Statistics.SizeOnDisk > 0
			}

			output = append(output, a)
		}
	}

	return http.StatusOK, output
}

func albumSearch(query, title string, releases []*lidarr.Release) bool {
	if strings.Contains(strings.ToLower(title), query) {
		return true
	}

	for _, t := range releases {
		if strings.Contains(strings.ToLower(t.Title), query) {
			return true
		}
	}

	return false
}

func lidarrGetTags(r *http.Request) (int, interface{}) {
	tags, err := getLidarr(r).GetTags()
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting tags: %w", err)
	}

	return http.StatusOK, tags
}

func lidarrUpdateTag(r *http.Request) (int, interface{}) {
	id, _ := strconv.Atoi(mux.Vars(r)["tid"])

	tagID, err := getLidarr(r).UpdateTag(id, mux.Vars(r)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating tag: %w", err)
	}

	return http.StatusOK, tagID
}

func lidarrSetTag(r *http.Request) (int, interface{}) {
	tagID, err := getLidarr(r).AddTag(mux.Vars(r)["label"])
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting tag: %w", err)
	}

	return http.StatusOK, tagID
}

func lidarrUpdateAlbum(r *http.Request) (int, interface{}) {
	var album lidarr.Album

	err := json.NewDecoder(r.Body).Decode(&album)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	_, err = getLidarr(r).UpdateAlbum(album.ID, &album)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating album: %w", err)
	}

	return http.StatusOK, "success"
}

func lidarrUpdateArtist(r *http.Request) (int, interface{}) {
	var artist lidarr.Artist

	err := json.NewDecoder(r.Body).Decode(&artist)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	_, err = getLidarr(r).UpdateArtist(&artist)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating artist: %w", err)
	}

	return http.StatusOK, "success"
}
