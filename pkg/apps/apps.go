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
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	"golift.io/cnfg"
	"golift.io/datacounter"
	"golift.io/starr"
	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

// Apps is the input configuration to relay requests to Starr apps.
type Apps struct {
	APIKey     string            `json:"apiKey" toml:"api_key" xml:"api_key" yaml:"apiKey"`
	ExKeys     []string          `json:"extraKeys" toml:"extra_keys" xml:"extra_keys" yaml:"extraKeys"`
	URLBase    string            `json:"urlbase" toml:"urlbase" xml:"urlbase" yaml:"urlbase"`
	MaxBody    int               `toml:"max_body" xml:"max_body" json:"maxBody"`
	Sonarr     []*SonarrConfig   `json:"sonarr,omitempty" toml:"sonarr" xml:"sonarr" yaml:"sonarr,omitempty"`
	Radarr     []*RadarrConfig   `json:"radarr,omitempty" toml:"radarr" xml:"radarr" yaml:"radarr,omitempty"`
	Lidarr     []*LidarrConfig   `json:"lidarr,omitempty" toml:"lidarr" xml:"lidarr" yaml:"lidarr,omitempty"`
	Readarr    []*ReadarrConfig  `json:"readarr,omitempty" toml:"readarr" xml:"readarr" yaml:"readarr,omitempty"`
	Prowlarr   []*ProwlarrConfig `json:"prowlarr,omitempty" toml:"prowlarr" xml:"prowlarr" yaml:"prowlarr,omitempty"`
	Deluge     []*DelugeConfig   `json:"deluge,omitempty" toml:"deluge" xml:"deluge" yaml:"deluge,omitempty"`
	Qbit       []*QbitConfig     `json:"qbit,omitempty" toml:"qbit" xml:"qbit" yaml:"qbit,omitempty"`
	Rtorrent   []*RtorrentConfig `json:"rtorrent,omitempty" toml:"rtorrent" xml:"rtorrent" yaml:"rtorrent,omitempty"`
	SabNZB     []*SabNZBConfig   `json:"sabnzbd,omitempty" toml:"sabnzbd" xml:"sabnzbd" yaml:"sabnzbd,omitempty"`
	NZBGet     []*NZBGetConfig   `json:"nzbget,omitempty" toml:"nzbget" xml:"nzbget" yaml:"nzbget,omitempty"`
	Tautulli   *TautulliConfig   `json:"tautulli,omitempty" toml:"tautulli" xml:"tautulli" yaml:"tautulli,omitempty"`
	Router     *mux.Router       `json:"-" toml:"-" xml:"-" yaml:"-"`
	mnd.Logger `toml:"-" xml:"-" json:"-"`
	keys       map[string]struct{} `toml:"-"` // for fast key lookup.
}

type starrConfig struct {
	Name     string        `toml:"name" xml:"name" json:"name"`
	Timeout  cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`
	Interval cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`
	ValidSSL bool          `toml:"valid_ssl" xml:"valid_ssl" json:"validSsl"`
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
func (a *Apps) handleAPI(app starr.App, api APIHandler) http.HandlerFunc { //nolint:cyclop,funlen,gocognit
	return func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		var (
			msg     interface{}
			ctx     = r.Context()
			code    = http.StatusUnprocessableEntity
			aID, _  = strconv.Atoi(mux.Vars(r)["id"])
			start   = time.Now()
			buf     bytes.Buffer
			tee     = io.TeeReader(r.Body, &buf) // must read tee first.
			post, _ = io.ReadAll(tee)
			appName = app.String()
		)

		r.Body.Close()              // we just read this into a buffer.
		r.Body = io.NopCloser(&buf) // someone else gets to read it now.

		// notifiarr.com uses 1-indexes; subtract 1 from the ID (turn 1 into 0 generally).
		switch aID--; {
		// Make sure the id is within range of the available service.
		case app == starr.Lidarr && (aID >= len(a.Lidarr) || aID < 0):
			msg = fmt.Errorf("%v: %w", aID, ErrNoLidarr)
		case app == starr.Prowlarr && (aID >= len(a.Prowlarr) || aID < 0):
			msg = fmt.Errorf("%v: %w", aID, ErrNoLidarr)
		case app == starr.Radarr && (aID >= len(a.Radarr) || aID < 0):
			msg = fmt.Errorf("%v: %w", aID, ErrNoRadarr)
		case app == starr.Readarr && (aID >= len(a.Readarr) || aID < 0):
			msg = fmt.Errorf("%v: %w", aID, ErrNoReadarr)
		case app == starr.Sonarr && (aID >= len(a.Sonarr) || aID < 0):
			msg = fmt.Errorf("%v: %w", aID, ErrNoSonarr)
			// Store the application configuration (starr) in a context then pass that into the api() method.
			// Retrieve the return code and output, and send a response via a.Respond().
		case app == starr.Lidarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Lidarr[aID])))
		case app == starr.Prowlarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Prowlarr[aID])))
		case app == starr.Radarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Radarr[aID])))
		case app == starr.Readarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Readarr[aID])))
		case app == starr.Sonarr:
			code, msg = api(r.WithContext(context.WithValue(ctx, app, a.Sonarr[aID])))
		case app == "":
			// no app, just run the handler.
			code, msg = api(r) // unknown app, just run the handler.
		default:
			// unknown app, add the ID to the context and run the handler.
			code, msg = api(r.WithContext(context.WithValue(ctx, app, aID)))
		}

		if len(post) > 0 {
			s, _ := json.MarshalIndent(msg, "", " ")
			a.Debugf("Incoming API: %s %s: %s\nStatus: %d, Reply: %s", r.Method, r.URL, string(post), code, s)
		}

		if appName == "" {
			appName = "Non-App"
		}

		wrote := a.Respond(w, code, msg)
		exp.APIHits.Add(appName+" Bytes Sent", wrote)
		exp.APIHits.Add(appName+" Bytes Received", int64(len(post)))
		exp.APIHits.Add(appName+" Requests", 1)
		exp.APIHits.Add("Total", 1)
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
	}
}

