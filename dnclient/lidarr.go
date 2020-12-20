package dnclient

import (
	"fmt"
	"net/http"
	"sync"

	"golift.io/starr"
	"golift.io/starr/lidarr"
)

/*
mbid - music brainz is the source for lidarr (todo)
*/

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	*starr.Config
	*lidarr.Lidarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// lidarrHandlers is called once on startup to register the web API paths.
func (c *Client) lidarrHandlers() {
	c.serveAPIpath(Lidarr, "/add/{id:[0-9]+}", "POST", c.lidarrAddAlbum)
	c.serveAPIpath(Lidarr, "/check/{id:[0-9]+}/{albumid:[0-9]+}", "GET", c.lidarrCheckAlbum)
	c.serveAPIpath(Lidarr, "/qualityProfiles/{id:[0-9]+}", "GET", c.lidarrProfiles)
	c.serveAPIpath(Lidarr, "/qualityDefinitions/{id:[0-9]+}", "GET", c.lidarrQualityDefs)
	c.serveAPIpath(Lidarr, "/rootFolder/{id:[0-9]+}", "GET", c.lidarrRootFolders)
}

func (c *Client) lidarrRootFolders(r *http.Request) (int, interface{}) {
	lidar := getLidarr(r)

	// Get folder list from Lidarr.
	folders, err := lidar.GetRootFolders()
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
	lidar := getLidarr(r)

	// Get the profiles from lidarr.
	profiles, err := lidar.GetQualityProfiles()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format profile ID=>Name into a nice map.
	p := make(map[int]string)
	for i := range profiles {
		p[profiles[i].ID] = profiles[i].Name
	}

	return http.StatusOK, p
}

func (c *Client) lidarrQualityDefs(r *http.Request) (int, interface{}) {
	lidar := getLidarr(r)

	// Get the profiles from lidarr.
	definitions, err := lidar.GetQualityDefinition()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("getting profiles: %w", err)
	}

	// Format definitions ID=>Title into a nice map.
	p := make(map[int]string)
	for i := range definitions {
		p[definitions[i].ID] = definitions[i].Title
	}

	return http.StatusOK, p
}

func (c *Client) lidarrCheckAlbum(r *http.Request) (int, interface{}) {
	return http.StatusLocked, "lidarr does not work yet"
}

func (c *Client) lidarrAddAlbum(r *http.Request) (int, interface{}) {
	return http.StatusLocked, "lidarr does not work yet"
}
