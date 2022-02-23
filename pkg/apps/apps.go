// Package apps provides the _incoming_ HTTP methods for notifiarr.com integrations.
// Methods are included for Radarr, Readrr, Lidarr and Sonarr. This library also
// holds the site API Key and the base HTTP server abstraction used throughout
// the Notifiarr client application. The configuration should be derived from
// a config file; a Router and an Error Log logger must also be provided.
package apps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/gorilla/mux"
	"github.com/miolini/datacounter"
	"golift.io/cnfg"
	"golift.io/starr"
	"golift.io/starr/lidarr"
	"golift.io/starr/prowlarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

var apiHits = exp.GetMap("apiHits") //nolint:gochecknoglobals

// Apps is the input configuration to relay requests to Starr apps.
type Apps struct {
	APIKey   string              `json:"apiKey" toml:"api_key" xml:"api_key" yaml:"apiKey"`
	ExKeys   []string            `json:"extraKeys" toml:"extra_keys" xml:"extra_keys" yaml:"extraKeys"`
	URLBase  string              `json:"urlbase" toml:"urlbase" xml:"urlbase" yaml:"urlbase"`
	Sonarr   []*SonarrConfig     `json:"sonarr,omitempty" toml:"sonarr" xml:"sonarr" yaml:"sonarr,omitempty"`
	Radarr   []*RadarrConfig     `json:"radarr,omitempty" toml:"radarr" xml:"radarr" yaml:"radarr,omitempty"`
	Lidarr   []*LidarrConfig     `json:"lidarr,omitempty" toml:"lidarr" xml:"lidarr" yaml:"lidarr,omitempty"`
	Readarr  []*ReadarrConfig    `json:"readarr,omitempty" toml:"readarr" xml:"readarr" yaml:"readarr,omitempty"`
	Prowlarr []*ProwlarrConfig   `json:"prowlarr,omitempty" toml:"prowlarr" xml:"prowlarr" yaml:"prowlarr,omitempty"`
	Deluge   []*DelugeConfig     `json:"deluge,omitempty" toml:"deluge" xml:"deluge" yaml:"deluge,omitempty"`
	Qbit     []*QbitConfig       `json:"qbit,omitempty" toml:"qbit" xml:"qbit" yaml:"qbit,omitempty"`
	SabNZB   []*SabNZBConfig     `json:"sabnzbd,omitempty" toml:"sabnzbd" xml:"sabnzbd" yaml:"sabnzbd,omitempty"`
	Tautulli *TautulliConfig     `json:"tautulli,omitempty" toml:"tautulli" xml:"tautulli" yaml:"tautulli,omitempty"`
	Router   *mux.Router         `json:"-" toml:"-" xml:"-" yaml:"-"`
	ErrorLog *log.Logger         `json:"-" toml:"-" xml:"-" yaml:"-"`
	DebugLog *log.Logger         `json:"-" toml:"-" xml:"-" yaml:"-"`
	keys     map[string]struct{} `toml:"-"` // for fast key lookup.
}

type starrConfig struct {
	Name      string        `toml:"name" xml:"name" json:"name"`
	Interval  cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	StuckItem bool          `toml:"stuck_items" xml:"stuck_items" json:"stuckItems"`
	Corrupt   string        `toml:"corrupt" xml:"corrupt" json:"corrupt"`
	Backup    string        `toml:"backup" xml:"backup" json:"backup"`
}

// Errors sent to client web requests.
var (
	ErrNoTMDB    = fmt.Errorf("TMDB ID must not be empty")
	ErrNoGRID    = fmt.Errorf("GRID ID must not be empty")
	ErrNoTVDB    = fmt.Errorf("TVDB ID must not be empty")
	ErrNoMBID    = fmt.Errorf("MBID ID must not be empty")
	ErrNoRadarr  = fmt.Errorf("configured %s ID not found", starr.Radarr)
	ErrNoSonarr  = fmt.Errorf("configured %s ID not found", starr.Sonarr)
	ErrNoLidarr  = fmt.Errorf("configured %s ID not found", starr.Lidarr)
	ErrNoReadarr = fmt.Errorf("configured %s ID not found", starr.Readarr)
	ErrNotFound  = fmt.Errorf("the request returned an empty payload")
	ErrNonZeroID = fmt.Errorf("provided ID must be non-zero")
	// ErrWrongCount is returned when an app returns the wrong item count.
	ErrWrongCount = fmt.Errorf("wrong item count returned")
	ErrInvalidApp = fmt.Errorf("invalid application configuration provided")
)

