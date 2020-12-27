package dnclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	apachelog "github.com/lestrrat-go/apache-logformat"
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

// StartWebServer starts the web server.
func (c *Client) StartWebServer() {
	// Create an apache-style logger.
	l, _ := apachelog.New(`%{X-Forwarded-For}i %l %u %t "%r" %>s %b "%{Referer}i" ` +
		`"%{User-agent}i" %{X-Request-Time}o %Dμs`)
	// Create a request router.
	c.router = mux.NewRouter()
	// Create a server.
	c.server = &http.Server{ // nolint: exhaustivestruct
		Handler:           l.Wrap(c.fixForwardedFor(c.router), c.Logger.Requests.Writer()),
		Addr:              c.Config.BindAddr,
		IdleTimeout:       time.Minute,
		WriteTimeout:      c.Config.Timeout.Duration,
		ReadTimeout:       c.Config.Timeout.Duration,
		ReadHeaderTimeout: c.Config.Timeout.Duration,
		ErrorLog:          c.Logger.Logger,
	}

	// Initialize all the application API paths.
	c.radarrHandlers()
	c.readarrHandlers()
	c.lidarrHandlers()
	c.sonarrHandlers()
	c.internalHandlers()
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
		c.Errorf("Web Server Failed: %v (shutting down)", err)
		c.signal <- os.Kill // stop the app.
	}
}

// StopWebServer stops the web servers. Panics if that causes an error or timeout.
func (c *Client) StopWebServer() {
	ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout.Duration)
	defer cancel()

	if err := c.server.Shutdown(ctx); err != nil {
		c.Errorf("Stopping Web Server: %v (shutting down)", err)
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
