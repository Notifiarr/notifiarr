package dnclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"golift.io/starr/lidarr"
)

/*
mbid - music brainz is the source for lidarr (todo)
*/

// lidarrHandlers is called once on startup to register the web API paths.
func (c *Client) lidarrHandlers() {
	c.handleAPIpath(Lidarr, "/add/{id:[0-9]+}", c.lidarrAddAlbum, "POST")
	c.handleAPIpath(Lidarr, "/check/{id:[0-9]+}/{albumid:[-a-z]+}", c.lidarrCheckAlbum, "GET")
	c.handleAPIpath(Lidarr, "/qualityProfiles/{id:[0-9]+}", c.lidarrProfiles, "GET")
	c.handleAPIpath(Lidarr, "/qualityDefinitions/{id:[0-9]+}", c.lidarrQualityDefs, "GET")
	c.handleAPIpath(Lidarr, "/rootFolder/{id:[0-9]+}", c.lidarrRootFolders, "GET")
}

func (c *Client) lidarrRootFolders(r *http.Request) (int, interface{}) {
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

func (c *Client) lidarrProfiles(r *http.Request) (int, interface{}) {
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

func (c *Client) lidarrQualityDefs(r *http.Request) (int, interface{}) {
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

func (c *Client) lidarrCheckAlbum(r *http.Request) (int, interface{}) {
	// Check for existing movie.
	if m, err := getLidarr(r).GetAlbum(mux.Vars(r)["albumid"]); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%s: %w", mux.Vars(r)["albumid"], ErrExists)
	}

	return http.StatusOK, http.StatusText(http.StatusNotFound)
}

func (c *Client) lidarrAddAlbum(r *http.Request) (int, interface{}) {
	var payload lidarr.AddAlbumInput
	// Extract payload and check for TMDB ID.
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	} else if payload == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("0: %w", ErrNoGRID)
	}

	lidar := getLidarr(r)
	// Check for existing album.
	/* broken:
	if m, err := lidar.GetAlbum(payload.AlbumID); err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("checking album: %w", err)
	} else if len(m) > 0 {
		return http.StatusConflict, fmt.Errorf("%d: %w", payload.AlbumID, ErrExists)
	}
	*/

	// Add book using payload.
	book, err := lidar.AddAlbum(&payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("adding album: %w", err)
	}

	return http.StatusCreated, book
}