// CheckAPIKey drops a 403 if the API key doesn't match, otherwise run next handler.
func (a *Apps) CheckAPIKey(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		if _, ok := a.keys[r.Header.Get("X-API-Key")]; !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
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
func (a *Apps) Setup() error { //nolint:cyclop
	a.APIKey = strings.TrimSpace(a.APIKey)

	if err := a.setupLidarr(); err != nil {
		return err
	}

	if err := a.setupProwlarr(); err != nil {
		return err
	}

	if err := a.setupRadarr(); err != nil {
		return err
	}

	if err := a.setupReadarr(); err != nil {
		return err
	}

	if err := a.setupSonarr(); err != nil {
		return err
	}

	if err := a.setupDeluge(); err != nil {
		return err
	}

	if err := a.setupNZBGet(); err != nil {
		return err
	}

	if err := a.setupQbit(); err != nil {
		return err
	}

	if err := a.setupSabNZBd(); err != nil {
		return err
	}

	if err := a.setupRtorrent(); err != nil {
		return err
	}

	a.Tautulli.setup()

	return nil
}

// Respond sends a standard response to our caller. JSON encoded blobs. Returns size of data sent.
func (a *Apps) Respond(w http.ResponseWriter, stat int, msg interface{}) int64 { //nolint:varnamelen
	statusTxt := strconv.Itoa(stat) + ": " + http.StatusText(stat)

	if stat == http.StatusFound || stat == http.StatusMovedPermanently ||
		stat == http.StatusPermanentRedirect || stat == http.StatusTemporaryRedirect {
		m, _ := msg.(string)
		w.Header().Set("Location", m)
		w.WriteHeader(stat)
		exp.APIHits.Add(statusTxt, 1)

		return 0
	}

	if m, ok := msg.(error); ok {
		a.Errorf("Request failed. Status: %s, Message: %v", statusTxt, m)
		msg = m.Error()
	}

	exp.APIHits.Add(statusTxt, 1)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stat)
	counter := datacounter.NewResponseWriterCounter(w)
	json := json.NewEncoder(counter)
	json.SetEscapeHTML(false)

	err := json.Encode(map[string]interface{}{"status": statusTxt, "message": msg})
	if err != nil {
		a.Errorf("Sending JSON response failed. Status: %s, Error: %v, Message: %v", statusTxt, err, msg)
	}

	return int64(counter.Count())
}

/* Every API call runs one of these methods to find the interface for the respective app. */

func getLidarr(r *http.Request) *lidarr.Lidarr {
	app, _ := r.Context().Value(starr.Lidarr).(*LidarrConfig)
	return app.Lidarr
}

// will be used when we add http handlers for prowlarr.
/* func getProwlarr(r *http.Request) *prowlarr.Prowlarr {
	app, _ := r.Context().Value(starr.Prowlarr).(*ProwlarrConfig)
	return app.Prowlarr
} */

func getRadarr(r *http.Request) *radarr.Radarr {
	app, _ := r.Context().Value(starr.Radarr).(*RadarrConfig)
	return app.Radarr
}

func getReadarr(r *http.Request) *readarr.Readarr {
	app, _ := r.Context().Value(starr.Readarr).(*ReadarrConfig)
	return app.Readarr
}

func getSonarr(r *http.Request) *sonarr.Sonarr {
	app, _ := r.Context().Value(starr.Sonarr).(*SonarrConfig)
	return app.Sonarr
}

func metricMaker(app string) func(string, string, int, int, error) {
	return func(status, method string, sent, rcvd int, err error) {
		exp.Apps.Add(app+"&&"+method+" Bytes Received", int64(rcvd))
		exp.Apps.Add(app+"&&"+method+" Requests", 1)

		if method != "GET" || sent > 0 {
			exp.Apps.Add(app+"&&"+method+" Bytes Sent", int64(sent))
		}

		if err != nil {
			exp.Apps.Add(app+"&&"+method+" Request Errors", 1)
		} else {
			exp.Apps.Add(app+"&&"+method+" Response: "+status, 1)
		}
	}
}
