package client

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/CAFxX/httpcompression"
	"github.com/Notifiarr/notifiarr/frontend"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
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

	defer func() {
		// SPA gets all the requests so it can handle its own page router.
		c.apps.Router.PathPrefix("/").Handler(gzip(c.loginHandler)).Methods("POST").Queries("login", "{login}")
		c.apps.Router.PathPrefix("/").Handler(gzip(frontend.IndexHandler)).Methods("GET")
		c.apps.Router.PathPrefix("/").Handler(gzip(c.notFound))
		// 404 (or redirect to base path) everything else
		c.apps.Router.PathPrefix("/").Handler(gzip(c.notFound))
	}()

	base := path.Join("/", c.Config.URLBase)
	frontend.URLBase = base

	c.apps.Router.Handle(strings.TrimSuffix(base, "/")+"/", gzip(c.slash)).Methods("GET")
	c.apps.Router.Handle(strings.TrimSuffix(base, "/")+"/", gzip(c.loginHandler)).Methods("POST")

	// Handle the same URLs as above on the different base URL too.
	if !strings.EqualFold(base, "/") {
		c.apps.Router.Handle(base, gzip(c.slash)).Methods("GET")
		c.apps.Router.Handle(base, gzip(c.loginHandler)).Methods("POST")
	}

	// If api key is set to "disabled", then the Web UI gets turned off.
	if c.Config.UIPassword == "disabled" {
		return
	}

	c.apps.Router.PathPrefix(path.Join(base, "/assets/")).
		Handler(http.StripPrefix(strings.TrimSuffix(base, "/"), gzip(frontend.IndexHandler)))
	c.apps.Router.Handle(path.Join(base, "/logout"), gzip(c.logoutHandler)).Methods("GET", "POST")

	// If there is no API key set, allow the user to set it from the GUI, and that's all they can do.
	if len(c.Config.APIKey) != website.APIKeyLength {
		gui := c.apps.Router.PathPrefix(path.Join(base, "/ui")).Subrouter()
		gui.Use(compress)
		gui.HandleFunc("/profile", c.handleProfileNoAPIKey).Methods("GET")
		gui.HandleFunc("/shutdown", c.handleShutdown).Methods("GET")
		gui.HandleFunc("/reload", c.handleReload).Methods("GET")
		c.apps.Router.NewRoute().Methods("PUT").Queries("setApiKey", "true").
			Headers("X-API-Key", "").HandlerFunc(c.handleAPIKey)
	} else {
		c.httpGuiHandlers(base, compress)
	}
}

func (c *Client) httpGuiHandlers(base string, compress func(handler http.Handler) http.Handler) {
	// gui is used for authorized paths. All these paths have a prefix of /ui.
	gui := c.apps.Router.PathPrefix(path.Join(base, "/ui")).Subrouter()
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
	gui.HandleFunc("/reconfig", c.handleConfigPost).Methods("POST").Queries("noreload", "{noreload}")
	gui.HandleFunc("/reconfig", c.handleConfigPost).Methods("POST")
	gui.HandleFunc("/reload", c.handleReload).Methods("GET")
	gui.HandleFunc("/ping", c.handlePing).Methods("GET")
	gui.HandleFunc("/services/check/{service}", c.handleServicesCheck).Methods("GET")
	gui.HandleFunc("/services/{action:stop|start}", c.handleServicesStopStart).Methods("GET")
	gui.HandleFunc("/shutdown", c.handleShutdown).Methods("GET")
	gui.HandleFunc("/profile", c.handleProfile).Methods("GET")
	gui.HandleFunc("/trigger/{trigger}/{content}", c.triggers.Handler).Methods("GET")
	gui.HandleFunc("/trigger/{trigger}", c.triggers.Handler).Methods("GET")
	gui.HandleFunc("/integrations", c.handleIntegrations).Methods("GET")
	gui.HandleFunc("/tunnel/ping", c.pingTunnels).Methods("GET")
	gui.HandleFunc("/tunnel/save", c.saveTunnels).Methods("POST")
	gui.HandleFunc("/checkAllInstances", c.handleCheckAll).Methods("GET")
	gui.HandleFunc("/checkInstance/{type}/{index}", c.handleInstanceCheck).Methods("POST")
	gui.HandleFunc("/stopFileWatch/{index}", c.handleStopFileWatcher).Methods("GET")
	gui.HandleFunc("/startFileWatch/{index}", c.handleStartFileWatcher).Methods("GET")
	gui.HandleFunc("/browse", c.handleNewFolder).Queries("dir", "{dir}", "new", "true").Methods("GET")
	gui.HandleFunc("/browse", c.handleNewFile).Queries("file", "{file}", "new", "true").Methods("GET")
	gui.HandleFunc("/browse", c.handleFileBrowser).Queries("dir", "{dir}").Methods("GET")
	gui.HandleFunc("/ajax/{path:cmdstats|cmdargs}/{hash}", c.handleCommandStats).Methods("GET")
	gui.HandleFunc("/runCommand/{hash}", c.handleRunCommand).Methods("POST")
	gui.HandleFunc("/ws", c.handleWebSockets).Queries("source", "{source}", "fileId", "{fileId}").Methods("GET")
	gui.PathPrefix("/").HandlerFunc(c.notFound)
}

