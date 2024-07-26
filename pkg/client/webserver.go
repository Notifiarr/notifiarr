package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	apachelog "github.com/lestrrat-go/apache-logformat/v2"
)

// StartWebServer starts the web server.
func (c *Client) StartWebServer(ctx context.Context) {
	c.Lock()
	defer c.Unlock()

	//nolint:lll // Create an apache-style logger.
	apache, _ := apachelog.New(`%{X-Forwarded-For}i - %{X-NotiClient-Username}i %t "%m %{X-Redacted-URI}i %H" %>s %b "%{Referer}i" "%{User-agent}i" %{X-Request-Time}i %{ms}Tms`)

	// Create a request router.
	c.Config.Router = mux.NewRouter()
	c.Config.Router.Use(c.fixForwardedFor)
	c.Config.Router.Use(c.countRequest)
	c.Config.Router.Use(c.addUsernameHeader)
	c.webauth = c.Config.UIPassword.Webauth() // this needs to be locked since password can be changed without reloading.
	c.noauth = c.Config.UIPassword.Noauth()
	c.authHeader = c.Config.UIPassword.Header()

	// Make a multiplexer because websockets can't use apache log.
	smx := http.NewServeMux()
	smx.Handle("/", c.stripSecrets(apache.Wrap(c.Config.Router, c.Logger.HTTPLog.Writer())))
	smx.Handle(path.Join(c.Config.URLBase, "ui", "ws"), c.Config.Router) // websockets cannot go through the apache logger.

	// Create a server.
	c.server = &http.Server{ //nolint: exhaustivestruct
		Handler:           smx,
		Addr:              c.Config.BindAddr,
		IdleTimeout:       mnd.DefaultTimeout,
		WriteTimeout:      c.Config.Timeout.Duration,
		ReadTimeout:       c.Config.Timeout.Duration,
		ReadHeaderTimeout: c.Config.Timeout.Duration,
		ErrorLog:          c.Logger.ErrorLog,
	}

	// Start the Notifiarr.com origin websocket tunnel.
	c.startTunnel(ctx)
	// Initialize all the application API paths.
	c.Config.Apps.InitHandlers()
	c.httpHandlers()
	// Run the server.
	go c.runWebServer()
}

// runWebServer starts the http or https listener.
func (c *Client) runWebServer() {
	defer c.CapturePanic()

	var err error

	if menu["stat"] != nil {
		menu["stat"].Enable()
		menu["stat"].Check()
		menu["stat"].SetTooltip("web server running, uncheck to pause")
	}

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		err = c.server.ListenAndServeTLS(c.Config.SSLCrtFile, c.Config.SSLKeyFile)
	} else {
		err = c.server.ListenAndServe()
	}

	if err != nil && !errors.Is(http.ErrServerClosed, err) {
		c.Errorf("Web Server Failed: %v (shutting down)", err)
		c.sigkil <- os.Kill // stop the app.
	}
}

// StopWebServer stops the web servers. Panics if that causes an error or timeout.
func (c *Client) StopWebServer(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()

	if c.server == nil {
		return nil
	}

	c.Print("==> Stopping Web Server!")

	ctx, cancel := context.WithTimeout(ctx, c.Config.Timeout.Duration)
	defer cancel()

	if menu["stat"] != nil {
		menu["stat"].Uncheck()
		menu["stat"].SetTooltip("web server paused, click to start")
	}

	if c.tunnel != nil {
		defer c.tunnel.Shutdown()
	}
	// Wait for any active requests before shutting down the tunnel.
	if err := c.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down web server: %w", err)
	}

	return nil
}

/* Wrap all incoming http calls, so we can stuff counters into expvar. */

var (
	_ = http.ResponseWriter(&responseWrapper{})
	_ = net.Conn(&netConnWrapper{})
)

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

type netConnWrapper struct {
	net.Conn
}

func (r *responseWrapper) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseWrapper) Write(b []byte) (int, error) {
	mnd.HTTPRequests.Add("Response Bytes", int64(len(b)))
	return r.ResponseWriter.Write(b) //nolint:wrapcheck
}

func (r *responseWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijack, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		// This fires if you move the /ui/ws endpoint to another name.
		// It needs to be updated in two places.
		panic("cannot hijack connection!")
	}

	conn, buf, err := hijack.Hijack()
	if err != nil {
		return conn, buf, err //nolint:wrapcheck
	}

	return &netConnWrapper{conn}, buf, nil
}

func (n *netConnWrapper) Write(b []byte) (int, error) {
	mnd.HTTPRequests.Add("Response Bytes", int64(len(b)))
	return n.Conn.Write(b) //nolint:wrapcheck
}
