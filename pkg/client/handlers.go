package client

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/CAFxX/httpcompression"
	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	"golift.io/starr"
)

// httpHandlers initializes GUI HTTP routes.
func (c *Client) httpHandlers() {
	c.httpAPIHandlers() // Init API handlers up front.

	compress, _ := httpcompression.DefaultAdapter()
	gzip := func(handler http.HandlerFunc) http.Handler {
		return compress(handler)
	}

	// 404 (or redirect to base path) everything else
	defer func() {
		c.Config.Router.PathPrefix("/").Handler(gzip(c.notFound))
	}()

	base := path.Join("/", c.Config.URLBase)

	c.Config.Router.Handle("/favicon.ico", gzip(c.favIcon)).Methods("GET")
	c.Config.Router.Handle(strings.TrimSuffix(base, "/")+"/", gzip(c.slash)).Methods("GET")
	c.Config.Router.Handle(strings.TrimSuffix(base, "/")+"/", gzip(c.loginHandler)).Methods("POST")

	// Handle the same URLs as above on the different base URL too.
	if !strings.EqualFold(base, "/") {
		c.Config.Router.Handle(path.Join(base, "favicon.ico"), gzip(c.favIcon)).Methods("GET")
		c.Config.Router.Handle(base, gzip(c.slash)).Methods("GET")
		c.Config.Router.Handle(base, gzip(c.loginHandler)).Methods("POST")
	}

	if c.Config.UIPassword == "" {
		return
	}

	c.Config.Router.PathPrefix(path.Join(base, "/files/")).
		Handler(http.StripPrefix(strings.TrimSuffix(base, "/"), http.HandlerFunc(c.handleStaticAssets))).Methods("GET")
	c.Config.Router.Handle(path.Join(base, "/logout"), gzip(c.logoutHandler)).Methods("GET", "POST")
	c.httpGuiHandlers(base, compress)
}

func (c *Client) httpGuiHandlers(base string, compress func(handler http.Handler) http.Handler) {
	// gui is used for authorized paths. All these paths have a prefix of /ui.
	gui := c.Config.Router.PathPrefix(path.Join(base, "/ui")).Subrouter()
	gui.Use(c.checkAuthorized) // check password or x-webauth-user header.
	gui.Use(compress)
	gui.Handle("/debug/vars", expvar.Handler()).Methods("GET")
	gui.HandleFunc("/deleteFile/{source}/{id}", c.getFileDeleteHandler).Methods("GET")
	gui.HandleFunc("/downloadFile/{source}/{id}", c.getFileDownloadHandler).Methods("GET")
	gui.HandleFunc("/uploadFile/{source}/{id}", c.uploadFileHandler).Methods("GET")
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
	gui.HandleFunc("/ping", c.handlePing).Methods("GET")
	gui.HandleFunc("/services/check/{service}", c.handleServicesCheck).Methods("GET")
	gui.HandleFunc("/services/{action:stop|start}", c.handleServicesStopStart).Methods("GET")
	gui.HandleFunc("/shutdown", c.handleShutdown).Methods("GET")
	gui.HandleFunc("/template/{template}", c.getTemplatePageHandler).Methods("GET")
	gui.HandleFunc("/trigger/{trigger}/{content}", c.triggers.Handler).Methods("GET")
	gui.HandleFunc("/trigger/{trigger}", c.triggers.Handler).Methods("GET")
	gui.HandleFunc("/tunnel/ping", c.pingTunnels).Methods("GET")
	gui.HandleFunc("/tunnel/save", c.saveTunnels).Methods("POST")
	gui.HandleFunc("/checkInstance/{type}/{index}", c.handleInstanceCheck).Methods("POST")
	gui.HandleFunc("/stopFileWatch/{index}", c.handleStopFileWatcher).Methods("GET")
	gui.HandleFunc("/startFileWatch/{index}", c.handleStartFileWatcher).Methods("GET")
	gui.HandleFunc("/browse", c.handleFileBrowser).Queries("dir", "{dir}").Methods("GET")
	gui.HandleFunc("/ajax/{path:cmdstats|cmdargs}/{hash}", c.handleCommandStats).Methods("GET")
	gui.HandleFunc("/runCommand/{hash}", c.handleRunCommand).Methods("POST")
	gui.HandleFunc("/ws", c.handleWebSockets).Queries("source", "{source}", "fileId", "{fileId}").Methods("GET")
	gui.HandleFunc("/docs/json/{instance}", c.handlerSwaggerDoc).Methods("GET")
	gui.HandleFunc("/ui.json", c.handlerSwaggerDoc).Methods("GET")
	gui.Handle("/docs", http.RedirectHandler(path.Join(base, "ui", "docs")+"/", http.StatusFound))
	gui.HandleFunc("/docs/", c.handleSwaggerIndex).Methods("GET")
	gui.PathPrefix("/").HandlerFunc(c.notFound)
}

