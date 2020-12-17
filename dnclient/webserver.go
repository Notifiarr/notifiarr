package dnclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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
)

// RunWebServer starts the web server.
func (c *Client) RunWebServer() {
	r := mux.NewRouter()
	// Generic
	r.PathPrefix("/").Handler(c.responseWrapper(c.notFound))
	r.Handle("/api/status", c.responseWrapper(c.statusResponse)).Methods("GET", "HEAD")

	// Radarr
	r.Handle("/api/radarr/add/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.radarrAddMovie))).Methods("POST")
	r.Handle("/api/radarr/qualityProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.radarrProfiles))).Methods("GET")
	r.Handle("/api/radarr/rootFolder/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.radarrRootFolders))).Methods("GET")

	// Readarr
	r.Handle("/api/readarr/add/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.readarrAddBook))).Methods("POST")
	r.Handle("/api/readarr/metadataProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.readarrMetaProfiles))).Methods("GET")
	r.Handle("/api/readarr/qualityProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.readarrProfiles))).Methods("GET")
	r.Handle("/api/readarr/rootFolder/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.readarrRootFolders))).Methods("GET")

	// Sonarr
	r.Handle("/api/sonarr/add/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.sonarrAddSeries))).Methods("POST")
	r.Handle("/api/sonarr/qualityProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.sonarrProfiles))).Methods("GET")
	r.Handle("/api/sonarr/languageProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.sonarrLangProfiles))).Methods("GET")
	r.Handle("/api/sonarr/rootFolder/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.sonarrRootFolders))).Methods("GET")

	// Lidarr
	r.Handle("/api/lidarr/add/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.lidarrAddAlbum))).Methods("POST")
	r.Handle("/api/lidarr/qualityProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.lidarrProfiles))).Methods("GET")
	r.Handle("/api/lidarr/qualityProfiles/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.lidarrQualityDefs))).Methods("GET")
	r.Handle("/api/lidarr/rootFolder/{id:[0-9]+}",
		c.checkAPIKey(c.responseWrapper(c.lidarrRootFolders))).Methods("GET")

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

// checkAPIKey drops a 403 if the API key doesn't match.
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

// Response formats all content-containing replies to client web requests.
type Response struct {
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
}

// responseWrapper formats all content-containing replies to clients.
func (c *Client) responseWrapper(next func(r *http.Request) (int, interface{})) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stat, msg := next(r)
		statusTxt := strconv.Itoa(stat) + ": " + http.StatusText(stat)

		if m, ok := msg.(error); ok {
			msg = m.Error()
		}

		if s, ok := msg.(string); ok {
			c.Printf("HTTP [%s] %s %s: %s: %s", r.RemoteAddr, r.Method, r.RequestURI, statusTxt, s)
		} else {
			c.Printf("HTTP [%s] %s %s: %s", r.RemoteAddr, r.Method, r.RequestURI, statusTxt)
		}

		w.WriteHeader(stat)

		b, _ := json.Marshal(&Response{Status: statusTxt, Message: msg})
		_, _ = w.Write(b)
		_, _ = w.Write([]byte("\n")) // curl likes new lines.
	})
}

// notFound is the handler for paths that are not found: 404s.
func (c *Client) statusResponse(r *http.Request) (int, interface{}) {
	return http.StatusOK, "I'm alive!"
}

// notFound is the handler for paths that are not found: 404s.
func (c *Client) notFound(r *http.Request) (int, interface{}) {
	return http.StatusNotFound, "The page you requested could not be found. Check your request parameters and try again."
}
