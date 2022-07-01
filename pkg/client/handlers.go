package client

import (
	"bytes"
	"expvar"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golift.io/starr"
)

// httpHandlers initializes GUI HTTP routes.
func (c *Client) httpHandlers() {
	c.httpAPIHandlers() // Init API handlers up front.
	// 404 (or redirect to base path) everything else
	defer func() {
		c.Config.Router.PathPrefix("/").HandlerFunc(c.notFound)
	}()

	base := path.Join("/", c.Config.URLBase)

	c.Config.Router.HandleFunc("/favicon.ico", c.favIcon).Methods("GET")
	c.Config.Router.HandleFunc(strings.TrimSuffix(base, "/")+"/", c.slash).Methods("GET")
	c.Config.Router.HandleFunc(strings.TrimSuffix(base, "/")+"/", c.loginHandler).Methods("POST")

	// Handle the same URLs as above on the different base URL too.
	if !strings.EqualFold(base, "/") {
		c.Config.Router.HandleFunc(path.Join(base, "favicon.ico"), c.favIcon).Methods("GET")
		c.Config.Router.HandleFunc(base, c.slash).Methods("GET")
		c.Config.Router.HandleFunc(base, c.loginHandler).Methods("POST")
	}

	if c.Config.UIPassword == "" {
		return
	}

	c.Config.Router.PathPrefix(path.Join(base, "/files/")).
		Handler(http.StripPrefix(strings.TrimSuffix(base, "/"), http.HandlerFunc(c.handleStaticAssets))).Methods("GET")
	c.Config.Router.HandleFunc(path.Join(base, "/logout"), c.logoutHandler).Methods("GET", "POST")
	c.httpGuiHandlers(base)
}

func (c *Client) httpGuiHandlers(base string) {
	// gui is used for authorized paths.
	gui := c.Config.Router.PathPrefix(base).Subrouter()
	gui.Use(c.checkAuthorized) // check password or x-webauth-user header.
	gui.Handle("/debug/vars", expvar.Handler()).Methods("GET")
	gui.HandleFunc("/deleteFile/{source}/{id}", c.getFileDeleteHandler).Methods("GET")
	gui.HandleFunc("/downloadFile/{source}/{id}", c.getFileDownloadHandler).Methods("GET")
	gui.HandleFunc("/getFile/{source}/{id}/{lines}/{skip}", c.getFileHandler).Methods("GET").Queries("sort", "{sort}")
	gui.HandleFunc("/getFile/{source}/{id}/{lines}/{skip}", c.getFileHandler).Methods("GET")
	gui.HandleFunc("/getFile/{source}/{id}/{lines}", c.getFileHandler).Methods("GET").Queries("sort", "{sort}")
	gui.HandleFunc("/getFile/{source}/{id}/{lines}", c.getFileHandler).Methods("GET")
	gui.HandleFunc("/getFile/{source}/{id}", c.getFileHandler).Methods("GET").Queries("sort", "{sort}")
	gui.HandleFunc("/getFile/{source}/{id}", c.getFileHandler).Methods("GET")
	gui.HandleFunc("/profile", c.handleProfilePost).Methods("POST")
	gui.HandleFunc("/ps", c.handleProcessList).Methods("GET")
	gui.HandleFunc("/regexTest", c.handleRegexTest).Methods("POST")
	gui.HandleFunc("/reconfig", c.handleConfigPost).Methods("POST")
	gui.HandleFunc("/reload", c.handleReload).Methods("GET")
	gui.HandleFunc("/services/check/{service}", c.handleServicesCheck).Methods("GET")
	gui.HandleFunc("/services/{action:stop|start}", c.handleServicesStopStart).Methods("GET")
	gui.HandleFunc("/shutdown", c.handleShutdown).Methods("GET")
	gui.HandleFunc("/template/{template}", c.getTemplatePageHandler).Methods("GET")
	gui.HandleFunc("/trigger/{action}/{content}", c.handleGUITrigger).Methods("GET")
	gui.HandleFunc("/trigger/{action}", c.handleGUITrigger).Methods("GET")
	gui.HandleFunc("/checkInstance/{type}/{index}", c.handleInstanceCheck).Methods("POST")
	gui.HandleFunc("/stopFileWatch/{index}", c.handleStopFileWatcher).Methods("GET")
	gui.HandleFunc("/startFileWatch/{index}", c.handleStartFileWatcher).Methods("GET")
	gui.HandleFunc("/ws", c.handleWebSockets).Queries("source", "{source}", "fileId", "{fileId}").Methods("GET")
	gui.PathPrefix("/").HandlerFunc(c.notFound)
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
		c.Config.HandleAPIpath(starr.Plex, "directory", c.Config.Plex.HandleDirectory, "GET")
		c.Config.HandleAPIpath(starr.Plex, "emptytrash/{key}", c.Config.Plex.HandleEmptyTrash, "GET")
		c.Config.HandleAPIpath(starr.Plex, "markwatched/{key}", c.Config.Plex.HandleMarkWatched, "GET")
		c.Config.HandleAPIpath(starr.Plex, "kill", c.Config.Plex.HandleKillSession, "GET").
			Queries("reason", "{reason:.*}", "sessionId", "{sessionId:[0-9a-z-]+}")

		tokens := fmt.Sprintf("{token:%s|%s}", c.Config.Plex.Token, c.Config.Apps.APIKey)
		c.Config.Router.HandleFunc("/plex", c.PlexHandler).Methods("POST").Queries("token", tokens)
		c.Config.Router.HandleFunc("/", c.PlexHandler).Methods("POST").Queries("token", tokens)

		if c.Config.URLBase != "/" {
			// Allow plex to use the base url too.
			c.Config.Router.HandleFunc(path.Join(c.Config.URLBase, "plex"), c.PlexHandler).
				Methods("POST").Queries("token", tokens)
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
		c.Logger.Errorf("Sending HTTP Reply: %v", err)
	}
}

