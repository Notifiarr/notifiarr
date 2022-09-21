package client

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/shirou/gopsutil/v3/disk"
)

const (
	minPasswordLen   = 9
	fileSourceLogs   = "logs"
	fileSourceConfig = "config"
)

// userNameValue is used a context value key.
type userNameValue int

//nolint:gochecknoglobals // used as context value key.
var userNameStr interface{} = userNameValue(1)

func (c *Client) checkAuthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		userName, dyanmic := c.getUserName(request)
		if userName != "" {
			ctx := context.WithValue(request.Context(), userNameStr, []interface{}{userName, dyanmic})
			next.ServeHTTP(response, request.WithContext(ctx))
		} else {
			http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
		}
	})
}

// getUserName returns the username and a bool if it's dynamic (not the one from the config file).
func (c *Client) getUserName(request *http.Request) (string, bool) {
	if userName := request.Context().Value(userNameStr); userName != nil {
		u, _ := userName.([]interface{})
		username, _ := u[0].(string)
		found, _ := u[1].(bool)

		return username, found
	}

	if userName := request.Header.Get("x-webauth-user"); userName != "" {
		// If the upstream is allowed and gave us a username header, use it.
		ip := strings.Trim(request.RemoteAddr[:strings.LastIndex(request.RemoteAddr, ":")], "[]")
		if c.Config.Allow.Contains(ip) {
			return userName, true
		}
	}

	cookie, err := request.Cookie("session")
	if err != nil {
		return "", false
	}

	cookieValue := make(map[string]string)
	if err = c.cookies.Decode("session", cookie.Value, &cookieValue); err != nil {
		return "", false
	}

	return cookieValue["username"], false
}

func (c *Client) setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"username": userName,
	}

	encoded, err := c.cookies.Encode("session", value)
	if err != nil {
		return
	}

	http.SetCookie(response, &http.Cookie{
		Name:  "session",
		Value: encoded,
		Path:  "/",
	})
}

func (c *Client) loginHandler(response http.ResponseWriter, request *http.Request) {
	loggedinUsername, _ := c.getUserName(request)
	providedUsername := request.FormValue("name")

	switch {
	case loggedinUsername != "": // already logged in.
		http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
	case request.Method == http.MethodGet: // dont handle login without POST
		c.indexPage(request.Context(), response, request, "")
	case c.webauth:
		c.indexPage(request.Context(), response, request, "Logins Disabled")
	case len(request.FormValue("password")) < minPasswordLen:
		c.indexPage(request.Context(), response, request, "Invalid Password Length")
	case c.checkUserPass(providedUsername, request.FormValue("password")):
		c.setSession(providedUsername, response)
		exp.HTTPRequests.Add("GUI Logins", 1)
		http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
	default: // Start over.
		c.indexPage(request.Context(), response, request, "Invalid Password")
	}
}

func (c *Client) checkUserPass(username, password string) bool {
	c.Lock()
	defer c.Unlock()

	return c.Config.UIPassword.Valid(username + ":" + password)
}

func (c *Client) logoutHandler(response http.ResponseWriter, request *http.Request) {
	http.SetCookie(response, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
}

// getFileDeleteHandler deletes log and config files.
func (c *Client) getFileDeleteHandler(response http.ResponseWriter, req *http.Request) {
	var fileInfos *logs.LogFileInfos

	backupPath := filepath.Join(filepath.Dir(c.Flags.ConfigFile), "backups", filepath.Base(c.Flags.ConfigFile))

	switch mux.Vars(req)["source"] {
	case fileSourceLogs:
		fileInfos = c.Logger.GetAllLogFilePaths()
	case fileSourceConfig:
		fileInfos = logs.GetFilePaths(c.Flags.ConfigFile, backupPath)
	default:
		http.Error(response, "invalid source", http.StatusBadRequest)
		return
	}

	id := mux.Vars(req)["id"]

	for _, fileInfo := range fileInfos.List {
		if fileInfo.ID != id {
			continue
		}

		user, _ := c.getUserName(req)

		if err := os.Remove(fileInfo.Path); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			c.Errorf("[gui '%s' requested] Deleting file: %v", user, err)
		}

		c.Printf("[gui '%s' requested] Deleted file: %s", user, fileInfo.Path)

		if _, err := response.Write([]byte("ok")); err != nil {
			c.Errorf("Writing HTTP Response: %v", err)
		}

		return
	}
}

