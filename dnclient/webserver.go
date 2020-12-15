package dnclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Errors sent to client web requests.
var (
	ErrNoTMDB    = fmt.Errorf("TMDB ID must not be empty")
	ErrNoGRID    = fmt.Errorf("GRID ID must not be empty")
	ErrNoTVDB    = fmt.Errorf("TVDB ID must not be empty")
	ErrNoRadarr  = fmt.Errorf("configured radarr ID not found")
	ErrNoSonarr  = fmt.Errorf("configured sonarr ID not found")
	ErrNoLidarr  = fmt.Errorf("configured lidarr ID not found")
	ErrNoReadarr = fmt.Errorf("configured readarr ID not found")
	ErrExists    = fmt.Errorf("the requested item already exists")
	ErrOnlyPOST  = fmt.Errorf("only POST is allowed to this endpoint")
	ErrOnlyGET   = fmt.Errorf("only GET is allowed to this endpoint")
)

// RunWebServer starts the web server.
func (c *Client) RunWebServer() {
	r := mux.NewRouter()
	r.Handle("/api/radarr/add/{id:[0-9]+}", c.checkAPIKey(c.responseWrapper(c.radarrAddMovie))).Methods("POST")
	r.Handle("/api/radarr/profile/{id:[0-9]+}", c.checkAPIKey(c.responseWrapper(c.radarrProfiles))).Methods("GET")
	// r.Handle("/api/sonarr/add/{id:[0-9]+}", c.checkAPIKey(c.responseWrapper(c.sonarrAddSeries))).Methods("POST")
	// r.Handle("/api/readarr/add/{id:[0-9]+}", c.checkAPIKey(c.responseWrapper(c.readarrAddBook))).Methods("POST")
	r.PathPrefix("/").Handler(http.HandlerFunc(c.notFound))

	c.server = &http.Server{
		Handler:      r,
		Addr:         c.Config.BindAddr,
		IdleTimeout:  time.Second,
		WriteTimeout: time.Second,
		ReadTimeout:  time.Second,
		ErrorLog:     c.Logger.Logger,
	}
	if err := c.server.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
		c.Printf("[ERROR] HTTP Server: %v", err)
	}
}

func (c *Client) checkAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != c.APIKey {
			c.Printf("HTTP [%s] %s %s: %s", r.RemoteAddr, r.Method, r.RequestURI, http.StatusText(http.StatusUnauthorized))
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (c *Client) responseWrapper(next func(r *http.Request) (int, string)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stat, msg := next(r)
		statusTxt := http.StatusText(stat)

		w.WriteHeader(stat)

		if len(msg) > 0 && (msg[0] == '{' || msg[0] == '[') {
			c.Printf("HTTP [%s] %s %s: %s", r.RemoteAddr, r.Method, r.RequestURI, statusTxt)
		} else {
			c.Printf("HTTP [%s] %s %s: %s: %s", r.RemoteAddr, r.Method, r.RequestURI, statusTxt, msg)
		}

		b, _ := json.Marshal(&Response{Status: statusTxt, Message: msg})
		_, _ = w.Write(b)
	})
}

// notFound is the handler for paths that are not found: 404s.
func (c *Client) notFound(w http.ResponseWriter, r *http.Request) {
	c.Printf("HTTP [%s] %s %s: %s", r.RemoteAddr, r.Method, r.RequestURI, http.StatusText(http.StatusNotFound))
	w.WriteHeader(http.StatusNotFound)
}