// APIHandler is our custom handler function for APIs.
// The powers the middleware procedure that stores the app interface in a request context.
// And the procedures to save and fetch an app interface into/from a request content.
type APIHandler func(r *http.Request) (int, interface{})

// HandleAPIpath makes adding APIKey authenticated API paths a little cleaner.
// An empty App may be passed in, but URI, API and at least one method are required.
// Automatically adds an id route to routes with an app name. In case you have > 1 of that app.
func (a *Apps) HandleAPIpath(app starr.App, uri string, api APIHandler, method ...string) *mux.Route {
	if len(method) == 0 {
		method = []string{"GET"}
	}

	id := "{id:[0-9]+}"
	if app == "" {
		id = ""
	}

	uri = path.Join(a.URLBase, "api", app.Lower(), id, uri)

	return a.Router.Handle(uri, a.CheckAPIKey(a.handleAPI(app, api))).Methods(method...)
}

// This grabs the app struct and saves it in a context before calling the handler.
// The purpose of this complicated monster is to keep API handler methods simple.
func (a *Apps) handleAPI(app starr.App, api APIHandler) http.HandlerFunc { //nolint:cyclop
	return func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		var (
			msg     interface{}
			ctx     = r.Context()
			code    = http.StatusUnprocessableEntity
			id, _   = strconv.Atoi(mux.Vars(r)["id"])
			start   = time.Now()
			buf     bytes.Buffer
			tee     = io.TeeReader(r.Body, &buf) // must read tee first.
			post, _ = ioutil.ReadAll(tee)
			appName = app.String()
		)

		if appName == "" {
			appName = "noApp"
		}

		apiHits.Add(appName+"bytesRcvd", int64(len(post)))
		apiHits.Add(appName+"count", 1)
		apiHits.Add("requests", 1)

		r.Body.Close() // we just read this into a buffer.

		// notifiarr.com uses 1-indexes; subtract 1 from the ID (turn 1 into 0 generally).
		switch id--; {
		// Make sure the id is within range of the available service.
		case app == starr.Lidarr && (id >= len(a.Lidarr) || id < 0):
			msg = fmt.Errorf("%v: %w", id, ErrNoLidarr)
		case app == starr.Prowlarr && (id >= len(a.Prowlarr) || id < 0):
			msg = fmt.Errorf("%v: %w", id, ErrNoLidarr)
		case app == starr.Radarr && (id >= len(a.Radarr) || id < 0):
			msg = fmt.Errorf("%v: %w", id, ErrNoRadarr)
		case app == starr.Readarr && (id >= len(a.Readarr) || id < 0):
			msg = fmt.Errorf("%v: %w", id, ErrNoReadarr)
		case app == starr.Sonarr && (id >= len(a.Sonarr) || id < 0):
			msg = fmt.Errorf("%v: %w", id, ErrNoSonarr)
			// Store the application configuration (starr) in a context then pass that into the api() method.
			// Retrieve the return code and output, and send a response via a.Respond().
		case app == starr.Lidarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Lidarr[id])))
		case app == starr.Prowlarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Prowlarr[id])))
		case app == starr.Radarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Radarr[id])))
		case app == starr.Readarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Readarr[id])))
		case app == starr.Sonarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Sonarr[id])))
		case app == "":
			// no app, just run the handler.
			code, msg = api(r) // unknown app, just run the handler.
		default:
			// unknown app, add the ID to the context and run the handler.
			code, msg = api(r.WithContext(context.WithValue(ctx, app, id)))
		}

		if len(post) > 0 {
			s, _ := json.MarshalIndent(msg, "", " ")
			a.DebugLog.Printf("Incoming API: %s %s: %s\nStatus: %d, Reply: %s", r.Method, r.URL, string(post), code, s)
		}

		wrote := a.Respond(w, code, msg)
		apiHits.Add(appName+"bytesSent", wrote)
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
	}
}