// getFileDownloadHandler downloads config and log files.
func (c *Client) getFileDownloadHandler(response http.ResponseWriter, req *http.Request) {
	var fileInfos *logs.LogFileInfos

	switch mux.Vars(req)["source"] {
	case fileSourceLogs:
		fileInfos = c.Logger.GetAllLogFilePaths()
	case fileSourceConfig:
		fileInfos = logs.GetFilePaths(c.Flags.ConfigFile)
	default:
		http.Error(response, "invalid source", http.StatusBadRequest)
		return
	}

	id := mux.Vars(req)["id"]
	for _, fileInfo := range fileInfos.List {
		if fileInfo.ID != id {
			continue
		}

		zipWriter := zip.NewWriter(response)
		defer zipWriter.Close()

		fileOpen, err := os.Open(fileInfo.Path)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		defer fileOpen.Close()

		newZippedFile, err := zipWriter.Create(fileInfo.Name)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}

		response.Header().Set("Content-Disposition", "attachment; filename="+fileInfo.Name+".zip")
		response.Header().Set("Content-Type", "application/zip")

		if _, err := io.Copy(newZippedFile, fileOpen); err != nil {
			c.Errorf("Sending Zipped File %s: %v", fileInfo.Path, err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}

		user, _ := c.getUserName(req)
		c.Printf("[gui '%s' requested] Downloaded file: %s", user, fileInfo.Path)

		return
	}
}

func (c *Client) handleShutdown(response http.ResponseWriter, _ *http.Request) {
	defer func() {
		c.sigkil <- &update.Signal{Text: "shutdown gui triggered"}
	}()

	http.Error(response, "OK", http.StatusOK)
}

func (c *Client) handleReload(response http.ResponseWriter, _ *http.Request) {
	c.Lock()
	c.reloading = true
	c.Unlock()

	defer c.triggerConfigReload(website.EventGUI, "GUI Requested")
	http.Error(response, "OK", http.StatusOK)
}

func (c *Client) handlePing(response http.ResponseWriter, _ *http.Request) {
	c.RLock()
	defer c.RUnlock()

	if c.reloading {
		http.Error(response, "Reloading", http.StatusLocked)
	} else {
		http.Error(response, "OK", http.StatusOK)
	}
}

func (c *Client) handleServicesStopStart(response http.ResponseWriter, req *http.Request) {
	user, _ := c.getUserName(req)

	switch action := mux.Vars(req)["action"]; action {
	case "stop":
		c.Config.Services.Stop(req.Context())
		c.Printf("[gui '%s' requested] Service Checks Stopped", user)
		http.Error(response, "Service Checks Stopped", http.StatusOK)
	case "start":
		c.Config.Services.Start(req.Context())
		c.Printf("[gui '%s' requested] Service Checks Started", user)
		http.Error(response, "Service Checks Started", http.StatusOK)
	default:
		http.Error(response, "invalid action: "+action, http.StatusBadRequest)
	}
}

func (c *Client) handleServicesCheck(response http.ResponseWriter, req *http.Request) {
	svc := mux.Vars(req)["service"]
	if err := c.Config.Services.RunCheck(website.EventAPI, svc); err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	user, _ := c.getUserName(req)
	c.Printf("[gui '%s' requested] Check Service: %s", user, svc)
	http.Error(response, "Service Check Initiated", http.StatusOK)
}

