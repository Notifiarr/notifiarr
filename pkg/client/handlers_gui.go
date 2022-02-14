package client

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"golift.io/version"
)

const minPasswordLen = 16

type templateData struct {
	Config      *configfile.Config `json:"config"`
	Flags       *configfile.Flags  `json:"flags"`
	Username    string             `json:"username"`
	Data        url.Values         `json:"data,omitempty"`
	Msg         string             `json:"msg,omitempty"`
	Version     map[string]string  `json:"version"`
	LogFiles    *logs.LogFileInfos `json:"logFileInfo"`
	ConfigFiles *logs.LogFileInfos `json:"configFileInfo"`
}

// userNameValue is used a context value key.
type userNameValue string

const (
	defaultUsername               = "admin"
	userNameStr     userNameValue = "username"
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

// getUserPass turns the UIPassword config value into a usernam and password.
// "password." => user:admin, pass:password.
// ":password." => user:admin, pass::password.
// "joe:password." => user:joe, pass:password.
func (c *Client) getUserPass() (string, string) {
	c.RLock()
	defer c.RUnlock()

	username, password := defaultUsername, c.Config.UIPassword
	if spl := strings.SplitN(password, ":", 2); len(spl) == 2 { //nolint:gomnd
		password = spl[1]

		if spl[0] != "" {
			username = spl[0]
		}
	}

	return username, password
}

func (c *Client) setUserPass(username, password string) error {
	c.Lock()
	defer c.Unlock()

	current := c.Config.UIPassword
	c.Config.UIPassword = username + ":" + password

	if err := c.saveNewConfig(c.Config); err != nil {
		c.Config.UIPassword = current
		return err
	}

	return nil
}

func (c *Client) loginHandler(response http.ResponseWriter, request *http.Request) {
	validUsername, validPassword := c.getUserPass()

	switch providedUsername := request.FormValue("name"); {
	case len(validPassword) < minPasswordLen:
		c.indexPage(response, request, "Invalid Password Configured")
	case c.getUserName(request) != "":
		http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
	case request.Method == http.MethodGet:
		c.indexPage(response, request, "")
	case providedUsername == validUsername && validPassword == request.FormValue("password"):
		c.setSession(providedUsername, response)
		http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
	default: // Start over.
		c.indexPage(response, request, "Invalid Password")
	}
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

func (c *Client) getFileDeleteHandler(response http.ResponseWriter, req *http.Request) {
	var fileInfos *logs.LogFileInfos

	switch mux.Vars(req)["source"] {
	case "logs":
		fileInfos = c.Logger.GetAllLogFilePaths()
	case "config":
		fileInfos = logs.GetFilePaths(c.Flags.ConfigFile)
	}

	id := mux.Vars(req)["id"]
	for _, fileInfo := range fileInfos.List {
		if fileInfo.ID != id {
			continue
		}

		if err := os.Remove(fileInfo.Path); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}

		if _, err := response.Write([]byte("ok")); err != nil {
			c.Errorf("Writing HTTP Response: %v", err)
		}

		return
	}
}

func (c *Client) getFileDownloadHandler(response http.ResponseWriter, req *http.Request) {
	var fileInfos *logs.LogFileInfos

	switch mux.Vars(req)["source"] {
	case "logs":
		fileInfos = c.Logger.GetAllLogFilePaths()
	case "config":
		fileInfos = logs.GetFilePaths(c.Flags.ConfigFile)
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

		return
	}
}

func (c *Client) getFileHandler(response http.ResponseWriter, req *http.Request) {
	var fileInfos *logs.LogFileInfos

	switch mux.Vars(req)["source"] {
	case "logs":
		fileInfos = c.Logger.GetAllLogFilePaths()
	case "config":
		fileInfos = logs.GetFilePaths(c.Flags.ConfigFile)
	}

	id := mux.Vars(req)["id"]
	skip, _ := strconv.Atoi(mux.Vars(req)["skip"])

	count, _ := strconv.Atoi(mux.Vars(req)["lines"])
	if count == 0 {
		count = 500
		skip = 0
	}

	for _, fileInfo := range fileInfos.List {
		if fileInfo.ID != id {
			continue
		}

		lines, err := getLastLinesInFile(fileInfo.Path, count, skip)
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
	realUser, realPass := c.getUserPass()
	if realPass != request.PostFormValue("Password") {
		http.Error(response, "Invalid existing (current) password provided.", http.StatusBadRequest)
		return
	}

	username := request.PostFormValue("NewUsername")
	if username == "" {
		username = realUser
	}

	newPassw := request.PostFormValue("NewPassword")
	if newPassw == "" {
		newPassw = realPass
	}

	if len(newPassw) < minPasswordLen {
		http.Error(response, fmt.Sprintf("New password must be at least %d characters.",
			minPasswordLen), http.StatusBadRequest)
		return
	}

	if newPassw == realPass && username == realUser {
		http.Error(response, "Values unchanged. Nothing to save.", http.StatusOK)
		return
	}

	if err := c.setUserPass(username, newPassw); err != nil {
		c.Errorf("[gui requested] Saving Config: %v", err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)

		return
	}

	if _, err := response.Write([]byte("New username and/or password saved.")); err != nil {
		c.Errorf("[gui requested] Writing HTTP Response: %v", err)
	}
}

func (c *Client) handleGUITrigger(response http.ResponseWriter, request *http.Request) {
	code, data := c.runTrigger(mux.Vars(request)["action"], mux.Vars(request)["content"])
	http.Error(response, data, code)
}

// handleProcessList just returns the running process list for a human to view.
func (c *Client) handleProcessList(response http.ResponseWriter, request *http.Request) {
	if ps, err := getProcessList(); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	} else if _, err = response.Write(ps.Bytes()); err != nil {
		c.Errorf("[gui requested] Writing HTTP Response: %v", err)
	}
}

func (c *Client) handleConfigPost(response http.ResponseWriter, request *http.Request) {
	// copy running config,
	config, err := c.Config.CopyConfig()
	if err != nil {
		c.Errorf("[gui requested] Copying Config (GUI request): %v", err)
		http.Error(response, "Error copying internal configuration (this is a bug): "+
			err.Error(), http.StatusInternalServerError)

		return
	}

	// update config.
	if err = c.mergeAndValidateNewConfig(config, request); err != nil {
		c.Errorf("[gui requested] Validating POSTed Config: %v", err)
		http.Error(response, err.Error(), http.StatusBadRequest)

		return
	}

	if err := c.saveNewConfig(config); err != nil {
		c.Errorf("[gui requested] Saving Config: %v", err)
		http.Error(response, "Saving Config: "+err.Error(), http.StatusInternalServerError)

		return
	}

	// reload.
	defer func() {
		c.sighup <- &update.Signal{Text: "reload gui triggered"}
	}()

	// respond.
	_, err = response.Write([]byte("Config Saved. Reloading in 5 seconds..."))
	if err != nil {
		c.Errorf("[gui requested] Writing HTTP Response: %v", err)
	}
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
		return fmt.Errorf("backing up config file: %w", err)
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

	err := c.templat.ExecuteTemplate(response, "index.html", &templateData{
		Config:      c.Config,
		Flags:       c.Flags,
		Username:    c.getUserName(request),
		Data:        request.PostForm,
		Msg:         msg,
		LogFiles:    c.Logger.GetAllLogFilePaths(),
		ConfigFiles: logs.GetFilePaths(c.Flags.ConfigFile),
		Version: map[string]string{
			"started":   version.Started.Round(time.Second).String(),
			"uptime":    time.Since(version.Started).Round(time.Second).String(),
			"program":   c.Flags.Name(),
			"version":   version.Version,
			"revision":  version.Revision,
			"branch":    version.Branch,
			"buildUser": version.BuildUser,
			"buildDate": version.BuildDate,
			"goVersion": version.GoVersion,
			"os":        runtime.GOOS,
			"arch":      runtime.GOARCH,
		},
	})
	if err != nil {
		c.Errorf("Sending HTTP Response: %v", err)
	}
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