// CheckAPIKey drops a 403 if the API key doesn't match, otherwise run next handler.
func (a *Apps) CheckAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		if _, ok := a.keys[r.Header.Get("X-API-Key")]; !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// InitHandlers activates all our handlers. This is part of the web server init.
func (a *Apps) InitHandlers() {
	a.keys = make(map[string]struct{})
	for _, key := range append(a.ExKeys, a.APIKey) {
		if len(key) > 3 { //nolint:gomnd
			a.keys[key] = struct{}{}
		}
	}

	a.lidarrHandlers()
	a.prowlarrHandlers()
	a.radarrHandlers()
	a.readarrHandlers()
	a.sonarrHandlers()
}

// Setup creates request interfaces and sets the timeout for each server.
// This is part of the config/startup init.
func (a *Apps) Setup(timeout time.Duration) error { //nolint:cyclop
	if a.DebugLog == nil {
		a.DebugLog = log.New(io.Discard, "", 0)
	}

	if a.ErrorLog == nil {
		a.ErrorLog = log.New(io.Discard, "", 0)
	}

	if err := a.setupLidarr(timeout); err != nil {
		return err
	}

	if err := a.setupProwlarr(timeout); err != nil {
		return err
	}

	if err := a.setupRadarr(timeout); err != nil {
		return err
	}

	if err := a.setupReadarr(timeout); err != nil {
		return err
	}

	if err := a.setupSonarr(timeout); err != nil {
		return err
	}

	if err := a.setupDeluge(timeout); err != nil {
		return err
	}

	if err := a.setupQbit(timeout); err != nil {
		return err
	}

	for i := range a.SabNZB {
		a.SabNZB[i].setup(timeout)
	}

	a.Tautulli.setup(timeout)

	return nil
}

// Respond sends a standard response to our caller. JSON encoded blobs. Returns size of data sent.
func (a *Apps) Respond(w http.ResponseWriter, stat int, msg interface{}) int64 { //nolint:varnamelen
	if stat == http.StatusFound || stat == http.StatusMovedPermanently ||
		stat == http.StatusPermanentRedirect || stat == http.StatusTemporaryRedirect {
		w.Header().Set("Location", msg.(string))
		w.WriteHeader(stat)
		apiHits.Add("codes:"+http.StatusText(stat), 1)

		return 0
	}

	statusTxt := strconv.Itoa(stat) + ": " + http.StatusText(stat)

	if m, ok := msg.(error); ok {
		a.ErrorLog.Printf("Request failed. Status: %s, Message: %v", statusTxt, m)
		msg = m.Error()
	}

	apiHits.Add("codes:"+http.StatusText(stat), 1)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stat)
	counter := datacounter.NewResponseWriterCounter(w)
	json := json.NewEncoder(counter)
	json.SetEscapeHTML(false)

	err := json.Encode(map[string]interface{}{"status": statusTxt, "message": msg})
	if err != nil {
		a.ErrorLog.Printf("Sending JSON response failed. Status: %s, Error: %v, Message: %v", statusTxt, err, msg)
	}

	return int64(counter.Count())
}

/* Every API call runs one of these methods to find the interface for the respective app. */

func getLidarr(r *http.Request) *lidarr.Lidarr {
	apiHits.Add("LidarrReqsFromApi", 1)
	return r.Context().Value(starr.Lidarr).(*LidarrConfig).Lidarr
}

//nolint:deadcode,unused // will be used when we add http handlers for prowlarr.
func getProwlarr(r *http.Request) *prowlarr.Prowlarr {
	apiHits.Add("ProwlarrReqsFromApi", 1)
	return r.Context().Value(starr.Prowlarr).(*ProwlarrConfig).Prowlarr
}

func getRadarr(r *http.Request) *radarr.Radarr {
	apiHits.Add("RadarrReqsFromApi", 1)
	return r.Context().Value(starr.Radarr).(*RadarrConfig).Radarr
}

func getReadarr(r *http.Request) *readarr.Readarr {
	apiHits.Add("ReadarrReqsFromApi", 1)
	return r.Context().Value(starr.Readarr).(*ReadarrConfig).Readarr
}

func getSonarr(r *http.Request) *sonarr.Sonarr {
	apiHits.Add("SonarrReqsFromApi", 1)
	return r.Context().Value(starr.Sonarr).(*SonarrConfig).Sonarr
}
