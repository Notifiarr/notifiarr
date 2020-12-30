package apps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/lidarr"
)

/*
mbid - music brainz is the source for lidarr (todo)
*/

// lidarrHandlers is called once on startup to register the web API paths.
func (a *Apps) lidarrHandlers() {
	a.HandleAPIpath(Lidarr, "/add", lidarrAddAlbum, "POST")
	a.HandleAPIpath(Lidarr, "/check/{mbid:[-a-z0-9]+}", lidarrCheckAlbum, "GET")
	a.HandleAPIpath(Lidarr, "/get/{albumid:[0-9]+}", lidarrGetAlbum, "GET")
	a.HandleAPIpath(Lidarr, "/update", lidarrUpdateAlbum, "PUT")
	a.HandleAPIpath(Lidarr, "/qualityProfiles", lidarrProfiles, "GET")
	a.HandleAPIpath(Lidarr, "/qualityDefinitions", lidarrQualityDefs, "GET")
	a.HandleAPIpath(Lidarr, "/rootFolder", lidarrRootFolders, "GET")
}

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	*starr.Config
	lidarr *lidarr.Lidarr
}

func (r *LidarrConfig) setup(timeout time.Duration) {
	r.lidarr = lidarr.New(r.Config)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = timeout
	}
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

func lidarrCheckAlbum(r *http.Request) (int, interface{}) {
	id := mux.Vars(r)["mbid"]

	m, err := getLidarr(r).GetAlbum(id)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%s: %w", id, ErrExists)
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

func lidarrUpdateAlbum(r *http.Request) (int, interface{}) {
	var album lidarr.Album

	err := json.NewDecoder(r.Body).Decode(&album)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	err = getLidarr(r).UpdateAlbum(album.ID, &album)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("updating album: %w", err)
	}

	return http.StatusOK, "lidarr seems to have worked"
}

func lidarrAddAlbum(r *http.Request) (int, interface{}) {
	var payload lidarr.AddAlbumInput

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoGRID)
	}

	app := getLidarr(r)
	// Check for existing album.
	/* broken:
	m, err := lidar.GetAlbum(payload.AlbumID)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.AlbumID, ErrExists)
	}
	*/

	album, err := app.AddAlbum(&payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding album: %w", err)
	}

	return http.StatusCreated, album
}
