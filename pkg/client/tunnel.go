package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/gorilla/schema"
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
func (c *Client) poolMax(info *clientinfo.ClientInfo) int {
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

	if poolmax > maxPoolSize || info.IsSub() {
		poolmax = maxPoolSize
	} else if poolmax < maxPoolMin {
		poolmax = maxPoolMin
	}

	return poolmax
}

func (c *Client) startTunnel(ctx context.Context) {
	// If clientinfo is nil, then we probably have a bad API key.
	info := clientinfo.Get()
	if info == nil {
		c.Errorf("Skipping tunnel creation because there is no client info.")
		return
	}

	c.makeTunnel(ctx, info)
	c.Printf("Tunneling to %d targets with %d connections; cleaner:%s, backoff:%s, url: %s, hash: %s",
		len(c.tunnel.Targets), c.tunnel.PoolMaxSize, c.tunnel.CleanInterval,
		c.tunnel.Backoff, info.User.TunnelURL, c.tunnel.GetID())
	c.Printf("Tunnel Targets: %s", strings.Join(c.tunnel.Targets, ", "))
	c.tunnel.Start(ctx)
}

func (c *Client) makeTunnel(ctx context.Context, info *clientinfo.ClientInfo) {
	hostname, _ := os.Hostname()
	if hostInfo, err := c.triggers.CI.GetHostInfo(ctx); err != nil {
		hostname = hostInfo.Hostname
	}

	// This apache logger is only used for client->server websocket-tunneled requests.
	remWs, _ := apachelog.New(`%{X-Forwarded-For}i %{X-User-ID}i env:%{X-User-Environment}i %t "%r" %>s %b ` +
		`"%{X-Client-ID}i" "%{User-agent}i" %{X-Request-Time}i %{ms}Tms`)

	//nolint:mnd // just attempting a tiny bit of splay.
	c.tunnel = mulery.NewClient(&mulery.Config{
		Name:             hostname,
		ID:               c.Config.HostID,
		ClientIDs:        []any{info.User.ID},
		Targets:          getTunnels(info),
		PoolIdleSize:     1,
		PoolMaxSize:      c.poolMax(info),
		CleanInterval:    time.Second + time.Duration(c.triggers.Rand().Intn(1000))*time.Millisecond,
		Backoff:          600*time.Millisecond + time.Duration(c.triggers.Rand().Intn(600))*time.Millisecond,
		SecretKey:        c.Config.APIKey,
		Handler:          remWs.Wrap(c.prefixURLbase(c.Config.Router), c.Logger.HTTPLog.Writer()).ServeHTTP,
		RoundRobinConfig: c.roundRobinConfig(info),
		Logger: &tunnelLogger{
			ctx:            ctx,
			Logger:         c.Logger,
			sendSiteErrors: info.User.DevAllowed,
		},
	})
}

//nolint:mnd // arbitrary failover time frames.
func (c *Client) roundRobinConfig(ci *clientinfo.ClientInfo) *mulery.RoundRobinConfig {
	interval := 10 * time.Minute
	if ci.IsSub() {
		interval = 2 * time.Minute
	} else if ci.IsPatron() || ci.User.DevAllowed {
		interval = 5 * time.Minute
	}

	return &mulery.RoundRobinConfig{
		RetryInterval: interval,
		Callback: func(_ context.Context, socket string) {
			defer data.Save("activeTunnel", socket)
			// Tell the website we connected to a new tunnel, so it knows how to reach us.
			c.Config.SendData(&website.Request{
				Route:      website.TunnelRoute,
				Event:      website.EventSignal,
				Payload:    map[string]interface{}{"socket": socket, "previous": data.Get("activeTunnel")},
				LogMsg:     fmt.Sprintf("Update Tunnel Target (%s)", socket),
				LogPayload: true,
			})
		},
	}
}

// getTunnels returns a list of tunnels the client will round robin.
func getTunnels(info *clientinfo.ClientInfo) []string {
	// If the user has already selected their preferred tunnels, use them.
	if len(info.User.Tunnels) > 1 {
		return info.User.Tunnels
	}

	// The above is the new way, the below is the 'transition' way.
	// The above allows the user to pick 2 or 3 tunnels. If they haven't
	// picked anything yet, then they get all of them (below) until they do
	// pick some. The below code probably can't be removed, so a client
	// can bootstrap with no configuration present on the website.

	// Otherwise, use the legacy selection and append all tunnels.
	tunnels := []string{}
	if len(info.User.Tunnels) != 0 {
		tunnels = append(tunnels, info.User.Tunnels[0])
	}

	for _, item := range info.User.Mulery {
		if len(info.User.Tunnels) == 0 || item.Socket != info.User.Tunnels[0] {
			tunnels = append(tunnels, item.Socket)
		}
	}

	return tunnels
}

