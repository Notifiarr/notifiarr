package client

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/gorilla/mux"
	"golift.io/starr"
)

// httpHandlers initializes GUI HTTP routes.
func (c *Client) httpHandlers() {
	c.httpAPIHandlers() // Init API handlers up front.

	base := path.Join("/", c.Config.URLBase)

	c.Config.Router.Handle("/favicon.ico", http.HandlerFunc(c.favIcon))
	c.Config.Router.Handle(strings.TrimSuffix(base, "/")+"/", http.HandlerFunc(c.slash)).Methods("GET")
	c.Config.Router.Handle(strings.TrimSuffix(base, "/")+"/", http.HandlerFunc(c.loginHandler)).Methods("POST")

	if !strings.EqualFold(base, "/") {
		// Handle the same URLs as above on the different base URL too.
		c.Config.Router.Handle(path.Join(base, "favicon.ico"), http.HandlerFunc(c.favIcon))
		c.Config.Router.Handle(base, http.HandlerFunc(c.slash)).Methods("GET") // "hi" page on /urlbase
		c.Config.Router.Handle(base, http.HandlerFunc(c.loginHandler)).Methods("POST")
	}

	if c.Config.UIPassword != "" {
		c.Config.Router.PathPrefix(path.Join(base, "/files/")).
			Handler(http.StripPrefix(strings.TrimSuffix(base, "/"), http.HandlerFunc(c.handleStaticAssets))).Methods("GET")
		c.Config.Router.HandleFunc(path.Join(base, "/login"), c.loginHandler).Methods("GET", "POST")
		c.Config.Router.HandleFunc(path.Join(base, "/logout"), c.logoutHandler).Methods("POST", "GET")
		c.Config.Router.Handle(path.Join(base, "/get"), c.checkAuthorized(c.getSettingsHandler)).Methods("GET")
		c.Config.Router.Handle(path.Join(base, "/getLog/{id}"), c.checkAuthorized(c.getLogHandler)).Methods("GET")
		c.Config.Router.Handle(path.Join(base, "/getLog/{id}/{lines}"), c.checkAuthorized(c.getLogHandler)).Methods("GET")
		c.Config.Router.Handle(path.Join(base, "/getLog/{id}/{lines}/{skip}"),
			c.checkAuthorized(c.getLogHandler)).Methods("GET")
		c.Config.Router.Handle(path.Join(base, "/get/{config}"), c.checkAuthorized(c.getSettingsHandler)).Methods("GET")
	}

	// 404 (or redirect to base path) everything else
	c.Config.Router.PathPrefix("/").Handler(http.HandlerFunc(c.notFound))
}

// httpAPIHandlers initializes API routes.
func (c *Client) httpAPIHandlers() {
	c.Config.HandleAPIpath("", "version", c.versionHandler, "GET", "HEAD")
	c.Config.HandleAPIpath("", "trigger/{trigger:[0-9a-z-]+}", c.handleTrigger, "GET")
	c.Config.HandleAPIpath("", "trigger/{trigger:[0-9a-z-]+}/{content}", c.handleTrigger, "GET")
	// Aggregate handlers. Non-app specific.
	c.Config.HandleAPIpath("", "/trash/{app}", c.aggregateTrash, "POST")

	if c.Config.Plex.Configured() {
		c.Config.HandleAPIpath(starr.Plex, "sessions", c.Config.Plex.HandleSessions, "GET")
		c.Config.HandleAPIpath(starr.Plex, "kill", c.Config.Plex.HandleKillSession, "GET").
			Queries("reason", "{reason:.*}", "sessionId", "{sessionId:[0-9a-z-]+}")

		tokens := fmt.Sprintf("{token:%s|%s}", c.Config.Plex.Token, c.Config.Apps.APIKey)
		c.Config.Router.Handle("/plex",
			http.HandlerFunc(c.website.PlexHandler)).Methods("POST").Queries("token", tokens)
		c.Config.Router.Handle("/",
			http.HandlerFunc(c.website.PlexHandler)).Methods("POST").Queries("token", tokens)

		if c.Config.URLBase != "/" {
			// Allow plex to use the base url too.
			c.Config.Router.Handle(path.Join(c.Config.URLBase, "plex"),
				http.HandlerFunc(c.website.PlexHandler)).Methods("POST").Queries("token", tokens)
		}
	}
}

