package client

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	apachelog "github.com/lestrrat-go/apache-logformat/v2"
	mulery "golift.io/mulery/client"
)

const (
	// maximum websocket connections to the origin (mulery server).
	maxPoolSize = 20
	// maximum is calculated, and this is the minimum it may be.
	maxPoolMin = 4
)

// poolMax returns a reasonable number for max tunnel connections.
// This basically dictates how many parallel requests the website
// may send to this client. Realistically very few clients create
// more than 4 or 5 connections.
func (c *Client) poolMax(ci *clientinfo.ClientInfo) int {
	poolmax := len(c.Config.Apps.Sonarr) + len(c.Config.Apps.Radarr) + len(c.Config.Apps.Lidarr) +
		len(c.Config.Apps.Readarr) + len(c.Config.Apps.Prowlarr) + len(c.Config.Apps.Deluge) +
		len(c.Config.Apps.Qbit) + len(c.Config.Apps.Rtorrent) + len(c.Config.Apps.SabNZB) +
		len(c.Config.Apps.NZBGet) + 1

	if c.Config.Apps.Plex.Enabled() {
		poolmax++
	}

	if c.Config.Apps.Tautulli.Enabled() {
		poolmax++
	}

	if poolmax > maxPoolSize || ci.IsSub() {
		poolmax = maxPoolSize
	} else if poolmax < maxPoolMin {
		poolmax = maxPoolMin
	}

	return poolmax
}

func (c *Client) startTunnel(ctx context.Context) {
	// If clientinfo is nil, then we probably have a bad API key.
	ci := clientinfo.Get()
	if ci == nil {
		c.Errorf("Skipping tunnel creation because there is no client info.")
		return
	}

	hostname, _ := os.Hostname()
	if hostInfo, err := c.clientinfo.GetHostInfo(ctx); err != nil {
		hostname = hostInfo.Hostname
	}

	// This apache logger is only used for client->server websocket-tunneled requests.
	remWs, _ := apachelog.New(`%{X-Forwarded-For}i %{X-User-ID}i env:%{X-User-Environment}i %t "%r" %>s %b ` +
		`"%{X-Client-ID}i" "%{User-agent}i" %{X-Request-Time}i %{ms}Tms`)

	//nolint:gomnd // just attempting a tiny bit of splay.
	c.tunnel = mulery.NewClient(&mulery.Config{
		Name:          hostname,
		ID:            c.Config.HostID,
		ClientIDs:     []any{ci.User.ID},
		Targets:       ci.User.Tunnels,
		PoolIdleSize:  1,
		PoolMaxSize:   c.poolMax(ci),
		CleanInterval: time.Second + time.Duration(c.triggers.Timers.Rand().Intn(1000))*time.Millisecond,
		Backoff:       600*time.Millisecond + time.Duration(c.triggers.Timers.Rand().Intn(600))*time.Millisecond,
		SecretKey:     c.Config.APIKey,
		Handler:       remWs.Wrap(c.prefixURLbase(c.Config.Router), c.Logger.HTTPLog.Writer()).ServeHTTP,
		Logger: &tunnelLogger{
			Logger:         c.Logger,
			sendSiteErrors: ci.User.DevAllowed,
		},
	})

	c.Printf("Tunneling to %q with %d connections; cleaner:%s, backoff:%s, url: %s, hash: %s",
		strings.Join(c.tunnel.Targets, ", "), c.tunnel.PoolMaxSize, c.tunnel.CleanInterval,
		c.tunnel.Backoff, ci.User.TunnelURL, c.tunnel.GetID())
	c.tunnel.Start(ctx)
}

// prefixURLbase adds a prefix to an http request.
// We need this to fix websocket-tunneled requests
// from the website when url base is not the default.
func (c *Client) prefixURLbase(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.Config.URLBase == "" || c.Config.URLBase == "/" {
			h.ServeHTTP(w, r)
			return
		}

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = path.Join(c.Config.URLBase, r.URL.Path)

		if r.URL.RawPath != "" {
			r2.URL.RawPath = path.Join(c.Config.URLBase, r.URL.RawPath)
		}

		h.ServeHTTP(w, r2)
	})
}

// tunnelLogger lets us tune the logs from the mulery tunnel.
type tunnelLogger struct {
	mnd.Logger
	// sendSiteErrors true sends tunnel errors to website as notifications.
	sendSiteErrors bool
}

// Debugf prints a message with DEBUG prefixed.
func (l *tunnelLogger) Debugf(format string, v ...interface{}) {
	l.Logger.Debugf(format, v...)
}

// Errorf prints a message with ERROR prefixed.
func (l *tunnelLogger) Errorf(format string, v ...interface{}) {
	// this is why we dont just pass the interface in as-is.
	if l.sendSiteErrors {
		l.Logger.Errorf(format, v...)
	} else {
		l.Logger.ErrorfNoShare(format, v...)
	}
}

// Printf prints a message with INFO prefixed.
func (l *tunnelLogger) Printf(format string, v ...interface{}) {
	l.Logger.Printf(format, v...)
}
