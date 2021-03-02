package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/bindata"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/plex"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/ui"
	"golift.io/version"
)

// allowedIPs determines who can set x-forwarded-for.
type allowedIPs []*net.IPNet

// internalHandlers initializes "special" internal API paths.
func (c *Client) internalHandlers() {
	c.Config.HandleAPIpath("", "slow", c.slowResponse, "HEAD") // log testing
	c.Config.HandleAPIpath("", "status", c.statusResponse, "GET", "HEAD")
	c.Config.HandleAPIpath("", "version", c.versionResponse, "GET", "HEAD")
	c.Config.HandleAPIpath("", "info", c.updateInfo, "PUT")
	c.Config.HandleAPIpath("", "info/alert", c.updateInfoAlert, "PUT")

	if c.Config.Plex != nil && c.Config.Plex.Token != "" {
		c.Config.Router.Handle("/plex",
			http.HandlerFunc(c.plexIncoming)).Methods("POST").Queries("token", c.Config.Plex.Token)
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

	c.Print("New Info from DiscordNotifier.com:", string(body))

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
		_, _ = ui.Warning(Title+" Alert", body)
		c.alert.Done() //nolint:wsl
	}()

	return code, err
}

// versionResponse returns application run and build time data: /api/version.
func (c *Client) versionResponse(r *http.Request) (int, interface{}) {
	return http.StatusOK, map[string]interface{}{
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
	}
}

// notFound is the handler for paths that are not found: 404s.
func (c *Client) notFound(w http.ResponseWriter, r *http.Request) {
	c.Config.Respond(w, http.StatusNotFound, "Check your request parameters and try again.", 0)
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
	if b, err := bindata.Asset("files/favicon.ico"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		http.ServeContent(w, r, r.URL.Path, time.Now(), bytes.NewReader(b))
	}
}

func (c *Client) plexIncoming(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(1000 * 100); err != nil { // nolint:gomnd // 100kbyte memory usage
		c.Errorf("Parsing Multipart Form (plex): %v", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	var v plex.Webhook

	switch err := json.Unmarshal([]byte(r.Form.Get("payload")), &v); {
	case err != nil:
		c.Errorf("Unmarshalling Plex payload: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	case !strings.HasPrefix(v.Event, "media"):
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ignored, non-media\n"))
	case c.plex.Active():
		c.Printf("Plex Incoming Webhook IGNORED (cooldown): %s, %s '%s' => %s",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ignored, cooldown\n"))
	default:
		go c.collectSessions(&v)
	}
}

func (c *Client) collectSessions(v *plex.Webhook) {
	defer c.plex.Done()
	c.Printf("Plex Incoming Webhook: %s, %s '%s' => %s (collecting sessions)",
		v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)

	body, err := c.notifiarr.SendMeta(v, plex.WaitTime)
	if err != nil {
		c.Errorf("Sending Plex Session to Notifiarr: %v", err)
		return
	}

	if fields := strings.Split(string(body), `"`); len(fields) > 3 { // nolint:gomnd
		c.Printf("Plex => Notifiarr: %s '%s' => %s (%s)", v.Account.Title, v.Event, v.Metadata.Title, fields[3])
	} else {
		c.Printf("Plex => Notifiarr: %s '%s' => %s", v.Account.Title, v.Event, v.Metadata.Title)
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

// String turns a list of allowedIPs into a printable masterpiece.
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