// httpAPIHandlers initializes API routes.
func (c *Client) httpAPIHandlers() {
	c.Config.HandleAPIpath("", "info", c.triggers.CI.InfoHandler, "GET", "HEAD")
	c.Config.HandleAPIpath("", "version", c.triggers.CI.VersionHandler, "GET", "HEAD")
	c.Config.HandleAPIpath("", "version/{app}/{instance:[0-9]+}", c.triggers.CI.VersionHandlerInstance, "GET", "HEAD")
	c.Config.HandleAPIpath("", "trigger/{trigger:[0-9a-z-]+}", c.triggers.APIHandler, "GET", "POST")
	c.Config.HandleAPIpath("", "trigger/{trigger:[0-9a-z-]+}/{content}", c.triggers.APIHandler, "GET", "POST")
	c.Config.HandleAPIpath("", "services/{action}", c.Config.Services.APIHandler, "GET")
	c.Config.HandleAPIpath("", "triggers", c.triggers.HandleGetTriggers, "GET")
	c.Config.HandleAPIpath("", "ping", c.handleInstancePing, "GET")
	c.Config.HandleAPIpath("", "ping/{app:[a-z,]+}", c.handleInstancePing, "GET")
	c.Config.HandleAPIpath("", "ping/{app:[a-z]+}/{instance:[0-9]+}", c.handleInstancePing, "GET")

	// Aggregate handlers. Non-app specific.
	c.Config.HandleAPIpath("", "/trash/{app}", c.triggers.CFSync.Handler, "POST")

	if c.Config.Plex.Enabled() {
		c.Config.HandleAPIpath(starr.Plex, "sessions", c.Config.Plex.HandleSessions, "GET")
		c.Config.HandleAPIpath(starr.Plex, "directory", c.Config.Plex.HandleDirectory, "GET")
		c.Config.HandleAPIpath(starr.Plex, "emptytrash/{key}", c.Config.Plex.HandleEmptyTrash, "GET")
		c.Config.HandleAPIpath(starr.Plex, "markwatched/{key}", c.Config.Plex.HandleMarkWatched, "GET")
		c.Config.HandleAPIpath(starr.Plex, "kill", c.Config.Plex.HandleKillSession, "GET").
			Queries("reason", "{reason:.*}", "sessionId", "{sessionId:.*}")

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

	if err := c.template.ExecuteTemplate(response, "404.html", nil); err != nil {
		c.Logger.Errorf("Sending HTTP Reply: %v", err)
	}
}

// slash is the GET handler for /.
func (c *Client) slash(response http.ResponseWriter, request *http.Request) {
	if !strings.HasSuffix(request.URL.Path, "/") {
		http.Redirect(response, request, request.URL.Path+"/", http.StatusPermanentRedirect)
		return
	}

	c.indexPage(request.Context(), response, request, "")
}

func (c *Client) favIcon(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
	ico, err := bindata.Files.ReadFile("files/images/favicon.ico")
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
	if c.Config.Plex.Enabled() {
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
		r.Header.Set("X-Redacted-Uri", uri)
		next.ServeHTTP(w, r)
	})
}

func (c *Client) addUsernameHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		if username, _ := c.getUserName(req); username != "" {
			req.Header.Set("X-Noticlient-Username", username)
		}

		next.ServeHTTP(response, req)
	})
}

func (c *Client) countRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		mnd.HTTPRequests.Add("Total Requests", 1)

		switch {
		case strings.HasPrefix(req.RequestURI, path.Join(c.Config.URLBase, "api")):
			mnd.HTTPRequests.Add("/api Requests", 1)
		case strings.HasPrefix(req.RequestURI, path.Join(c.Config.URLBase, "ws")):
			mnd.HTTPRequests.Add("Websocket Requests", 1)
		default:
			mnd.HTTPRequests.Add("Non-/api Requests", 1)
		}

		wrap := &responseWrapper{ResponseWriter: response, statusCode: http.StatusOK}
		next.ServeHTTP(wrap, req)
		mnd.HTTPRequests.Add(fmt.Sprintf("Response %d %s", wrap.statusCode, http.StatusText(wrap.statusCode)), 1)
	})
}

