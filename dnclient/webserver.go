package dnclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/bindata"
	"github.com/gorilla/mux"
	apachelog "github.com/lestrrat-go/apache-logformat"
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

// allowedIPs determines who can set x-forwarded-for.
type allowedIPs []*net.IPNet

// StartWebServer starts the web server.
func (c *Client) StartWebServer() {
	// Create an apache-style logger.
	l, _ := apachelog.New(`%{X-Forwarded-For}i %l %u %t "%r" %>s %b "%{Referer}i" ` +
		`"%{User-agent}i" %{X-Request-Time}o %Dμs`)
	// Create a request router.
	c.router = mux.NewRouter()
	// Create a server.
	c.server = &http.Server{ // nolint: exhaustivestruct
		Handler:           l.Wrap(c.fixForwardedFor(c.router), c.Logger.Logger.Writer()),
		Addr:              c.Config.BindAddr,
		IdleTimeout:       time.Minute,
		WriteTimeout:      c.Config.Timeout.Duration,
		ReadTimeout:       c.Config.Timeout.Duration,
		ReadHeaderTimeout: c.Config.Timeout.Duration,
		ErrorLog:          c.Logger.Errors,
	}

	// Initialize all the application API paths.
	c.radarrHandlers()
	c.readarrHandlers()
	c.lidarrHandlers()
	c.sonarrHandlers()

	// Initialize "special" internal API paths.
	c.router.Handle(path.Join("/", c.Config.URLBase, "api", "status"), // does not return any data
		http.HandlerFunc(c.statusResponse)).Methods("GET", "HEAD") // does not require a key
	c.handleAPIpath("", "info", c.updateInfo, "PUT")                 // requires a key
	c.handleAPIpath("", "version", c.versionResponse, "GET", "HEAD") // requires a key

	// Initialize internal-only paths.
	c.router.Handle("/favicon.ico", http.HandlerFunc(c.favIcon))   // built-in icon.
	c.router.Handle("/", http.HandlerFunc(c.slash))                // "hi" page on /
	c.router.PathPrefix("/").Handler(http.HandlerFunc(c.notFound)) // 404 everything

	// Run the server.
	go c.runWebServer()
}

// runWebServer starts the http or https listener.
func (c *Client) runWebServer() {
	var err error

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		err = c.server.ListenAndServeTLS(c.Config.SSLCrtFile, c.Config.SSLKeyFile)
	} else {
		err = c.server.ListenAndServe()
	}

	c.server = nil

	if err != nil && !errors.Is(http.ErrServerClosed, err) {
		c.Errorf("Web Server: %v (shutting down)", err)
		c.signal <- os.Kill // stop the app.
	}
}

// StopWebServer stops the web servers. Panics if that causes an error or timeout.
func (c *Client) StopWebServer() {
	ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout.Duration)
	defer cancel()

	if err := c.server.Shutdown(ctx); err != nil {
		c.Errorf("Web Server: %v (shutting down)", err)
		c.signal <- os.Kill
	}
}

// RestartWebServer stop and starts the web server.
// Panics if that causes an error or timeout.
func (c *Client) RestartWebServer(run func()) {
	c.StopWebServer()
	defer c.StartWebServer()

	if run != nil {
		run()
	}
}

// checkAPIKey drops a 403 if the API key doesn't match.
func (c *Client) checkAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != c.Config.APIKey {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (c *Client) respond(w http.ResponseWriter, stat int, msg interface{}) {
	statusTxt := strconv.Itoa(stat) + ": " + http.StatusText(stat)

	if m, ok := msg.(error); ok {
		c.Errorf("Status: %s, Message: %v", statusTxt, m)
		msg = m.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stat)

	b, _ := json.Marshal(map[string]interface{}{"status": statusTxt, "message": msg})
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n")) // curl likes new lines.
}

func (c *Client) updateInfo(r *http.Request) (int, interface{}) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("reading POST body: %w", err)
	} else if _, ok := c.menu["dninfo"]; !ok {
		return http.StatusNotAcceptable, "menu is not active"
	}

	c.info = string(body)
	c.menu["dninfo"].Show()

	return http.StatusOK, "info updated and menu shown"
}

// versionResponse returns application run and build time data: /api/version.
func (c *Client) versionResponse(r *http.Request) (int, interface{}) {
	return http.StatusOK, map[string]interface{}{
		"version":        version.Version,
		"uptime":         time.Since(version.Started).Round(time.Second).String(),
		"uptime_seconds": time.Since(version.Started).Round(time.Second).Seconds(),
		"build_date":     version.BuildDate,
		"branch":         version.Branch,
		"go_version":     version.GoVersion,
		"revision":       version.Revision,
	}
}

// notFound is the handler for paths that are not found: 404s.
func (c *Client) notFound(w http.ResponseWriter, r *http.Request) {
	c.respond(w, http.StatusNotFound, "Check your request parameters and try again.")
}

// statusResponse is the handler for /api/status.
func (c *Client) statusResponse(w http.ResponseWriter, r *http.Request) {
	c.respond(w, http.StatusOK, c.Flags.Name()+" alive!")
}

// slash is the handler for /.
func (c *Client) slash(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("<p>" + c.Flags.Name() + ": <strong>working</strong></p>\n"))
}

func (c *Client) favIcon(w http.ResponseWriter, r *http.Request) {
	if b, err := bindata.Asset("files/favicon.ico"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		http.ServeContent(w, r, r.URL.Path, time.Now(), bytes.NewReader(b))
	}
}

// fixForwardedFor sets the X-Forwarded-For header to the client IP
// under specific circumstances.
func (c *Client) fixForwardedFor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Trim(r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")], "[]")
		if x := r.Header.Get("X-Forwarded-For"); x == "" || !c.allow.contains(ip) {
			r.Header.Set("X-Forwarded-For", ip)
		} else if l := strings.LastIndexAny(x, ", "); l != -1 {
			r.Header.Set("X-Forwarded-For", strings.Trim(x[l:len(x)-1], ", "))
		}

		next.ServeHTTP(w, r)
	})
}

var _ = fmt.Stringer(allowedIPs(nil))

func (n allowedIPs) String() (s string) {
	if len(n) < 1 {
		return "(none)"
	}

	for i := range n {
		if s != "" {
			s += ", "
		}

		s += n[i].String()
	}

	return s
}

func (n allowedIPs) contains(ip string) bool {
	for i := range n {
		if n[i].Contains(net.ParseIP(ip)) {
			return true
		}
	}

	return false
}
