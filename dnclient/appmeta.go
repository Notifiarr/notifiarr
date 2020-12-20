//nolint:dupl
package dnclient

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"golift.io/starr"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

/* This file contains:
 *** The middleware procedure that stores the app interface in a request context.
 *** Procedures to save and fetch an app interface into/from a request content.
 */

// App allows safely storing context values.
type App string

// Constant for each app to unique identify itself.
// These strings are also used as a suffix to the /api/ web path.
const (
	Sonarr  App = "sonarr"
	Readarr App = "readarr"
	Radarr  App = "radarr"
	Lidarr  App = "lidarr"
)

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	*starr.Config
	*lidarr.Lidarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// RadarrConfig represents the input data for a Radarr server.
type RadarrConfig struct {
	*starr.Config
	*radarr.Radarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// ReadarrConfig represents the input data for a Readarr server.
type ReadarrConfig struct {
	*starr.Config
	*readarr.Readarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// SonarrConfig represents the input data for a Sonarr server.
type SonarrConfig struct {
	*starr.Config
	*sonarr.Sonarr
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// handleAPIpath makes adding API paths a little cleaner.
// This also grabs the app struct and saves it in a context before calling the handler.
func (c *Client) handleAPIpath(app App, api string, next apiHandle, method ...string) {
	if len(method) == 0 {
		method = []string{"GET"}
	}

	c.router.Handle(path.Join("/", c.Config.WebRoot, "api", string(app), api),
		c.checkAPIKey(c.responseWrapper(func(r *http.Request) (int, interface{}) {
			switch app {
			case Radarr:
				return c.setRadarr(r, next)
			case Lidarr:
				return c.setLidarr(r, next)
			case Sonarr:
				return c.setSonarr(r, next)
			case Readarr:
				return c.setReadarr(r, next)
			default: // unknown app, just run the handler.
				return next(r)
			}
		}))).Methods(method...)
}

/* Every API call runs one of these methods to save the interface into a request context for the respective app. */

// setLidarr saves the lidar config struct in a context so other methods can use it easily.
func (c *Client) setLidarr(r *http.Request, fn apiHandle) (int, interface{}) {
	var (
		id, _ = strconv.Atoi(mux.Vars(r)["id"])
		app   *LidarrConfig
	)

	for i, a := range c.Config.Lidarr {
		if i == id-1 { // discordnotifier wants 1-indexes
			app = a

			break
		}
	}

	if app == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoLidarr)
	}

	return fn(r.WithContext(context.WithValue(r.Context(), Lidarr, app)))
}

// setRadarr saves the radar config struct in a context so other methods can use it easily.
func (c *Client) setRadarr(r *http.Request, fn apiHandle) (int, interface{}) {
	var (
		id, _ = strconv.Atoi(mux.Vars(r)["id"])
		app   *RadarrConfig
	)

	for i, a := range c.Config.Radarr {
		if i == id-1 { // discordnotifier wants 1-indexes
			app = a

			break
		}
	}

	if app == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoRadarr)
	}

	return fn(r.WithContext(context.WithValue(r.Context(), Radarr, app)))
}

// setReadarr saves the readar config struct in a context so other methods can use it easily.
func (c *Client) setReadarr(r *http.Request, next apiHandle) (int, interface{}) {
	var (
		id, _ = strconv.Atoi(mux.Vars(r)["id"])
		app   *ReadarrConfig
	)

	for i, a := range c.Config.Readarr {
		if i == id-1 { // discordnotifier wants 1-indexes
			app = a

			break
		}
	}

	if app == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoReadarr)
	}

	return next(r.WithContext(context.WithValue(r.Context(), Readarr, app)))
}

// setSonarr saves the sonar config struct in a context so other methods can use it easily.
func (c *Client) setSonarr(r *http.Request, fn apiHandle) (int, interface{}) {
	var (
		id, _ = strconv.Atoi(mux.Vars(r)["id"])
		app   *SonarrConfig
	)

	for i, a := range c.Config.Sonarr {
		if i == id-1 { // discordnotifier wants 1-indexes
			app = a

			break
		}
	}

	if app == nil {
		return http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoSonarr)
	}

	return fn(r.WithContext(context.WithValue(r.Context(), Sonarr, app)))
}

/* Every API call runs one of these methods to find the interface for the respective app. */

func getLidarr(r *http.Request) *LidarrConfig {
	return r.Context().Value(Lidarr).(*LidarrConfig)
}

func getRadarr(r *http.Request) *RadarrConfig {
	return r.Context().Value(Radarr).(*RadarrConfig)
}

func getReadarr(r *http.Request) *ReadarrConfig {
	return r.Context().Value(Readarr).(*ReadarrConfig)
}

func getSonarr(r *http.Request) *SonarrConfig {
	return r.Context().Value(Sonarr).(*SonarrConfig)
}