// notFound is the handler for paths that are not found: 404s.
func (c *Client) notFound(response http.ResponseWriter, request *http.Request) {
	if !strings.HasPrefix(request.URL.Path, c.Config.URLBase) {
		// If the request did not have the base url, redirect.
		http.Redirect(response, request, path.Join(c.Config.URLBase, request.URL.Path), http.StatusPermanentRedirect)
		return
	}

	response.WriteHeader(http.StatusNotFound)

	if err := c.templat.ExecuteTemplate(response, "404.html", nil); err != nil {
		c.Errorf("Sending HTTP Reply: %v", err)
	}
}

// slash is the GET handler for /.
func (c *Client) slash(response http.ResponseWriter, request *http.Request) {
	c.renderHTTPtemplate(response, request, "index.html", "")
}

func (c *Client) favIcon(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
	ico, err := bindata.Asset("files/images/favicon.ico")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, r.URL.Path, time.Now(), bytes.NewReader(ico))
}

// stripSecrets runs first to save a redacted URI in a special request header.
// The logger uses this special value to save a redacted URI in the log file.
func (c *Client) stripSecrets(next http.Handler) http.Handler {
	secrets := []string{c.Config.Apps.APIKey}
	secrets = append(secrets, c.Config.ExKeys...)
	// gather configured/known secrets.
	if c.Config.Plex != nil {
		secrets = append(secrets, c.Config.Plex.Token)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		uri := r.RequestURI
		// then redact secrets from request.
		for _, s := range secrets {
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		ip := strings.Trim(r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")], "[]")
		if x := r.Header.Get("X-Forwarded-For"); x == "" || !c.Config.Allow.Contains(ip) {
			r.Header.Set("X-Forwarded-For", ip)
		} else if l := strings.LastIndexAny(x, ", "); l != -1 {
			r.Header.Set("X-Forwarded-For", strings.Trim(x[l:len(x)-1], ", "))
		}

		next.ServeHTTP(w, r)
	})
}

func (c *Client) handleTrigger(r *http.Request) (int, interface{}) { //nolint:cyclop
	content := mux.Vars(r)["content"]
	trigger := mux.Vars(r)["trigger"]

	if content != "" {
		c.Debugf("Incoming API Trigger: %s (%s)", trigger, content)
	} else {
		c.Debugf("Incoming API Trigger: %s", trigger)
	}

	switch trigger {
	case "cfsync":
		c.website.Trigger.SyncCF(notifiarr.EventAPI)
	case "services":
		c.Config.Services.RunChecks(notifiarr.EventAPI)
	case "sessions":
		if !c.Config.Plex.Configured() {
			return http.StatusNotImplemented, "sessions not enabled"
		}

		c.website.Trigger.SendPlexSessions(notifiarr.EventAPI)
	case "stuckitems":
		c.website.Trigger.SendStuckQueueItems(notifiarr.EventAPI)
	case "dashboard":
		c.website.Trigger.SendDashboardState(notifiarr.EventAPI)
	case "snapshot":
		c.website.Trigger.SendSnapshot(notifiarr.EventAPI)
	case "gaps":
		c.website.Trigger.SendGaps(notifiarr.EventAPI)
	case "corrupt":
		err := c.website.Trigger.Corruption(notifiarr.EventAPI, starr.App(strings.Title(content)))
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("trigger failed: %w", err)
		}
	case "backup":
		err := c.website.Trigger.Backup(notifiarr.EventAPI, starr.App(strings.Title(content)))
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("trigger failed: %w", err)
		}
	case "reload":
		c.sighup <- &update.Signal{Text: "reload http triggered"}
	case "notification":
		if content != "" {
			ui.Notify("Notification: %s", content) //nolint:errcheck
			c.Printf("NOTIFICATION: %s", content)
		} else {
			return http.StatusBadRequest, "missing notification content"
		}
	default:
		return http.StatusBadRequest, "unknown trigger '" + trigger + "'"
	}

	if content != "" {
		return http.StatusOK, trigger + " (" + content + ") initiated"
	}

	return http.StatusOK, trigger + " initiated"
}
