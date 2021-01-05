package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

/* This file contains:
 *** The middleware procedure that stores the app interface in a request context.
 *** Procedures to save and fetch an app interface into/from a request content.
 */

// Apps is the input configuration to relay requests to Starr apps.
type Apps struct {
	APIKey   string           `json:"api_key" toml:"api_key" xml:"api_key" yaml:"api_key"`
	URLBase  string           `json:"urlbase" toml:"urlbase" xml:"urlbase" yaml:"urlbase"`
	Sonarr   []*SonarrConfig  `json:"sonarr,omitempty" toml:"sonarr" xml:"sonarr" yaml:"sonarr,omitempty"`
	Radarr   []*RadarrConfig  `json:"radarr,omitempty" toml:"radarr" xml:"radarr" yaml:"radarr,omitempty"`
	Lidarr   []*LidarrConfig  `json:"lidarr,omitempty" toml:"lidarr" xml:"lidarr" yaml:"lidarr,omitempty"`
	Readarr  []*ReadarrConfig `json:"readarr,omitempty" toml:"readarr" xml:"readarr" yaml:"readarr,omitempty"`
	Router   *mux.Router      `json:"-" toml:"-" xml:"-" yaml:"-"`
	ErrorLog *log.Logger      `json:"-" toml:"-" xml:"-" yaml:"-"`
}

// Responder converts all our data to a JSON response.

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

// Errors sent to client web requests.
var (
	ErrNoTMDB    = fmt.Errorf("TMDB ID must not be empty")
	ErrNoGRID    = fmt.Errorf("GRID ID must not be empty")
	ErrNoTVDB    = fmt.Errorf("TVDB ID must not be empty")
	ErrNoMBID    = fmt.Errorf("MBID ID must not be empty")
	ErrNoRadarr  = fmt.Errorf("configured radarr ID not found")
	ErrNoSonarr  = fmt.Errorf("configured sonarr ID not found")
	ErrNoLidarr  = fmt.Errorf("configured lidarr ID not found")
	ErrNoReadarr = fmt.Errorf("configured readarr ID not found")
	ErrExists    = fmt.Errorf("the requested item already exists")
	ErrNotFound  = fmt.Errorf("the request returned an empty payload")
)

// APIHandler is our custom handler function for APIs.
type APIHandler func(r *http.Request) (int, interface{})

// HandleAPIpath makes adding API paths a little cleaner.
// This grabs the app struct and saves it in a context before calling the handler.
func (a *Apps) HandleAPIpath(app App, uri string, api APIHandler, method ...string) {
	if len(method) == 0 {
		method = []string{"GET"}
	}

	id := "{id:[0-9]+}"
	if app == "" {
		id = ""
	}

	uri = path.Join("/", a.URLBase, "api", string(app), id, uri)
	a.Router.Handle(uri, a.checkAPIKey(a.handleAPIpath(app, api))).Methods(method...)
}

func (a *Apps) handleAPIpath(app App, api APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(mux.Vars(r)["id"])

		switch start := time.Now(); {
		case app == Radarr && (id > len(a.Radarr) || id < 1):
			a.Respond(w, http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoRadarr), time.Since(start))
		case app == Lidarr && (id > len(a.Lidarr) || id < 1):
			a.Respond(w, http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoLidarr), time.Since(start))
		case app == Sonarr && (id > len(a.Sonarr) || id < 1):
			a.Respond(w, http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoSonarr), time.Since(start))
		case app == Readarr && (id > len(a.Readarr) || id < 1):
			a.Respond(w, http.StatusUnprocessableEntity, fmt.Errorf("%v: %w", id, ErrNoReadarr), time.Since(start))
		// These store the application configuration (starr) in a context then pass that into the next method.
		// They retrieve the return code and output, then send a response (a.Respond).
		// disccordnotifier.com uses 1-indexes, so we subtract 1 from the ID (turn 1 into 0).
		case app == Radarr:
			code, msg := api(r.WithContext(context.WithValue(r.Context(), Radarr, a.Radarr[id-1])))
			a.Respond(w, code, msg, time.Since(start))
		case app == Lidarr:
			code, msg := api(r.WithContext(context.WithValue(r.Context(), Lidarr, a.Lidarr[id-1])))
			a.Respond(w, code, msg, time.Since(start))
		case app == Sonarr:
			code, msg := api(r.WithContext(context.WithValue(r.Context(), Sonarr, a.Sonarr[id-1])))
			a.Respond(w, code, msg, time.Since(start))
		case app == Readarr:
			code, msg := api(r.WithContext(context.WithValue(r.Context(), Readarr, a.Readarr[id-1])))
			a.Respond(w, code, msg, time.Since(start))
		default:
			// unknown app, just run the handler.
			code, msg := api(r)
			a.Respond(w, code, msg, time.Since(start))
		}
	}
}

// checkAPIKey drops a 403 if the API key doesn't match.
func (a *Apps) checkAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != a.APIKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// InitHandlers activates all our handlers.
func (a *Apps) InitHandlers() {
	a.radarrHandlers()
	a.readarrHandlers()
	a.lidarrHandlers()
	a.sonarrHandlers()
}

// Setup creates request interfaces and sets the timeout for each server.
func (a *Apps) Setup(timeout time.Duration) {
	for i := range a.Radarr {
		a.Radarr[i].setup(timeout)
	}

	for i := range a.Readarr {
		a.Readarr[i].setup(timeout)
	}

	for i := range a.Sonarr {
		a.Sonarr[i].setup(timeout)
	}

	for i := range a.Lidarr {
		a.Lidarr[i].setup(timeout)
	}
}

// Respond sends a standard response to our caller. JSON encoded blobs.
func (a *Apps) Respond(w http.ResponseWriter, stat int, msg interface{}, reqTime time.Duration) {
	w.Header().Set("X-Request-Time", fmt.Sprintf("%dms", reqTime.Milliseconds()))

	statusTxt := strconv.Itoa(stat) + ": " + http.StatusText(stat)

	if m, ok := msg.(error); ok {
		a.ErrorLog.Printf("Request failed. Status: %s, Message: %v", statusTxt, m)
		msg = m.Error()
	}

	if stat == http.StatusFound || stat == http.StatusMovedPermanently ||
		stat == http.StatusPermanentRedirect || stat == http.StatusTemporaryRedirect {
		w.Header().Set("Location", msg.(string))
		w.WriteHeader(stat)

		return
	}

	b, err := json.Marshal(map[string]interface{}{"status": statusTxt, "message": msg})
	if err != nil {
		a.ErrorLog.Printf("JSON marshal failed. Status: %s, Error: %v, Message: %v", statusTxt, err, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stat)

	size, err := w.Write(append(b, '\n')) // curl likes new lines.
	if err != nil {
		a.ErrorLog.Printf("Response failed. Written: %d/%d, Status: %s, Error: %v", size, len(b)+1, statusTxt, err)
	}
}

/* Every API call runs one of these methods to find the interface for the respective app. */

func getLidarr(r *http.Request) *lidarr.Lidarr {
	return r.Context().Value(Lidarr).(*LidarrConfig).lidarr
}

func getRadarr(r *http.Request) *radarr.Radarr {
	return r.Context().Value(Radarr).(*RadarrConfig).radarr
}

func getReadarr(r *http.Request) *readarr.Readarr {
	return r.Context().Value(Readarr).(*ReadarrConfig).readarr
}

func getSonarr(r *http.Request) *sonarr.Sonarr {
	return r.Context().Value(Sonarr).(*SonarrConfig).sonarr
}
