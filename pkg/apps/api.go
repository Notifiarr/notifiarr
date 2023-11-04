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
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	"golift.io/datacounter"
	"golift.io/starr"
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
func (a *Apps) handleAPI(app starr.App, api APIHandler) http.HandlerFunc { //nolint:cyclop,funlen
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
			msg = fmt.Errorf("%v: %w", aID, ErrNoProwlarr)
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

		wrote := a.Respond(w, code, msg)

		if str, _ := json.MarshalIndent(msg, "", " "); len(post) > 0 {
			a.Debugf("Incoming API: %s %s (%s): %s\nStatus: %d, Reply (%s): %s",
				r.Method, r.URL, mnd.FormatBytes(len(post)), string(post), code, mnd.FormatBytes(wrote), str)
		} else {
			a.Debugf("Incoming API: %s %s, Status: %d, Reply (%s): %s", r.Method, r.URL, code, mnd.FormatBytes(wrote), str)
		}

		if appName == "" {
			appName = "Non-App"
		}

		mnd.APIHits.Add(appName+" Bytes Sent", wrote)
		mnd.APIHits.Add(appName+" Bytes Received", int64(len(post)))
		mnd.APIHits.Add(appName+" Requests", 1)
		mnd.APIHits.Add("Total", 1)
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

// Respond sends a standard response to our caller. JSON encoded blobs. Returns size of data sent.
func (a *Apps) Respond(w http.ResponseWriter, stat int, msg interface{}) int64 { //nolint:varnamelen
	statusTxt := strconv.Itoa(stat) + ": " + http.StatusText(stat)

	if stat == http.StatusFound || stat == http.StatusMovedPermanently ||
		stat == http.StatusPermanentRedirect || stat == http.StatusTemporaryRedirect {
		m, _ := msg.(string)
		w.Header().Set("Location", m)
		w.WriteHeader(stat)
		mnd.APIHits.Add(statusTxt, 1)

		return 0
	}

	if m, ok := msg.(error); ok {
		a.Errorf("Request failed. Status: %s, Message: %v", statusTxt, m)
		msg = m.Error()
	}

	//
	type apiResponse struct {
		// The status always matches the HTTP response.
		Status string `json:"status"`
		// This message contains the request-specific response payload.
		Msg interface{} `json:"message"`
	}

	mnd.APIHits.Add(statusTxt, 1)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stat)
	counter := datacounter.NewResponseWriterCounter(w)
	json := json.NewEncoder(counter)
	json.SetEscapeHTML(false)

	err := json.Encode(&apiResponse{Status: statusTxt, Msg: msg})
	if err != nil {
		a.Errorf("Sending JSON response failed. Status: %s, Error: %v, Message: %v", statusTxt, err, msg)
	}

	return int64(counter.Count())
}