// getFileHandler returns portions of a config or log file based on request paraeters.
func (c *Client) getFileHandler(response http.ResponseWriter, req *http.Request) {
	var fileInfos *logs.LogFileInfos

	switch mux.Vars(req)["source"] {
	case fileSourceLogs:
		fileInfos = c.Logger.GetAllLogFilePaths()
	case fileSourceConfig:
		fileInfos = logs.GetFilePaths(c.Flags.ConfigFile)
	default:
		http.Error(response, "invalid source", http.StatusBadRequest)
		return
	}

	skip, _ := strconv.Atoi(mux.Vars(req)["skip"])

	count, _ := strconv.Atoi(mux.Vars(req)["lines"])
	if count == 0 {
		count = 500
		skip = 0
	}

	for _, fileInfo := range fileInfos.List {
		if fileInfo.ID != mux.Vars(req)["id"] {
			continue
		}

		lines, err := getLinesFromFile(fileInfo.Path, mux.Vars(req)["sort"], count, skip)
		if err != nil {
			c.Errorf("Handling Log File Request: %v", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
		} else if fileInfo.Size == 0 {
			http.Error(response, "the file is empty", http.StatusInternalServerError)
		} else if _, err = response.Write(lines); err != nil {
			c.Errorf("Writing HTTP Response: %v", err)
		}

		return
	}
}

func (c *Client) handleProfilePost(response http.ResponseWriter, request *http.Request) {
	currUser, dynamic := c.getUserName(request)
	if dynamic {
		http.Error(response, "Dynamic accounts cannot make profile changes.", http.StatusBadRequest)
		return
	}

	currPass := request.PostFormValue("Password")

	if !c.checkUserPass(currUser, currPass) {
		http.Error(response, "Invalid existing (current) password provided.", http.StatusBadRequest)
		return
	}

	username := request.PostFormValue("NewUsername")
	if username == "" {
		username = currUser
	}

	newPassw := request.PostFormValue("NewPassword")
	if newPassw == "" {
		newPassw = currPass
	}

	if len(newPassw) < minPasswordLen {
		http.Error(response, fmt.Sprintf("New password must be at least %d characters.",
			minPasswordLen), http.StatusBadRequest)
		return
	}

	if newPassw == currPass && username == currUser {
		http.Error(response, "Values unchanged. Nothing to save.", http.StatusOK)
		return
	}

	if err := c.setUserPass(request.Context(), username, newPassw); err != nil {
		c.Errorf("[gui '%s' requested] Saving Config: %v", currUser, err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)

		return
	}

	c.Printf("[gui '%s' requested] Updated primary username and password, new username: %s", currUser, username)
	c.setSession(username, response)
	http.Error(response, "New username and/or password saved.", http.StatusOK)
}

func (c *Client) handleInstanceCheck(response http.ResponseWriter, request *http.Request) {
	configPostDecoder.RegisterConverter([]string{}, func(input string) reflect.Value {
		return reflect.ValueOf(strings.Fields(input))
	})

	if err := request.ParseForm(); err != nil {
		http.Error(response, "Parsing form data failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	c.testInstance(response, request)
}

// handleFileBrowser returns a list of files and folders in a path.
// part of the file browser javascript code.
func (c *Client) handleFileBrowser(response http.ResponseWriter, request *http.Request) {
	type dir struct {
		Dirs  []string `json:"dirs"`
		Files []string `json:"files"`
	}

	output := dir{Dirs: []string{}, Files: []string{}}

	switch dirPath := mux.Vars(request)["dir"]; {
	case dirPath != "":
		dir, err := os.ReadDir(filepath.Join(dirPath, "/"))
		if err != nil {
			http.Error(response, err.Error(), http.StatusNotAcceptable)
			return
		}

		for _, file := range dir {
			if file.IsDir() {
				output.Dirs = append(output.Dirs, file.Name())
			} else {
				output.Files = append(output.Files, file.Name())
			}
		}
	case runtime.GOOS == mnd.Windows:
		partitions, err := disk.PartitionsWithContext(request.Context(), false)
		if err != nil {
			c.Errorf("Getting disk partitions: %v", err)
		}
		// this runs anyway.
		for _, partition := range partitions {
			output.Dirs = append(output.Dirs, partition.Mountpoint)
		}
	default:
		output.Dirs = []string{"/"}
	}

	if err := json.NewEncoder(response).Encode(&output); err != nil {
		c.Errorf("Encoding file browser directory: %v", err)
	}
}

func (c *Client) handleCommandStats(response http.ResponseWriter, request *http.Request) {
	cID, _ := strconv.Atoi(mux.Vars(request)["command"])
	if cID < 0 || cID >= len(c.triggers.Commands.List()) {
		http.Error(response, "Invalid command ID provided", http.StatusBadRequest)
	}

	cmd := c.triggers.Commands.List()[cID]
	if err := c.templat.ExecuteTemplate(response, "ajax/cmdstats.html", cmd); err != nil {
		http.Error(response, "template error: "+err.Error(), http.StatusOK)
	}
}

// handleProcessList just returns the running process list for a human to view.
func (c *Client) handleProcessList(response http.ResponseWriter, request *http.Request) {
	if ps, err := getProcessList(request.Context()); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	} else if _, err = ps.WriteTo(response); err != nil {
		user, _ := c.getUserName(request)
		c.Errorf("[gui '%s' requested] Writing HTTP Response: %v", user, err)
	}
}

func (c *Client) handleStartFileWatcher(response http.ResponseWriter, request *http.Request) {
	idx, err := strconv.Atoi(mux.Vars(request)["index"])
	if err != nil {
		http.Error(response, "invalid index provided:"+mux.Vars(request)["index"], http.StatusBadRequest)
		return
	}

	if idx < 0 || idx >= len(c.triggers.FileWatch.Files()) {
		http.Error(response, "unknown index provided:"+mux.Vars(request)["index"], http.StatusBadRequest)
		return
	}

	watch := c.triggers.FileWatch.Files()[idx]
	if watch.Active() {
		http.Error(response, "Watcher already running! "+watch.Path, http.StatusNotAcceptable)
		return
	}

	if err := c.triggers.FileWatch.AddFileWatcher(watch); err != nil {
		http.Error(response, "Start Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Error(response, "Started: "+watch.Path, http.StatusOK)
}

func (c *Client) handleStopFileWatcher(response http.ResponseWriter, request *http.Request) {
	idx, err := strconv.Atoi(mux.Vars(request)["index"])
	if err != nil {
		http.Error(response, "invalid index provided:"+mux.Vars(request)["index"], http.StatusBadRequest)
		return
	}

	if idx < 0 || idx >= len(c.triggers.FileWatch.Files()) {
		http.Error(response, "unknown index provided:"+mux.Vars(request)["index"], http.StatusBadRequest)
		return
	}

	watch := c.triggers.FileWatch.Files()[idx]
	if !watch.Active() {
		http.Error(response, "Watcher already stopped! "+watch.Path, http.StatusNotAcceptable)
		return
	}

	if err := watch.Stop(); err != nil {
		http.Error(response, "Stop Failed: "+err.Error(), http.StatusInternalServerError)

		user, _ := c.getUserName(request)
		c.Errorf("[gui '%s' requested] Stopping File Watcher: %v", user, err)

		return
	}

	http.Error(response, "Stopped: "+watch.Path, http.StatusOK)
}

// handleRegexTest tests a regular expression.
func (c *Client) handleRegexTest(response http.ResponseWriter, request *http.Request) {
	regex := request.PostFormValue("regexTestRegex")
	line := request.PostFormValue("regexTestLine")

	switch reg, err := regexp.Compile(regex); {
	case err != nil:
		http.Error(response, "Regex Parse Failed: "+err.Error(), http.StatusNotAcceptable)
	case regex == "":
		http.Error(response, "Regular Expression is blank!", http.StatusBadRequest)
	case reg.MatchString(line):
		http.Error(response, "Regular Expression matches! Found: "+reg.FindString(line), http.StatusOK)
	default:
		http.Error(response, "Regular Expression does not match!", http.StatusBadRequest)
	}
}

// handleConfigPost handles the reconfig endpoint.
func (c *Client) handleConfigPost(response http.ResponseWriter, request *http.Request) {
	user, _ := c.getUserName(request)
	// copy running config,
	config, err := c.Config.CopyConfig()
	if err != nil {
		c.Errorf("[gui '%s' requested] Copying Config (GUI request): %v", user, err)
		http.Error(response, "Error copying internal configuration (this is a bug): "+
			err.Error(), http.StatusInternalServerError)

		return
	}

	// update config.
	if err = c.mergeAndValidateNewConfig(config, request); err != nil {
		c.Errorf("[gui '%s' requested] Validating POSTed Config: %v", user, err)
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	// Check app integration configs before saving.
	config.Apps.Logger = c.Logger
	if err := config.Apps.Setup(); err != nil {
		http.Error(response, err.Error(), http.StatusNotAcceptable)
		return
	}

	if err := c.saveNewConfig(request.Context(), config); err != nil {
		c.Errorf("[gui '%s' requested] Saving Config: %v", user, err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)

		return
	}

	// reload.
	defer c.triggerConfigReload(website.EventGUI, "GUI Requested")

	c.Lock()
	c.reloading = true
	c.Unlock()

	// respond.
	c.Printf("[gui '%s' requested] Updated Configuration.", user)
	http.Error(response, "Config Saved. Reloading in 5 seconds...", http.StatusOK)
}

// saveNewConfig takes a fully built (copy) of config data, and saves it as the config file.
func (c *Client) saveNewConfig(ctx context.Context, config *configfile.Config) error {
	date := time.Now().Format("20060102T150405") // for file names.

	// make config file backup.
	if err := configfile.BackupFile(c.Flags.ConfigFile); err != nil {
		return fmt.Errorf("backing up config file: %w", err)
	}

	// write new config file to temporary path.
	destFile := filepath.Join(filepath.Dir(c.Flags.ConfigFile), "_tmpConfig."+date)
	if _, err := config.Write(ctx, destFile); err != nil { // write our config file template.
		return fmt.Errorf("writing new config file: %w", err)
	}

	// move new config file to existing config file.
	if err := os.Rename(destFile, c.Flags.ConfigFile); err != nil {
		return fmt.Errorf("renaming temporary file: %w", err)
	}

	return nil
}

// Set a Decoder instance as a package global, because it caches
// meta-data about structs, and an instance can be shared safely.
var configPostDecoder = schema.NewDecoder() //nolint:gochecknoglobals

func (c *Client) mergeAndValidateNewConfig(config *configfile.Config, request *http.Request) error {
	// This turns text fields into a []string (extra keys and upstreams use this).
	configPostDecoder.RegisterConverter([]string{}, func(input string) reflect.Value {
		return reflect.ValueOf(strings.Fields(input))
	})

	if err := request.ParseForm(); err != nil {
		return fmt.Errorf("parsing form data failed: %w", err)
	}

	if config.Snapshot == nil {
		config.Snapshot = &snapshot.Config{}
	}

	if config.Snapshot.Plugins == nil {
		config.Snapshot.Plugins = &snapshot.Plugins{}
	}

	if config.Apps != nil {
		config.Apps.Lidarr = nil
		config.Apps.Prowlarr = nil
		config.Apps.Radarr = nil
		config.Apps.Readarr = nil
		config.Apps.Sonarr = nil
		config.Apps.Qbit = nil
		config.Apps.Rtorrent = nil
		config.Apps.Deluge = nil
		config.Apps.SabNZB = nil
		config.Apps.NZBGet = nil
		config.Apps.Tautulli = nil
	}

	config.Plex = nil
	config.WatchFiles = nil
	config.Commands = nil
	config.Service = nil
	config.Snapshot.Plugins.MySQL = nil

	if err := configPostDecoder.Decode(config, request.PostForm); err != nil {
		return fmt.Errorf("decoding POST data into Go data structure failed: %w", err)
	}

	return c.validateNewServiceConfig(config)
}

func (c *Client) validateNewServiceConfig(config *configfile.Config) error {
	// Check service checks for non-unique names.
	serviceNames := make(map[string]struct{})
	index := 0

	for _, svc := range config.Service {
		if svc == nil {
			continue
		}

		config.Service[index] = svc
		index++

		if _, ok := serviceNames[svc.Name]; ok {
			return fmt.Errorf("%w (%d): %s", services.ErrNoName, index+1, svc.Name)
		}

		if err := svc.Validate(); err != nil {
			return fmt.Errorf("validating service check %d: %w", index+1, err)
		}

		serviceNames[svc.Name] = struct{}{}
	}

	// Clean up to avoid leaking memory.
	for j := index; j < len(config.Service); j++ {
		config.Service[j] = nil
	}

	config.Service = config.Service[:index]

	return nil
}

func (c *Client) indexPage(ctx context.Context, response http.ResponseWriter, request *http.Request, msg string) {
	response.Header().Add("content-type", "text/html")

	user, _ := c.getUserName(request)
	if request.Method != http.MethodGet || (user == "" && c.webauth) {
		response.WriteHeader(http.StatusUnauthorized)
	}

	c.renderTemplate(ctx, response, request, "index.html", msg)
}

func (c *Client) getTemplatePageHandler(response http.ResponseWriter, req *http.Request) {
	c.renderTemplate(req.Context(), response, req, mux.Vars(req)["template"]+".html", "")
}

// handleStaticAssets checks for a file on disk then falls back to compiled-in files.
func (c *Client) handleStaticAssets(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/files/css/custom.css" {
		if cssFileDir := c.haveCustomFile("custom.css"); cssFileDir != "" {
			// custom css file exists on disk, use http.FileServer to serve the dir it's in.
			http.StripPrefix("/files/css", http.FileServer(http.Dir(filepath.Dir(cssFileDir)))).ServeHTTP(response, request)
			return
		}
	}

	if c.Flags.Assets == "" {
		c.handleInternalAsset(response, request)
		return
	}

	// get the absolute path to prevent directory traversal
	f, err := filepath.Abs(filepath.Join(c.Flags.Assets, request.URL.Path))
	if _, err2 := os.Stat(f); err != nil || err2 != nil { // Check if it exists.
		c.handleInternalAsset(response, request)
		return
	}

	// file exists on disk, use http.FileServer to serve the static dir it's in.
	http.FileServer(http.Dir(c.Flags.Assets)).ServeHTTP(response, request)
}

func (c *Client) handleInternalAsset(response http.ResponseWriter, request *http.Request) {
	data, err := bindata.Asset(request.URL.Path[1:])
	if err != nil {
		http.Error(response, err.Error(), http.StatusNotFound)
		return
	}

	mime := mime.TypeByExtension(path.Ext(request.URL.Path))
	response.Header().Set("content-type", mime)

	if _, err = response.Write(data); err != nil {
		c.Errorf("Writing HTTP Response: %v", err)
	}
}
