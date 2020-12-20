//nolint:dupl
package dnclient

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

/* This file contains:
 *** The middleware procedure that stores the app interface in a request context.
 *** Startup logging procedures for each app.
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

// serveAPIpath makes adding API paths a little cleaner.
// This also grabs the app struct and saves it in a context before calling the handler.
func (c *Client) serveAPIpath(app App, webPath, method string, next apiHandle) {
	c.router.Handle(path.Join("/", c.Config.WebRoot, "api", string(app), webPath),
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
		}))).Methods(method)
}

// initLidarr is called on startup to fix the config and print info about each configured server.
func (c *Client) initLidarr() {
	for i := range c.Config.Lidarr {
		if c.Config.Lidarr[i].Timeout.Duration == 0 {
			c.Config.Lidarr[i].Timeout.Duration = c.Config.Timeout.Duration
		}

		c.Config.Lidarr[i].Lidarr = lidarr.New(c.Config.Lidarr[i].Config)
	}

	if count := len(c.Config.Lidarr); count == 1 {
		c.Printf(" => Lidarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Lidarr[0].URL, c.Config.Lidarr[0].APIKey != "", c.Config.Lidarr[0].Timeout, c.Config.Lidarr[0].ValidSSL)
	} else {
		c.Print(" => Lidarr Config:", count, "servers")

		for _, f := range c.Config.Lidarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// initRadarr is called on startup to fix the config and print info about each configured server.
func (c *Client) initRadarr() {
	for i := range c.Config.Radarr {
		if c.Config.Radarr[i].Timeout.Duration == 0 {
			c.Config.Radarr[i].Timeout.Duration = c.Config.Timeout.Duration
		}

		c.Config.Radarr[i].Radarr = radarr.New(c.Config.Radarr[i].Config)
	}

	if count := len(c.Config.Radarr); count == 1 {
		c.Printf(" => Radarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Radarr[0].URL, c.Config.Radarr[0].APIKey != "", c.Config.Radarr[0].Timeout, c.Config.Radarr[0].ValidSSL)
	} else {
		c.Print(" => Radarr Config:", count, "servers")

		for _, f := range c.Config.Radarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// initReadarr is called on startup to fix the config and print info about each configured server.
func (c *Client) initReadarr() {
	for i := range c.Config.Readarr {
		if c.Config.Readarr[i].Timeout.Duration == 0 {
			c.Config.Readarr[i].Timeout.Duration = c.Config.Timeout.Duration
		}

		c.Config.Readarr[i].Readarr = readarr.New(c.Config.Readarr[i].Config)
	}

	if count := len(c.Config.Readarr); count == 1 {
		c.Printf(" => Readarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Readarr[0].URL, c.Config.Readarr[0].APIKey != "", c.Config.Readarr[0].Timeout, c.Config.Readarr[0].ValidSSL)
	} else {
		c.Print(" => Readarr Config:", count, "servers")

		for _, f := range c.Config.Readarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// initSonarr is called on startup to fix the config and print info about each configured server.
func (c *Client) initSonarr() {
	for i := range c.Config.Sonarr {
		if c.Config.Sonarr[i].Timeout.Duration == 0 {
			c.Config.Sonarr[i].Timeout.Duration = c.Config.Timeout.Duration
		}

		c.Config.Sonarr[i].Sonarr = sonarr.New(c.Config.Sonarr[i].Config)
	}

	if count := len(c.Config.Sonarr); count == 1 {
		c.Printf(" => Sonarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Sonarr[0].URL, c.Config.Sonarr[0].APIKey != "", c.Config.Sonarr[0].Timeout, c.Config.Sonarr[0].ValidSSL)
	} else {
		c.Print(" => Sonarr Config:", count, "servers")

		for _, f := range c.Config.Sonarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
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
