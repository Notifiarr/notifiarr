package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/bindata"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/ui"
	"golift.io/version"
)

// allowedIPs determines who can set x-forwarded-for.
type allowedIPs []*net.IPNet

// internalHandlers initializes "special" internal API paths.
func (c *Client) internalHandlers() {
	// GET  /api/status   (w/o key)
	c.Config.Router.Handle(path.Join("/", c.Config.URLBase, "api", "status"),
		http.HandlerFunc(c.statusResponse)).Methods("GET", "HEAD")
	// PUT  /api/info     (w/ key)
	c.Config.HandleAPIpath("", "info", c.updateInfo, "PUT")
	// PUT  /api/info/alert     (w/ key)
	c.Config.HandleAPIpath("", "info/alert", c.updateInfoAlert, "PUT")
	// GET  /api/version  (w/ key)
	c.Config.HandleAPIpath("", "version", c.versionResponse, "GET", "HEAD")

	// Initialize internal-only paths.
	c.Config.Router.Handle("/favicon.ico", http.HandlerFunc(c.favIcon))   // built-in icon.
	c.Config.Router.Handle("/", http.HandlerFunc(c.slash))                // "hi" page on /
	c.Config.Router.PathPrefix("/").Handler(http.HandlerFunc(c.notFound)) // 404 everything
}

func (c *Client) respond(w http.ResponseWriter, stat int, msg interface{}, start time.Time) {
	w.Header().Set("X-Request-Time", time.Since(start).Round(time.Microsecond).String())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(stat)

	statusTxt := strconv.Itoa(stat) + ": " + http.StatusText(stat)

	if m, ok := msg.(error); ok {
		c.Errorf("Status: %s, Message: %v", statusTxt, m)
		msg = m.Error()
	}

	b, _ := json.Marshal(map[string]interface{}{"status": statusTxt, "message": msg})
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n")) // curl likes new lines.
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

	c.alert.Lock()
	defer c.alert.Unlock()

	if c.alert.active {
		return http.StatusLocked, "previous alert not acknowledged"
	}

	c.alert.active = true

	go func() {
		_, _ = ui.Warning(Title+" Alert", body)

		c.alert.Lock()
		defer c.alert.Unlock()

		c.alert.active = false
	}()

	return code, err
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
	c.respond(w, http.StatusNotFound, "Check your request parameters and try again.", time.Now())
}

// statusResponse is the handler for /api/status.
func (c *Client) statusResponse(w http.ResponseWriter, r *http.Request) {
	c.respond(w, http.StatusOK, c.Flags.Name()+" alive!", time.Now())
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
