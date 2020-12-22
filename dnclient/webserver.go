package dnclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"golift.io/version"
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

// apiHandle is our custom handler function for APIs.
type apiHandle func(r *http.Request) (int, interface{})

// RunWebServer starts the web server.
func (c *Client) RunWebServer() {
	// Create a request router.
	c.router = mux.NewRouter()
	// Create a server.
	c.server = &http.Server{ // nolint: exhaustivestruct
		Handler:      c.router,
		Addr:         c.Config.BindAddr,
		IdleTimeout:  c.Config.Timeout.Duration,
		WriteTimeout: c.Config.Timeout.Duration,
		ReadTimeout:  c.Config.Timeout.Duration,
		ErrorLog:     c.Logger.Logger,
	}

	// Initialize all the application API paths.
	c.radarrHandlers()
	c.readarrHandlers()
	c.lidarrHandlers()
	c.sonarrHandlers()

	// Initialize "special" internal API paths.
	c.router.Handle(path.Join("/", c.Config.URLBase, "api", "status"), // does not return any data
		c.responseWrapper(c.statusResponse)).Methods("GET", "HEAD") // does not require a key
	c.handleAPIpath("", "version", c.versionResponse, "GET", "HEAD") // requires a key

	// Initialize internal-only paths.
	c.router.Handle("/favicon.ico", http.HandlerFunc(c.favIcon))    // built-in icon.
	c.router.Handle("/", http.HandlerFunc(c.slash))                 // "hi" page on /
	c.router.PathPrefix("/").Handler(c.responseWrapper(c.notFound)) // 404 everything

	// Run the server.
	go c.runWebServer()
}

// runWebServer starts the http or https listener.
func (c *Client) runWebServer() {
	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		err := c.server.ListenAndServeTLS(c.Config.SSLCrtFile, c.Config.SSLKeyFile)
		if err != nil && !errors.Is(http.ErrServerClosed, err) {
			c.Printf("[ERROR] HTTPS Server: %v (shutting down)", err)
		}
	} else if err := c.server.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
		c.Printf("[ERROR] HTTP Server: %v (shutting down)", err)
	}

	c.server = nil
	c.signal <- os.Kill // stop the app.
}

// checkAPIKey drops a 403 if the API key doesn't match.
func (c *Client) checkAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != c.Config.APIKey {
			c.Printf("HTTP [%s] %s %s: %d: Unauthorized: bad API key",
				r.RemoteAddr, r.Method, r.RequestURI, http.StatusUnauthorized)
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
func (c *Client) responseWrapper(next apiHandle) http.Handler {
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

// versionResponse returns application run and build time data.
func (c *Client) versionResponse(r *http.Request) (int, interface{}) {
	return http.StatusOK, struct {
		V string  `json:"version"`
		U string  `json:"uptime"`
		S float64 `json:"uptime_seconds"`
		D string  `json:"build_date"`
		B string  `json:"branch"`
		G string  `json:"go_version"`
		R string  `json:"revision"`
	}{
		version.Version,
		time.Since(version.Started).Round(time.Second).String(),
		time.Since(version.Started).Round(time.Second).Seconds(),
		version.BuildDate,
		version.Branch,
		version.GoVersion,
		version.Revision,
	}
}

// notFound is the handler for paths that are not found: 404s.
func (c *Client) notFound(r *http.Request) (int, interface{}) {
	return http.StatusNotFound, "Check your request parameters and try again."
}

// slash is the handler for /.
func (c *Client) slash(w http.ResponseWriter, r *http.Request) {
	msg := "<p>" + c.Flags.Name() + ": <strong>working</strong></p>\n"
	c.Printf("HTTP [%s] %s %s: OK: %s", r.RemoteAddr, r.Method, r.RequestURI, msg)
	_, _ = w.Write([]byte(msg))
}

func (c *Client) favIcon(w http.ResponseWriter, r *http.Request) {
	if b, err := Asset("init/windows/application.ico"); err != nil {
		c.Printf("HTTP [%s] %s %s: 500: Internal Server Error: %v", r.RemoteAddr, r.Method, r.RequestURI, err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		c.Printf("HTTP [%s] %s %s: 200 OK", r.RemoteAddr, r.Method, r.RequestURI)
		http.ServeContent(w, r, r.URL.Path, time.Now(), bytes.NewReader(b))
	}
}
