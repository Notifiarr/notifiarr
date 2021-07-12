package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"golift.io/version"
)

// internalHandlers initializes "special" internal API paths.
func (c *Client) internalHandlers() {
	c.Config.HandleAPIpath("", "slow", c.slowResponse, "HEAD") // log testing
	c.Config.HandleAPIpath("", "status", c.statusResponse, "GET", "HEAD")
	c.Config.HandleAPIpath("", "version", c.versionResponse, "GET", "HEAD")
	c.Config.HandleAPIpath("", "info", c.updateInfo, "PUT")
	c.Config.HandleAPIpath("", "info/alert", c.updateInfoAlert, "PUT")

	if c.Config.Plex.Configured() {
		c.Config.HandleAPIpath(plex.Plex, "sessions", c.Config.Plex.HandleSessions, "GET")
		c.Config.HandleAPIpath(plex.Plex, "kill", c.Config.Plex.HandleKillSession, "GET").
			Queries("reason", "{reason:.*}", "sessionId", "{sessionId:[0-9a-z-]+}")

		tokens := fmt.Sprintf("{token:%s|%s}", c.Config.Plex.Token, c.Config.Apps.APIKey)
		c.Config.Router.Handle("/plex",
			http.HandlerFunc(c.plexIncoming)).Methods("POST").Queries("token", tokens)
	}

	// Initialize internal-only paths.
	c.Config.Router.Handle("/favicon.ico", http.HandlerFunc(c.favIcon))   // built-in icon.
	c.Config.Router.Handle("/", http.HandlerFunc(c.slash))                // "hi" page on /
	c.Config.Router.PathPrefix("/").Handler(http.HandlerFunc(c.notFound)) // 404 everything
}

func (c *Client) updateInfoAny(r *http.Request) (int, string, interface{}) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("reading PUT body: %w", err)
	}

	c.Print("New Info from Notifiarr.com:", string(body))

	if _, ok := c.menu["dninfo"]; !ok {
		return http.StatusAccepted, "", "menu UI is not active"
	}

	c.info = string(body) // not lockd, prob should be.
	c.menu["dninfo"].Show()

	return http.StatusOK, string(body), "info updated and menu shown"
}

func (c *Client) updateInfo(r *http.Request) (int, interface{}) {
	code, _, err := c.updateInfoAny(r)

	return code, err
}

// updateInfoAlert is the same as updateInfo except it adds a popup window.
func (c *Client) updateInfoAlert(r *http.Request) (int, interface{}) {
	code, body, err := c.updateInfoAny(r)
	if body == "" {
		return code, err
	}

	if c.alert.Active() {
		return http.StatusLocked, "previous alert not acknowledged"
	}

	go func() {
		_, _ = ui.Warning(mnd.Title+" Alert", body)
		c.alert.Done() //nolint:wsl
	}()

	return code, err
}

type appStatus struct {
	Radarr  []*conTest `json:"radarr"`
	Readarr []*conTest `json:"readarr"`
	Sonarr  []*conTest `json:"sonarr"`
	Lidarr  []*conTest `json:"lidarr"`
	Plex    []*conTest `json:"plex"`
}

type conTest struct {
	Instance int         `json:"instance"`
	Up       bool        `json:"up"`
	Status   interface{} `json:"systemStatus,omitempty"`
}

// getVersion returns application run and build time data.
func (c *Client) getVersion() map[string]interface{} {
	numPlex := 0 // maybe one day we'll support more than 1 plex.
	if c.Config.Plex.Configured() {
		numPlex = 1
	}

	return map[string]interface{}{
		"version":        version.Version,
		"os_arch":        runtime.GOOS + "." + runtime.GOARCH,
		"uptime":         time.Since(version.Started).Round(time.Second).String(),
		"uptime_seconds": time.Since(version.Started).Round(time.Second).Seconds(),
		"build_date":     version.BuildDate,
		"branch":         version.Branch,
		"go_version":     version.GoVersion,
		"revision":       version.Revision,
		"gui":            ui.HasGUI(),
		"num_lidarr":     len(c.Config.Apps.Lidarr),
		"num_sonarr":     len(c.Config.Apps.Sonarr),
		"num_radarr":     len(c.Config.Apps.Radarr),
		"num_readarr":    len(c.Config.Apps.Readarr),
		"num_plex":       numPlex,
	}
}

