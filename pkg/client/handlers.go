package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/gorilla/mux"
	"golift.io/starr"
)

// internalHandlers initializes "special" internal API paths.
func (c *Client) internalHandlers() {
	c.Config.HandleAPIpath("", "slow", c.slowResponse, "HEAD") // log testing
	c.Config.HandleAPIpath("", "status", c.statusResponse, "GET", "HEAD")
	c.Config.HandleAPIpath("", "version", c.notifiarr.VersionHandler, "GET", "HEAD")
	c.Config.HandleAPIpath("", "info", c.updateInfo, "PUT")
	c.Config.HandleAPIpath("", "info/alert", c.updateInfoAlert, "PUT")
	c.Config.HandleAPIpath("", "trigger/{trigger:[0-9a-z-]+}", c.handleTrigger, "GET")

	if c.Config.Plex.Configured() {
		c.Config.HandleAPIpath(starr.Plex, "sessions", c.Config.Plex.HandleSessions, "GET")
		c.Config.HandleAPIpath(starr.Plex, "kill", c.Config.Plex.HandleKillSession, "GET").
			Queries("reason", "{reason:.*}", "sessionId", "{sessionId:[0-9a-z-]+}")

		tokens := fmt.Sprintf("{token:%s|%s}", c.Config.Plex.Token, c.Config.Apps.APIKey)
		c.Config.Router.Handle("/plex",
			http.HandlerFunc(c.notifiarr.PlexHandler)).Methods("POST").Queries("token", tokens)

		if c.Config.URLBase != "/" {
			// Allow plex to use the base url too.
			c.Config.Router.Handle(path.Join(c.Config.URLBase, "plex"),
				http.HandlerFunc(c.notifiarr.PlexHandler)).Methods("POST").Queries("token", tokens)
		}
	}

	// Initialize internal-only paths.
	c.Config.Router.Handle("/favicon.ico", http.HandlerFunc(c.favIcon)) // built-in icon.
	c.Config.Router.Handle("/", http.HandlerFunc(c.slash))              // "hi" page on /

	if base := c.Config.URLBase; !strings.EqualFold(base, "/") {
		// Handle the same URLs on the different base URL too.
		c.Config.Router.Handle(path.Join(base, "favicon.ico"), http.HandlerFunc(c.favIcon))
		c.Config.Router.Handle(base, http.HandlerFunc(c.slash))     // "hi" page on /urlbase
		c.Config.Router.Handle(base+"/", http.HandlerFunc(c.slash)) // "hi" page on /urlbase/
	}

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

func (c *Client) handleTrigger(r *http.Request) (int, interface{}) {
	trigger := mux.Vars(r)["trigger"]
	c.Debugf("Incoming API Trigger: %s", trigger)

	const apiTrigger = "apitrigger"

	switch {
	case trigger == "cfsync":
		c.notifiarr.Trigger.SyncCF(false)
	case trigger == "services" && c.Config.Services.Disabled:
		return http.StatusNotImplemented, "services not enabled"
	case trigger == "services":
		c.Config.Services.RunAllChecksSendResult(apiTrigger)
	case trigger == "sessions" && !c.Config.Plex.Configured():
		return http.StatusNotImplemented, "sessions not enabled"
	case trigger == "sessions":
		c.notifiarr.Trigger.SendPlexSessions(apiTrigger)
	case trigger == "stuckitems":
		c.notifiarr.Trigger.SendFinishedQueueItems(c.notifiarr.BaseURL)
	case trigger == "dashboard":
		c.notifiarr.Trigger.GetState()
	case trigger == "snapshot":
		c.notifiarr.Trigger.SendSnapshot(apiTrigger)
	default:
		return http.StatusBadRequest, "unknown trigger '" + trigger + "'"
	}

	return http.StatusOK, trigger + " initiated"
}
