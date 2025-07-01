package client

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/frontend"
	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/checkapp"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
	"github.com/shirou/gopsutil/v4/disk"
	"golift.io/cnfgfile"
)

// @title Notifiarr Client GUI API Documentation
// @description Monitors local services and sends notifications.
// @termsOfService https://notifiarr.com
// @contact.name Notifiarr Discord
// @contact.url https://notifiarr.com/discord
// @license.name MIT
// @license.url https://github.com/Notifiarr/notifiarr/blob/main/LICENSE
// @BasePath /

const (
	minPasswordLen = 9
	fileSourceLogs = "logs"
)

// userNameValue is used as a context value key.
type userNameValue int

//nolint:gochecknoglobals // used as context value key.
var userNameStr = userNameValue(1)

var ErrConfigVersionMismatch = errors.New("config version mismatch")

func (c *Client) checkAuthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		userName, dynamic := c.getUserName(request)
		if userName != "" {
			ctx := context.WithValue(request.Context(), userNameStr, []any{userName, dynamic})
			next.ServeHTTP(response, request.WithContext(ctx))
		} else {
			http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
		}
	})
}

// getUserName returns the username and a bool if it's dynamic (not the one from the config file).
func (c *Client) getUserName(request *http.Request) (string, bool) {
	if userName := request.Context().Value(userNameStr); userName != nil {
		u, _ := userName.([]any)
		username, _ := u[0].(string)
		found, _ := u[1].(bool)

		return username, found
	}

	if c.allow.Contains(request.RemoteAddr) && c.webauth {
		// If the upstream is allowed and gave us a username header, use it.
		if userName := request.Header.Get(c.authHeader); userName != "" {
			return userName, true
		}

		// If the upstream IP is allowed and no auth is enabled, set a username.
		if c.noauth { // c.webauth is always true if c.noauth is true.
			return configfile.DefaultUsername, true
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

func (c *Client) setSession(userName string, response http.ResponseWriter, request *http.Request) *http.Request {
	value := map[string]string{
		"username": userName,
	}

	encoded, err := c.cookies.Encode("session", value)
	if err != nil {
		return request
	}

	http.SetCookie(response, &http.Cookie{
		Name:  "session",
		Value: encoded,
		Path:  "/",
	})

	return request.WithContext(context.WithValue(request.Context(), userNameStr, []any{userName, true}))
}

func (c *Client) loginHandler(response http.ResponseWriter, request *http.Request) {
	loggedinUsername, _ := c.getUserName(request)
	providedUsername := request.FormValue("name")
	switch {
	case loggedinUsername != "": // already logged in.
		http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
	case request.Method == http.MethodGet: // dont handle login without POST
		c.indexPage(request.Context(), response, request)
	case c.webauth:
		c.indexPage(request.Context(), response, request)
	case len(request.FormValue("password")) < minPasswordLen:
		c.indexPage(request.Context(), response, request)
	case c.checkUserPass(providedUsername, request.FormValue("password")):
		request = c.setSession(providedUsername, response, request)
		mnd.HTTPRequests.Add("GUI Logins", 1)

		c.handleProfile(response, request)

	default: // Start over.
		c.indexPage(request.Context(), response, request)
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
	if mux.Vars(req)["source"] != fileSourceLogs {
		http.Error(response, "invalid source", http.StatusBadRequest)
		return
	}

	fileInfos := logs.Log.GetAllLogFilePaths()
	id := mux.Vars(req)["id"]

	for _, fileInfo := range fileInfos.List {
		if fileInfo.ID != id {
			continue
		}

		user, _ := c.getUserName(req)

		if err := os.Remove(fileInfo.Path); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			logs.Log.Errorf("[gui '%s' requested] Deleting file: %v", user, err)
		}

		logs.Log.Printf("[gui '%s' requested] Deleted file: %s", user, fileInfo.Path)

		if _, err := response.Write([]byte("ok")); err != nil {
			logs.Log.Errorf("Writing HTTP Response: %v", err)
		}

		return
	}
}

// uploadFileHandler uploads a log file to notifiarr.com.
func (c *Client) uploadFileHandler(response http.ResponseWriter, req *http.Request) {
	if mux.Vars(req)["source"] != fileSourceLogs {
		http.Error(response, "invalid source", http.StatusBadRequest)
		return
	}

	id := mux.Vars(req)["id"]
	for _, fileInfo := range logs.Log.GetAllLogFilePaths().List {
		if fileInfo.ID != id {
			continue
		}

		err := c.triggers.FileUpload.Upload(website.EventGUI, fileInfo.Path)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}

		user, _ := c.getUserName(req)
		logs.Log.Printf("[gui '%s' requested] Uploaded file: %s", user, fileInfo.Path)

		return
	}
}

// getFileDownloadHandler downloads log files to the browser.
func (c *Client) getFileDownloadHandler(response http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	for _, fileInfo := range logs.Log.GetAllLogFilePaths().List {
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
			logs.Log.Errorf("Sending Zipped File %s: %v", fileInfo.Path, err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}

		user, _ := c.getUserName(req)
		logs.Log.Printf("[gui '%s' requested] Downloaded file: %s", user, fileInfo.Path)

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
	c.reloadAppNow()
	http.Error(response, "OK", http.StatusOK)
}

func (c *Client) reloadAppNow() {
	c.Lock()
	c.reloading = true
	c.Unlock()

	defer c.triggerConfigReload(website.EventGUI, "GUI Requested")
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
		c.Services.Stop()
		logs.Log.Printf("[gui '%s' requested] Service Checks Stopped", user)
		http.Error(response, "Service Checks Stopped", http.StatusOK)
	case "start":
		c.Services.Start(req.Context(), c.apps.Plex.Name())
		logs.Log.Printf("[gui '%s' requested] Service Checks Started", user)
		http.Error(response, "Service Checks Started", http.StatusOK)
	default:
		http.Error(response, "invalid action: "+action, http.StatusBadRequest)
	}
}

func (c *Client) handleServicesCheck(response http.ResponseWriter, req *http.Request) {
	svc := mux.Vars(req)["service"]
	if err := c.Services.RunCheck(website.EventAPI, svc); err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	user, _ := c.getUserName(req)
	logs.Log.Printf("[gui '%s' requested] Check Service: %s", user, svc)
	http.Error(response, "Service Check Initiated", http.StatusOK)
}

// getFileHandler returns portions of a config or log file based on request parameters.
func (c *Client) getFileHandler(response http.ResponseWriter, req *http.Request) {
	var fileInfos *logs.LogFileInfos

	switch mux.Vars(req)["source"] {
	case fileSourceLogs:
		fileInfos = logs.Log.GetAllLogFilePaths()
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
			logs.Log.Errorf("Handling Log File Request: %v", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
		} else if fileInfo.Size == 0 {
			http.Error(response, "the file is empty", http.StatusInternalServerError)
		} else if _, err = response.Write(lines); err != nil {
			logs.Log.Errorf("Writing HTTP Response: %v", err)
		}

		return
	}

	logs.Log.Errorf("Handling Log File Request: file ID not found: %s", mux.Vars(req)["id"])
	http.Error(response, "no file found", http.StatusNotFound)
}

func (c *Client) handleInstanceCheck(response http.ResponseWriter, request *http.Request) {
	mnd.ConfigPostDecoder.RegisterConverter([]string{}, func(input string) reflect.Value {
		return reflect.ValueOf(strings.Fields(input))
	})

	if err := request.ParseForm(); err != nil {
		http.Error(response, "Parsing form data failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	checkapp.Test(c.Config, response, request)
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
			logs.Log.Errorf("Getting disk partitions: %v", err)
		}
		// this runs anyway.
		for _, partition := range partitions {
			output.Dirs = append(output.Dirs, partition.Mountpoint)
		}
	default:
		output.Dirs = []string{"/"}
	}

	if err := json.NewEncoder(response).Encode(&output); err != nil {
		logs.Log.Errorf("Encoding file browser directory: %v", err)
	}
}

// handleCommandStats is for js getCmdStats.
func (c *Client) handleCommandStats(response http.ResponseWriter, request *http.Request) {
	cmd := c.triggers.Commands.GetByHash(mux.Vars(request)["hash"])
	if cmd == nil {
		http.Error(response, "Invalid command Hash provided", http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(response).Encode(cmd.Stats()); err != nil {
		logs.Log.Errorf("Encoding command stats: %v", err)
	}
}

// handleRunCommand only handles commands with arguments.
// Commands without arguments are handled as an instance test.
func (c *Client) handleRunCommand(response http.ResponseWriter, request *http.Request) {
	cmd := c.triggers.Commands.GetByHash(mux.Vars(request)["hash"])
	if cmd == nil {
		http.Error(response, "Invalid command Hash provided", http.StatusBadRequest)
		return
	}

	_ = request.ParseForm()

	cmd.Run(&common.ActionInput{
		Type: website.EventGUI,
		Args: request.PostForm["args"],
	})
	http.Error(response, "Check command output after a few seconds.", http.StatusOK)
}

// handleProcessList just returns the running process list for a human to view.
func (c *Client) handleProcessList(response http.ResponseWriter, request *http.Request) {
	if ps, err := getProcessList(request.Context()); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	} else if _, err = ps.WriteTo(response); err != nil {
		user, _ := c.getUserName(request)
		logs.Log.Errorf("[gui '%s' requested] Writing HTTP Response: %v", user, err)
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
		logs.Log.Errorf("[gui '%s' requested] Stopping File Watcher: %v", user, err)

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

	// update config.
	config := &configfile.Config{}

	if err := json.NewDecoder(request.Body).Decode(&config); err != nil {
		logs.Log.Errorf("[gui '%s' requested] Decoding POSTed Config: %v, %#v", user, err, config)
		http.Error(response, "Error decoding POSTed Config: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validateNewConfig(config); err != nil {
		logs.Log.Errorf("[gui '%s' requested] Validating POSTed Config: %v", user, err)
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	// Check app integration configs before saving.
	if err := apps.CheckURLs(&config.AppsConfig); err != nil {
		http.Error(response, err.Error(), http.StatusNotAcceptable)
		return
	}

	if err := c.saveNewConfig(request.Context(), config); err != nil {
		logs.Log.Errorf("[gui '%s' requested] Saving Config: %v", user, err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)

		return
	}

	// reload.
	reload := " Not Reloading!"

	if mux.Vars(request)["noreload"] != "true" {
		defer c.triggerConfigReload(website.EventGUI, "GUI Requested")

		c.Lock()
		c.reloading = true
		c.Unlock()

		reload = "Reloading in 5 seconds..."
	}

	// respond.
	logs.Log.Printf("[gui '%s' requested] Updated Configuration.%s", user, reload)
	http.Error(response, "Config Saved."+reload, http.StatusOK)
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
	if _, err := config.Write(ctx, destFile, true); err != nil { // write our config file template.
		return fmt.Errorf("writing new config file: %w", err)
	}

	// move new config file to existing config file.
	if err := os.Rename(destFile, c.Flags.ConfigFile); err != nil {
		return fmt.Errorf("renaming temporary file: %w", err)
	}

	return nil
}

func (c *Client) validateNewConfig(config *configfile.Config) error {
	if config.Version != c.Config.Version {
		return fmt.Errorf("%w: provided: %d, running: %d",
			ErrConfigVersionMismatch, config.Version, c.Config.Version)
	}

	for idx, cmd := range config.Commands {
		if err := cmd.SetupRegexpArgs(); err != nil {
			return fmt.Errorf("command %d '%s' failed setup: %w", idx+1, cmd.Name, err)
		}
	}

	if err := c.validateNewServiceConfig(config); err != nil {
		return err
	}

	copied, err := config.CopyConfig()
	if err != nil {
		return fmt.Errorf("copying config: %w", err)
	}

	_, err = cnfgfile.Parse(copied, &cnfgfile.Opts{
		Name:          mnd.Title,
		TransformPath: configfile.ExpandHomedir,
		Prefix:        "filepath:",
	})
	if err != nil {
		return fmt.Errorf("filepath: %w", err)
	}

	return nil
}

func (c *Client) validateNewServiceConfig(config *configfile.Config) error {
	// Check service checks for non-unique names.
	serviceNames := make(map[string]struct{})
	index := 0

	for _, svc := range config.Service {
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

	config.Service = config.Service[:index]

	return nil
}

func (c *Client) indexPage(_ context.Context, response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "text/html; charset=utf-8")

	user, _ := c.getUserName(request)
	if request.Method != http.MethodGet || (user == "" && c.webauth) {
		response.WriteHeader(http.StatusUnauthorized)
	}

	frontend.IndexHandler(response, request)
}