// versionResponse returns application run and build time data and application statuses: /api/version.
func (c *Client) versionResponse(r *http.Request) (int, interface{}) {
	var (
		output = c.getVersion()
		rad    = make([]*conTest, len(c.Config.Radarr))
		read   = make([]*conTest, len(c.Config.Readarr))
		son    = make([]*conTest, len(c.Config.Sonarr))
		lid    = make([]*conTest, len(c.Config.Lidarr))
		status = &appStatus{Radarr: rad, Readarr: read, Sonarr: son, Lidarr: lid}
	)

	for i, app := range c.Config.Radarr {
		stat, err := app.GetSystemStatus()
		rad[i] = &conTest{Instance: i + 1, Up: err == nil, Status: stat}
	}

	for i, app := range c.Config.Readarr {
		stat, err := app.GetSystemStatus()
		read[i] = &conTest{Instance: i + 1, Up: err == nil, Status: stat}
	}

	for i, app := range c.Config.Sonarr {
		stat, err := app.GetSystemStatus()
		son[i] = &conTest{Instance: i + 1, Up: err == nil, Status: stat}
	}

	for i, app := range c.Config.Lidarr {
		stat, err := app.GetSystemStatus()
		lid[i] = &conTest{Instance: i + 1, Up: err == nil, Status: stat}
	}

	if c.Config.Plex.Configured() {
		stat, err := c.Config.Plex.GetInfo()
		if stat == nil {
			stat = &plex.PMSInfo{}
		}

		status.Plex = []*conTest{{
			Instance: 1,
			Up:       err == nil,
			Status: map[string]interface{}{
				"friendlyName":             stat.FriendlyName,
				"version":                  stat.Version,
				"updatedAt":                stat.UpdatedAt,
				"platform":                 stat.Platform,
				"platformVersion":          stat.PlatformVersion,
				"size":                     stat.Size,
				"myPlexSigninState":        stat.MyPlexSigninState,
				"myPlexSubscription":       stat.MyPlexSubscription,
				"pushNotifications":        stat.PushNotifications,
				"streamingBrainVersion":    stat.StreamingBrainVersion,
				"streamingBrainABRVersion": stat.StreamingBrainABRVersion,
			},
		}}
	}

	output["app_status"] = status

	return http.StatusOK, output
}

// notFound is the handler for paths that are not found: 404s.
func (c *Client) notFound(w http.ResponseWriter, r *http.Request) {
	c.Config.Respond(w, http.StatusNotFound, "Check your request parameters and try again.")
}

func (c *Client) statusResponse(r *http.Request) (int, interface{}) {
	return http.StatusOK, c.Flags.Name() + " alive!"
}

func (c *Client) slowResponse(r *http.Request) (int, interface{}) {
	time.Sleep(100 * time.Millisecond) //nolint:gomnd
	return http.StatusOK, ""
}

// slash is the handler for /.
func (c *Client) slash(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("<p>" + c.Flags.Name() + ": <strong>working</strong></p>\n"))
}

func (c *Client) favIcon(w http.ResponseWriter, r *http.Request) {
	b, err := bindata.Asset("files/favicon.ico")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, r.URL.Path, time.Now(), bytes.NewReader(b))
}

// stripSecrets runs first to save a redacted URI in a special request header.
// The logger uses this special value to save a redacted URI in the log file.
func (c *Client) stripSecrets(next http.Handler) http.Handler {
	s := []string{c.Config.Apps.APIKey}
	// gather configured/known secrets.
	if c.Config.Plex != nil {
		s = append(s, c.Config.Plex.Token)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		// then redact secrets from request.
		for _, s := range s {
			if s != "" {
				uri = strings.ReplaceAll(uri, s, "<redacted>")
			}
		}

		// save into a request header for the logger.
		r.Header.Set("X-Redacted-URI", uri)
		next.ServeHTTP(w, r)
	})
}

// fixForwardedFor sets the X-Forwarded-For header to the client IP
// under specific circumstances.
func (c *Client) fixForwardedFor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Trim(r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")], "[]")
		if x := r.Header.Get("X-Forwarded-For"); x == "" || !c.Config.Allow.Contains(ip) {
			r.Header.Set("X-Forwarded-For", ip)
		} else if l := strings.LastIndexAny(x, ", "); l != -1 {
			r.Header.Set("X-Forwarded-For", strings.Trim(x[l:len(x)-1], ", "))
		}

		next.ServeHTTP(w, r)
	})
}
