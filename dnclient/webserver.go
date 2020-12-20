package dnclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
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

	c.radarrMethods(r)
	c.readarrMethods(r)
	c.lidarrMethods(r)
	c.sonarrMethods(r)
	r.Handle("/favicon.ico", http.HandlerFunc(c.favIcon))
	r.Handle(path.Join("/", c.Config.WebRoot, "/api/status"), c.responseWrapper(c.statusResponse)).Methods("GET", "HEAD")
	r.PathPrefix("/").Handler(c.responseWrapper(c.notFound))

	c.server = &http.Server{ // nolint: exhaustivestruct
		Handler:      r,
		Addr:         c.Config.BindAddr,
		IdleTimeout:  time.Second,
		WriteTimeout: time.Second,
		ReadTimeout:  time.Second,
		ErrorLog:     c.Logger.Logger,
	}

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		err := c.server.ListenAndServeTLS(c.Config.SSLCrtFile, c.Config.SSLKeyFile)
		if err != nil && !errors.Is(http.ErrServerClosed, err) {
			c.Printf("[ERROR] HTTPS Server: %v", err)
		}
	} else if err := c.server.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
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

		w.Header().Set("Content-Type", "application/json")
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

func (c *Client) favIcon(w http.ResponseWriter, r *http.Request) {
	if b, err := Asset("init/windows/application.ico"); err != nil {
		statusTxt := strconv.Itoa(http.StatusInternalServerError) + ": " + http.StatusText(http.StatusInternalServerError)
		c.Printf("HTTP [%s] %s %s: %s: %v", r.RemoteAddr, r.Method, r.RequestURI, statusTxt, err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		statusTxt := strconv.Itoa(http.StatusOK) + ": " + http.StatusText(http.StatusOK)
		c.Printf("HTTP [%s] %s %s: %s", r.RemoteAddr, r.Method, r.RequestURI, statusTxt)
		http.ServeContent(w, r, r.URL.Path, time.Now(), bytes.NewReader(b))
	}
}