// httpAPIHandlers initializes API routes.
func (c *Client) httpAPIHandlers() {
	c.apps.HandleAPIpath("", "info", c.triggers.CI.InfoHandler, "GET", "HEAD")
	c.apps.HandleAPIpath("", "version", c.triggers.CI.VersionHandler, "GET", "HEAD")
	c.apps.HandleAPIpath("", "version/{app}/{instance:[0-9]+}", c.triggers.CI.VersionHandlerInstance, "GET", "HEAD")
	c.apps.HandleAPIpath("", "trigger/{trigger:[0-9a-z-]+}", c.triggers.APIHandler, "GET", "POST")
	c.apps.HandleAPIpath("", "trigger/{trigger:[0-9a-z-]+}/{content}", c.triggers.APIHandler, "GET", "POST")
	c.apps.HandleAPIpath("", "services/{action}", c.Services.APIHandler, "GET")
	c.apps.HandleAPIpath("", "triggers", c.triggers.HandleGetTriggers, "GET")
	c.apps.HandleAPIpath("", "ping", c.handleInstancePing, "GET")
	c.apps.HandleAPIpath("", "ping/{app:[a-z,]+}", c.handleInstancePing, "GET")
	c.apps.HandleAPIpath("", "ping/{app:[a-z]+}/{instance:[0-9]+}", c.handleInstancePing, "GET")

	// Aggregate handlers. Non-app specific.
	c.apps.HandleAPIpath("", "/trash/{app}", c.triggers.CFSync.Handler, "POST")

	if c.plexEnabled() {
		// Use first Plex instance for API handlers.
		plex := &c.apps.Plex[0]
		c.apps.HandleAPIpath(starr.Plex, "sessions", plex.HandleSessions, "GET")
		c.apps.HandleAPIpath(starr.Plex, "directory", plex.HandleDirectory, "GET")
		c.apps.HandleAPIpath(starr.Plex, "emptytrash/{key}", plex.HandleEmptyTrash, "GET")
		c.apps.HandleAPIpath(starr.Plex, "markwatched/{key}", plex.HandleMarkWatched, "GET")
		c.apps.HandleAPIpath(starr.Plex, "kill", plex.HandleKillSession, "GET").
			Queries("reason", "{reason:.*}", "sessionId", "{sessionId:.*}")

		tokens := c.plexTokenPattern()
		c.apps.Router.HandleFunc("/plex", c.PlexHandler).Methods("POST").Queries("token", tokens)
		c.apps.Router.HandleFunc("/", c.PlexHandler).Methods("POST").Queries("token", tokens)
		// Give it an api path to get around some proxies that block /plex.
		c.apps.Router.HandleFunc("/api/plex/post", c.PlexHandler).
			Methods("POST").Queries("token", tokens)

		if c.Config.URLBase != "/" {
			// Allow plex to use the base url too.
			c.apps.Router.HandleFunc(path.Join(c.Config.URLBase, "plex"), c.PlexHandler).
				Methods("POST").Queries("token", tokens)
			// Give it an api path to get around some proxies that block /plex.
			c.apps.Router.HandleFunc(path.Join(c.Config.URLBase, "api", "plex", "post"), c.PlexHandler).
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

	_, _ = response.Write([]byte("404: Not Found"))
}

// slash is the GET handler for /.
func (c *Client) slash(response http.ResponseWriter, request *http.Request) {
	if !strings.HasSuffix(request.URL.Path, "/") {
		http.Redirect(response, request, request.URL.Path+"/", http.StatusPermanentRedirect)
		return
	}

	c.indexPage(request.Context(), response, request)
}

// stripSecrets runs first to save a redacted URI in a special request header.
// The logger uses this special value to save a redacted URI in the log file.
func (c *Client) stripSecrets(next http.Handler) http.Handler {
	secrets := []string{c.Config.AppsConfig.APIKey}
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
		if xff := r.Header.Get("X-Forwarded-For"); xff == "" || !c.allow.Contains(r.RemoteAddr) {
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

// @Description	Returns true or false for 1 requested instance. True is up, false is down.
// @Summary		Ping 1 starr instance.
// @Tags			Client
// @Produce		json
// @Param			app			path		string												true	"Application"	Enums(lidarr, prowlarr, radarr, readarr, sonarr)
// @Param			instance	path		int64												true	"Application instance (1-index)."
// @Success		200			{object}	apps.ApiResponse{message=map[string]map[int]bool}	"map for app->instance->up"
// @Failure		404			{object}	string												"bad token or api key"
// @Router			/ping/{app}/{instance} [get]
// @Security		ApiKeyAuth
func _() {}

// @Description	Returns true or false for every instance for starr app requested. True is up, false is down.
// @Description	Multiple apps may be provided by separating them with a comma. ie /api/ping/radarr,sonarr
// @Summary		Ping all instances for 1 or more starr apps.
// @Tags			Client
// @Produce		json
// @Param			apps	path		string												true	"Application, comma separated"	Enums(lidarr, prowlarr, radarr, readarr, sonarr)
// @Success		200		{object}	apps.ApiResponse{message=map[string]map[int]bool}	"map for app->instance->up"
// @Failure		404		{object}	string												"bad token or api key"
// @Router			/ping/{apps} [get]
// @Security		ApiKeyAuth
//
//nolint:lll
func _() {}

// @Description	Returns true or false for each configured starr instance. True is up, false is down.
// @Summary		Ping all starr instances.
// @Tags			Client
// @Produce		json
// @Success		200	{object}	apps.ApiResponse{message=map[string]map[int]bool}	"map for app->instance->up"
// @Failure		404	{object}	string												"bad token or api key"
// @Router			/ping [get]
// @Security		ApiKeyAuth
func (c *Client) handleInstancePing(req *http.Request) (int, any) { //nolint:cyclop
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
			for idx := range c.apps.Lidarr {
				c.pingInstance(req.Context(), c.apps.Lidarr[idx], app, idx, instance, output)
			}
		case starr.Radarr.Lower():
			for idx := range c.apps.Radarr {
				c.pingInstance(req.Context(), c.apps.Radarr[idx], app, idx, instance, output)
			}
		case starr.Readarr.Lower():
			for idx := range c.apps.Readarr {
				c.pingInstance(req.Context(), c.apps.Readarr[idx], app, idx, instance, output)
			}
		case starr.Sonarr.Lower():
			for idx := range c.apps.Sonarr {
				c.pingInstance(req.Context(), c.apps.Sonarr[idx], app, idx, instance, output)
			}
		case starr.Prowlarr.Lower():
			for idx := range c.apps.Prowlarr {
				c.pingInstance(req.Context(), c.apps.Prowlarr[idx], app, idx, instance, output)
			}
		}
	}

	return http.StatusOK, output
}

func (c *Client) pingInstance(
	ctx context.Context,
	pinger mnd.InstancePinger,
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