// prefixURLbase adds a prefix to an http request.
// We need this to fix websocket-tunneled requests
// from the website when url base is not the default.
func (c *Client) prefixURLbase(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if c.Config.URLBase == "" || c.Config.URLBase == "/" {
			handler.ServeHTTP(writer, req)
			return
		}

		req2 := new(http.Request)
		*req2 = *req
		req2.URL = new(url.URL)
		*req2.URL = *req.URL
		req2.URL.Path = path.Join(c.Config.URLBase, req.URL.Path)

		if req.URL.RawPath != "" {
			req2.URL.RawPath = path.Join(c.Config.URLBase, req.URL.RawPath)
		}

		handler.ServeHTTP(writer, req2)
	})
}

// tunnelLogger lets us tune the logs from the mulery tunnel.
type tunnelLogger struct {
	// hide the app context here so we can use it when we restart a tunnel from an http request
	ctx context.Context //nolint:containedctx
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

const pingTimeout = 7 * time.Second

// pingTunnels is a gui request to check timing to each tunnel.
func (c *Client) pingTunnels(response http.ResponseWriter, request *http.Request) {
	info := clientinfo.Get()
	if info == nil {
		http.Error(response, "no client info, cannot ping tunnels", http.StatusInternalServerError)
		return
	}

	var (
		wait sync.WaitGroup
		list = make(map[int]string)
		inCh = make(chan map[int]string)
	)

	defer close(inCh)

	go func() {
		for data := range inCh {
			for k, v := range data {
				list[k] = v
			}

			wait.Done()
		}
	}()

	for idx, tunnel := range info.User.Mulery {
		wait.Add(1)
		time.Sleep(70 * time.Millisecond) //nolint:mnd

		go c.pingTunnel(request.Context(), idx, tunnel.Socket, inCh)
	}

	wait.Wait()

	if err := json.NewEncoder(response).Encode(list); err != nil {
		c.Errorf("Pinging Tunnel: encoding json: %v", err)
	}
}

func (c *Client) pingTunnel(ctx context.Context, idx int, socket string, inCh chan map[int]string) {
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.Replace(socket, "wss://", "https://", 1), nil)
	if err != nil {
		c.Errorf("Pinging Tunnel: creating request: %v", err)
		return
	}

	start := time.Now()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Errorf("Pinging Tunnel: making request: %v", err)
		inCh <- map[int]string{idx: "error"}

		return
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)
	inCh <- map[int]string{idx: time.Since(start).Round(time.Millisecond).String()}
}

func (c *Client) saveTunnels(response http.ResponseWriter, request *http.Request) {
	body, _ := io.ReadAll(request.Body)

	type tunnelS struct {
		PrimaryTunnel string
		BackupTunnel  []string
	}

	var input tunnelS

	decodedValue, err := url.ParseQuery(string(body))
	if err != nil {
		c.Errorf("Saving Tunnel: parsing request: %v", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)

		return
	}

	err = schema.NewDecoder().Decode(&input, decodedValue)
	if err != nil {
		c.Errorf("Saving Tunnel: decoding request: %v", err)
		http.Error(response, err.Error(), http.StatusInternalServerError)

		return
	}

	sockets := []string{input.PrimaryTunnel}

	for _, socket := range input.BackupTunnel {
		if socket != input.PrimaryTunnel {
			sockets = append(sockets, socket)
		}
	}

	c.Config.SendData(&website.Request{
		Route:      website.TunnelRoute,
		Event:      website.EventGUI,
		Payload:    map[string]any{"sockets": sockets},
		LogMsg:     "Update Tunnel Config",
		LogPayload: true,
	})

	ci := clientinfo.Get()
	ci.User.Tunnels = sockets // pass different data to makeTunnels().
	tl, _ := c.tunnel.Config.Logger.(*tunnelLogger)

	c.tunnel.Shutdown()
	c.makeTunnel(tl.ctx, ci) //nolint:contextcheck // these cannot be inherited from the http request.
	c.tunnel.Start(tl.ctx)   //nolint:contextcheck
	http.Error(response, fmt.Sprintf("saved tunnel config. primary: %s, %d backups",
		input.PrimaryTunnel, len(input.BackupTunnel)), http.StatusOK)
}