// slash is the GET handler for /.
func (c *Client) slash(response http.ResponseWriter, request *http.Request) {
	c.indexPage(response, request, "")
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

func (c *Client) addUsernameHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		if username, _ := c.getUserName(req); username != "" {
			req.Header.Set("X-Username", username)
		}
		next.ServeHTTP(response, req)
	})
}

func (c *Client) countRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		exp.HTTPRequests.Add("Total Requests", 1)

		switch {
		case strings.HasPrefix(req.RequestURI, path.Join(c.Config.URLBase, "api")):
			exp.HTTPRequests.Add("/api Requests", 1)
		case strings.HasPrefix(req.RequestURI, path.Join(c.Config.URLBase, "ws")):
			exp.HTTPRequests.Add("Websocket Requests", 1)
		default:
			exp.HTTPRequests.Add("Non-/api Requests", 1)
		}

		wrap := &responseWrapper{ResponseWriter: response, statusCode: http.StatusOK}
		next.ServeHTTP(wrap, req)
		exp.HTTPRequests.Add(fmt.Sprintf("Response %d %s", wrap.statusCode, http.StatusText(wrap.statusCode)), 1)
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

func (c *Client) handleTrigger(r *http.Request) (int, interface{}) {
	return c.runTrigger(website.EventAPI, mux.Vars(r)["trigger"], mux.Vars(r)["content"])
}

func (c *Client) runTrigger(source website.EventType, trigger, content string) (int, string) { //nolint:cyclop,funlen
	if content != "" {
		c.Logger.Debugf("Incoming API Trigger: %s (%s)", trigger, content)
	} else {
		c.Logger.Debugf("Incoming API Trigger: %s", trigger)
	}

	title := cases.Title(language.AmericanEnglish)

	switch trigger {
	case "cfsync":
		c.triggers.CFSync.SyncRadarrCF(source)
		return http.StatusOK, "TRaSH Custom Formats Radarr Sync initiated."
	case "qpsync":
		c.triggers.CFSync.SyncSonarrRP(source)
		return http.StatusOK, "TRaSH Release Profile Sonarr Sync initiated."
	case "services":
		c.Config.Services.RunChecks(source)
		return http.StatusOK, "All service checks rescheduled for immediate exeution."
	case "sessions":
		if !c.Config.Plex.Configured() {
			return http.StatusNotImplemented, "Plex Sessions are not enabled."
		}

		c.triggers.PlexCron.Send(source)

		return http.StatusOK, "Plex sessions triggered."
	case "stuckitems":
		c.triggers.StuckItems.Send(source)
		return http.StatusOK, "Stuck Queue Items triggered."
	case "dashboard":
		c.triggers.Dashboard.Send(source)
		return http.StatusOK, "Dashboard states triggered."
	case "snapshot":
		c.triggers.SnapCron.Send(source)
		return http.StatusOK, "System Snapshot triggered."
	case "gaps":
		c.triggers.Gaps.Send(source)
		return http.StatusOK, "Radarr Collections Gaps initiated."
	case "corrupt":
		err := c.triggers.Backups.Corruption(source, starr.App(title.String(content)))
		if err != nil {
			return http.StatusBadRequest, "Corruption trigger failed: " + err.Error()
		}

		return http.StatusOK, title.String(content) + " corruption checks initiated."
	case "backup":
		err := c.triggers.Backups.Backup(source, starr.App(title.String(content)))
		if err != nil {
			return http.StatusBadRequest, "Backup trigger failed: " + err.Error()
		}

		return http.StatusOK, title.String(content) + " backups check initiated."
	case "reload":
		defer c.triggerConfigReload(website.EventAPI, "HTTP Triggered Reload")
		return http.StatusOK, "Application reload initiated."
	case "notification":
		if content != "" {
			ui.Notify("Notification: %s", content) //nolint:errcheck
			c.Logger.Printf("NOTIFICATION: %s", content)

			return http.StatusOK, "Local Nntification sent."
		}

		return http.StatusBadRequest, "Missing notification content."
	default:
		return http.StatusBadRequest, "Unknown trigger provided:'" + trigger + "'"
	}
}
