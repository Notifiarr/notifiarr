package client

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

const (
	minPasswordLen   = 16
	fileSourceLogs   = "logs"
	fileSourceConfig = "config"
)

// userNameValue is used a context value key.
type userNameValue string

const (
	userNameStr userNameValue = "username"
)

func (c *Client) checkAuthorized(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		userName := c.getUserName(request)
		if userName != "" {
			ctx := context.WithValue(request.Context(), userNameStr, userName)
			next.ServeHTTP(response, request.WithContext(ctx))
		} else {
			http.Redirect(response, request, path.Join(c.Config.URLBase, "login"), http.StatusFound)
		}
	})
}

func (c *Client) getUserName(request *http.Request) string {
	if userName := request.Context().Value(userNameStr); userName != nil {
		return userName.(string)
	}

	cookie, err := request.Cookie("session")
	if err != nil {
		return ""
	}

	cookieValue := make(map[string]string)
	if err = c.cookies.Decode("session", cookie.Value, &cookieValue); err != nil {
		return ""
	}

	return cookieValue["username"]
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
	switch providedUsername := request.FormValue("name"); {
	case c.getUserName(request) != "": // already logged in.
		http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
	case request.Method == http.MethodGet: // dont handle login without POST
		c.indexPage(response, request, "")
	case len(request.FormValue("password")) < minPasswordLen:
		c.indexPage(response, request, "Invalid Password Length")
	case c.checkUserPass(providedUsername, request.FormValue("password")):
		c.setSession(providedUsername, response)
		exp.HTTPRequests.Add("GUI Logins", 1)
		http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
	default: // Start over.
		c.indexPage(response, request, "Invalid Password")
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

		user := c.getUserName(req)

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

		c.Printf("[gui '%s' requested] Downloaded file: %s", c.getUserName(req), fileInfo.Path)

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
	defer c.triggerConfigReload(notifiarr.EventGUI, "GUI Requested")
	http.Error(response, "OK", http.StatusOK)
}

func (c *Client) handleServicesStopStart(response http.ResponseWriter, req *http.Request) {
	switch action := mux.Vars(req)["action"]; action {
	case "stop":
		c.Config.Services.Stop()
		c.Printf("[gui '%s' requested] Service Checks Stopped", c.getUserName(req))
		http.Error(response, "Service Checks Stopped", http.StatusOK)
	case "start":
		c.Config.Services.Start()
		c.Printf("[gui '%s' requested] Service Checks Started", c.getUserName(req))
		http.Error(response, "Service Checks Started", http.StatusOK)
	default:
		http.Error(response, "invalid action: "+action, http.StatusBadRequest)
	}
}

func (c *Client) handleServicesCheck(response http.ResponseWriter, req *http.Request) {
	svc := mux.Vars(req)["service"]
	if err := c.Config.Services.RunCheck(notifiarr.EventAPI, svc); err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	c.Printf("[gui '%s' requested] Check Service: %s", c.getUserName(req), svc)
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
	currUser := c.getUserName(request)
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

	if err := c.setUserPass(username, newPassw); err != nil {
		c.Errorf("[gui '%s' requested] Saving Config: %v", currUser, err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)

		return
	}

	c.Printf("[gui '%s' requested] Updated primary username and password, new username: %s", currUser, username)
	c.setSession(username, response)
	http.Error(response, "New username and/or password saved.", http.StatusOK)
}

func (c *Client) handleGUITrigger(response http.ResponseWriter, request *http.Request) {
	code, data := c.runTrigger(notifiarr.EventGUI, mux.Vars(request)["action"], mux.Vars(request)["content"])
	http.Error(response, data, code)
}

// handleProcessList just returns the running process list for a human to view.
func (c *Client) handleProcessList(response http.ResponseWriter, request *http.Request) {
	if ps, err := getProcessList(); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	} else if _, err = ps.WriteTo(response); err != nil {
		c.Errorf("[gui '%s' requested] Writing HTTP Response: %v", c.getUserName(request), err)
	}
}

func (c *Client) handleConfigPost(response http.ResponseWriter, request *http.Request) {
	user := c.getUserName(request)
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

	if err := c.saveNewConfig(config); err != nil {
		c.Errorf("[gui '%s' requested] Saving Config: %v", user, err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)

		return
	}

	// reload.
	defer c.triggerConfigReload(notifiarr.EventGUI, "GUI Requested")

	// respond.
	c.Printf("[gui '%s' requested] Updated Configuration.", user)
	http.Error(response, "Config Saved. Reloading in 5 seconds...", http.StatusOK)
}

// saveNewConfig takes a fully built (copy) of config data, and saves it as the config file.
func (c *Client) saveNewConfig(config *configfile.Config) error {
	date := time.Now().Format("20060102T150405") // for file names.

	// write new config file to temporary path.
	destFile := filepath.Join(filepath.Dir(c.Flags.ConfigFile), "_tmpConfig."+date)
	if _, err := config.Write(destFile); err != nil { // write our config file template.
		return fmt.Errorf("writing new config file: %w", err)
	}

	// make config file backup.
	bckupFile := filepath.Join(filepath.Dir(c.Flags.ConfigFile), "backup.notifiarr."+date+".conf")
	if err := configfile.CopyFile(c.Flags.ConfigFile, bckupFile); err != nil {
		notexist := os.ErrNotExist
		if !errors.As(err, &notexist) {
			return fmt.Errorf("backing up config file: %w", err)
		}
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
		config.Apps.Deluge = nil
		config.Apps.SabNZB = nil
	}

	config.Service = nil
	config.Snapshot.Plugins.MySQL = nil

	if err := configPostDecoder.Decode(config, request.PostForm); err != nil {
		return fmt.Errorf("decoding POST data into Go data structure failed: %w", err)
	}

	// Check service checks for non-unique names.
	serviceNames := make(map[string]struct{})
	for index, svc := range config.Service {
		if _, ok := serviceNames[svc.Name]; ok {
			return fmt.Errorf("%w (%d): %s", services.ErrNoName, index+1, svc.Name)
		}

		if err := svc.Validate(); err != nil {
			return fmt.Errorf("validating service check %d: %w", index+1, err)
		}

		serviceNames[svc.Name] = struct{}{}
	}

	return nil
}

func (c *Client) indexPage(response http.ResponseWriter, request *http.Request, msg string) {
	response.Header().Add("content-type", "text/html")

	if request.Method != http.MethodGet {
		response.WriteHeader(http.StatusUnauthorized)
	}

	c.renderTemplate(response, request, "index.html", msg)
}

func (c *Client) getTemplatePageHandler(response http.ResponseWriter, req *http.Request) {
	c.renderTemplate(response, req, mux.Vars(req)["template"]+".html", "")
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