// fixForwardedFor sets the X-Forwarded-For header to the client IP
// under specific circumstances.
func (c *Client) fixForwardedFor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
		if xff := r.Header.Get("X-Forwarded-For"); xff == "" || !c.Config.Allow.Contains(r.RemoteAddr) {
			if end := strings.LastIndex(r.RemoteAddr, ":"); end != -1 {
				r.Header.Set("X-Forwarded-For", strings.Trim(r.RemoteAddr[:end], "[]"))
			} else if ra := strings.Trim(r.RemoteAddr, "[]"); ra != "" {
				r.Header.Set("X-Forwarded-For", ra)
			} else {
				r.Header.Set("X-Forwarded-For", "unknown")
			}
		} else if l := strings.LastIndexAny(xff, ", "); l != -1 {
			r.Header.Set("X-Forwarded-For", strings.Trim(xff[l:len(xff)-1], ", "))
		}

		next.ServeHTTP(w, r)
	})
}

// @Description  Returns true or false for 1 requested instance. True is up, false is down.
// @Summary      Ping 1 starr instance.
// @Tags         Client
// @Produce      json
// @Param        app      path string  true  "Application" Enums(lidarr, prowlarr, radarr, readarr, sonarr)
// @Param        instance path int64   true  "Application instance (1-index)."
// @Success      200  {object} apps.Respond.apiResponse{message=map[string]map[int]bool} "map for app->instance->up"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/ping/{app}/{instance} [get]
// @Security     ApiKeyAuth
func _() {}

// @Description  Returns true or false for every instance for starr app requested. True is up, false is down.
// @Description  Multiple apps may be provided by separating them with a comma. ie /api/ping/radarr,sonarr
// @Summary      Ping all instances for 1 or more starr apps.
// @Tags         Client
// @Produce      json
// @Param        apps  path   string  true  "Application, comma separated" Enums(lidarr, prowlarr, radarr, readarr, sonarr)
// @Success      200  {object} apps.Respond.apiResponse{message=map[string]map[int]bool} "map for app->instance->up"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/ping/{apps} [get]
// @Security     ApiKeyAuth
//
//nolint:lll
func _() {}

// @Description  Returns true or false for each configured starr instance. True is up, false is down.
// @Summary      Ping all starr instances.
// @Tags         Client
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=map[string]map[int]bool} "map for app->instance->up"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/ping [get]
// @Security     ApiKeyAuth
func (c *Client) handleInstancePing(req *http.Request) (int, interface{}) { //nolint:cyclop
	apps := strings.Split(mux.Vars(req)["app"], ",")
	instance, _ := strconv.Atoi(mux.Vars(req)["instance"])
	output := make(map[string]map[int]bool)

	if len(apps) == 0 || len(apps) == 1 && apps[0] == "" {
		instance, apps = 0, []string{
			starr.Lidarr.Lower(),
			starr.Radarr.Lower(),
			starr.Readarr.Lower(),
			starr.Sonarr.Lower(),
			starr.Prowlarr.Lower(),
		}
	}

	for _, app := range apps {
		switch app {
		case starr.Lidarr.Lower():
			for idx := range c.Config.Apps.Lidarr {
				c.pingInstance(req.Context(), c.Config.Apps.Lidarr[idx], app, idx, instance, output)
			}
		case starr.Radarr.Lower():
			for idx := range c.Config.Apps.Radarr {
				c.pingInstance(req.Context(), c.Config.Apps.Radarr[idx], app, idx, instance, output)
			}
		case starr.Readarr.Lower():
			for idx := range c.Config.Apps.Readarr {
				c.pingInstance(req.Context(), c.Config.Apps.Readarr[idx], app, idx, instance, output)
			}
		case starr.Sonarr.Lower():
			for idx := range c.Config.Apps.Sonarr {
				c.pingInstance(req.Context(), c.Config.Apps.Sonarr[idx], app, idx, instance, output)
			}
		case starr.Prowlarr.Lower():
			for idx := range c.Config.Apps.Prowlarr {
				c.pingInstance(req.Context(), c.Config.Apps.Prowlarr[idx], app, idx, instance, output)
			}
		}
	}

	return http.StatusOK, output
}

type instancePinger interface {
	PingContext(ctx context.Context) error
	Enabled() bool
}

func (c *Client) pingInstance(
	ctx context.Context,
	pinger instancePinger,
	app string,
	idx, instance int,
	output map[string]map[int]bool,
) {
	if pinger.Enabled() && (instance == 0 || instance == idx+1) {
		if output[app] == nil {
			output[app] = make(map[int]bool)
		}

		output[app][idx+1] = pinger.PingContext(ctx) == nil
	}
}
