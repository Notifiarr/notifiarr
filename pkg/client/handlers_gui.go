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
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/frontend"
	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/Notifiarr/notifiarr/pkg/checkapp"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/dashboard"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/gorilla/mux"
	"github.com/shirou/gopsutil/v4/disk"
	"golift.io/cnfgfile"
	"golift.io/starr/lidarr"
	"golift.io/starr/prowlarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

//	@title			Notifiarr Client GUI API Documentation
//	@description	Monitors local services and sends notifications.
//	@termsOfService	https://notifiarr.com
//	@contact.name	Notifiarr Discord
//	@contact.url	https://notifiarr.com/discord
//	@license.name	MIT
//	@license.url	https://github.com/Notifiarr/notifiarr/blob/main/LICENSE
//	@BasePath		/ui

const fileSourceLogs = "logs"

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
	case request.Method != http.MethodPost: // dont handle login without POST
		c.indexPage(request.Context(), response, request)
	case c.webauth:
		c.indexPage(request.Context(), response, request)
	case c.Config.UIPassword.Valid(providedUsername, request.FormValue("password")):
		c.updateToNewPasswordMD5(request.Context(), loggedinUsername, providedUsername, request.FormValue("sha"))
		logs.Log.Printf("[gui '%s' requested] Updated config with new password format.", providedUsername)
		fallthrough
	case c.Config.UIPassword.Valid(providedUsername, request.FormValue("sha")):
		request = c.setSession(providedUsername, response, request)
		c.handleProfile(response, request)
		mnd.HTTPRequests.Add("GUI Logins", 1)
		logs.Log.Printf("[gui '%s' requested] Authenticated with local credentials", providedUsername)
	case clientinfo.CheckPassword(providedUsername, request.FormValue("sha")):
		providedUsername = clientinfo.Get().User.Username
		if providedUsername == "" {
			providedUsername = "admin"
		}

		request = c.setSession(providedUsername, response, request)
		c.handleProfile(response, request)
		mnd.HTTPRequests.Add("GUI Logins", 1)
		logs.Log.Printf("[gui '%s' requested] Authenticated with website credentials", providedUsername)
	default: // Start over.
		http.Error(response, "Unauthorized", http.StatusUnauthorized)
	}
}

// updateToNewPasswordMD5 saves the md5 version of the password to the config file.
// This is used to update the password from the old plaintext version to the new md5 version.
// In the future, the frontend will stop sending the plaintext password, and this will be removed.
func (c *Client) updateToNewPasswordMD5(ctx context.Context, loggedinUsername, providedUsername, password string) {
	if password == "" {
		return
	}

	logs.Log.Printf("[gui '%s' requested] Updating Trust Profile settings, username: %s",
		loggedinUsername, providedUsername)

	if err := c.setUserPass(ctx, configfile.AuthPassword, providedUsername, password); err != nil {
		logs.Log.Errorf("[gui '%s' requested] Setting user pass: %v", loggedinUsername, err)
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

// getFileDeleteHandler deletes log and config files.
//
//	@Summary		Delete files
//	@Description	Deletes a log or config file.
//	@Tags			Files
//	@Produce		text/plain
//	@Param			source	path		string	true	"log or config"
//	@Param			id		path		string	true	"file id"
//	@Success		200		{string}	string	"ok"
//	@Failure		400		{string}	string	"bad input"
//	@Failure		500		{string}	string	"error removing file"
//	@Router			/deleteFile/{source}/{id} [get]
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
//
//	@Summary		Upload files
//	@Description	Uploads a log file to notifiarr.com.
//	@Tags			Files
//	@Produce		text/plain
//	@Param			source	path		string	true	"log"
//	@Param			id		path		string	true	"file id"
//	@Success		200		{string}	string	"ok"
//	@Failure		400		{string}	string	"bad input"
//	@Failure		500		{string}	string	"error uploading file"
//	@Router			/uploadFile/{source}/{id} [get]
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

		if _, err := response.Write([]byte("ok")); err != nil {
			logs.Log.Errorf("Writing HTTP Response: %v", err)
		}

		return
	}
}

// getFileDownloadHandler downloads log files to the browser.
//
//	@Summary		Download files
//	@Description	Downloads a log file (to a browser) as a zip file.
//	@Tags			Files
//	@Produce		application/zip
//	@Param			source	path		string	true	"log or config"
//	@Param			id		path		string	true	"file id"
//	@Success		200		{object}	any		"zip file content"
//	@Failure		400		{string}	string	"bad input"
//	@Failure		500		{string}	string	"error opening file"
//	@Router			/downloadFile/{source}/{id} [get]
func (c *Client) getFileDownloadHandler(response http.ResponseWriter, req *http.Request) {
	if mux.Vars(req)["source"] != fileSourceLogs {
		http.Error(response, "invalid source", http.StatusBadRequest)
		return
	}

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

// handleShutdown initiates application shutdown.
//
//	@Summary		Shutdown application
//	@Description	Initiates graceful shutdown of the application.
//	@Tags			System
//	@Produce		text/plain
//	@Success		200	{string}	string	"OK"
//	@Router			/shutdown [get]
func (c *Client) handleShutdown(response http.ResponseWriter, _ *http.Request) {
	defer func() {
		c.sigkil <- &update.Signal{Text: "shutdown gui triggered"}
	}()

	http.Error(response, "OK", http.StatusOK)
}

// handleReload triggers a configuration reload.
//
//	@Summary		Reload configuration
//	@Description	Triggers an immediate reload of the application configuration.
//	@Tags			System
//	@Produce		text/plain
//	@Success		200	{string}	string	"OK"
//	@Router			/reload [get]
func (c *Client) handleReload(response http.ResponseWriter, _ *http.Request) {
	c.reloadAppNow()
	http.Error(response, "OK", http.StatusOK)
}

func (c *Client) reloadAppNow() {
	c.Lock()
	defer c.Unlock()
	defer c.triggerConfigReload(website.EventGUI, "GUI Requested")
}

// handlePing returns the application status.
//
//	@Summary		Ping application
//	@Description	Returns application status, indicating if it's reloading or running normally.
//	@Tags			System
//	@Produce		text/plain
//	@Success		200	{string}	string	"OK"
//	@Failure		423	{string}	string	"Reloading"
//	@Router			/ping [get]
func (c *Client) handlePing(response http.ResponseWriter, _ *http.Request) {
	c.RLock()
	defer c.RUnlock()

	if c.reloading {
		http.Error(response, "Reloading", http.StatusLocked)
	} else {
		http.Error(response, "OK", http.StatusOK)
	}
}

// handleServicesStopStart stops or starts service checks.
//
//	@Summary		Pause/resume service checks
//	@Description	Pauses or resumes all service checks.
//	@Tags			Integrations
//	@Produce		text/plain
//	@Param			action	path		string	true	"Action to perform"	Enums(stop, start)
//	@Success		200		{string}	string	"Service Checks Paused or Service Checks Resumed"
//	@Failure		400		{string}	string	"invalid action"
//	@Router			/services/{action} [get]
func (c *Client) handleServicesStopStart(response http.ResponseWriter, req *http.Request) {
	user, _ := c.getUserName(req)

	switch action := mux.Vars(req)["action"]; action {
	case "stop":
		c.Services.Pause()
		logs.Log.Printf("[gui '%s' requested] Service Checks Paused", user)
		http.Error(response, "Service Checks Paused", http.StatusOK)

		if menu["svcs"] != nil {
			menu["svcs"].Uncheck()
		}
	case "start":
		c.Services.Resume()
		logs.Log.Printf("[gui '%s' requested] Service Checks Resumed", user)
		http.Error(response, "Service Checks Resumed", http.StatusOK)

		if menu["svcs"] != nil {
			menu["svcs"].Check()
		}
	default:
		http.Error(response, "invalid action: "+action, http.StatusBadRequest)
	}
}

// handleServicesCheck runs a specific service check.
//
//	@Summary		Check service
//	@Description	Runs a specific service check by name.
//	@Tags			Integrations
//	@Produce		text/plain
//	@Param			service	path		string	true	"Service name to check"
//	@Success		200		{string}	string	"Service Check Initiated"
//	@Failure		400		{string}	string	"error running service check"
//	@Router			/services/check/{service} [get]
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
//
//	@Summary		Get file contents
//	@Description	Returns portions of a log or config file based on request parameters like lines, skip, and sort.
//	@Tags			Files
//	@Produce		text/plain
//	@Param			source	path		string	true	"log or config"
//	@Param			id		path		string	true	"file id"
//	@Param			lines	path		int		true	"number of lines to return"
//	@Param			skip	path		int		true	"number of lines to skip"
//	@Param			sort	query		string	false	"sort order (asc/desc)"
//	@Success		200		{string}	string	"file contents"
//	@Failure		400		{string}	string	"invalid source"
//	@Failure		500		{string}	string	"error reading file"
//	@Router			/getFile/{source}/{id}/{lines}/{skip} [get]
//	@Router			/getFile/{source}/{id}/{lines} [get]
//	@Router			/getFile/{source}/{id} [get]
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

// handleInstanceCheck validates instance configuration.
//
//	@Summary		Check instance configuration
//	@Description	Validates and tests the configuration for a specific application instance.
//	@Tags			Integrations
//	@Produce		text/plain
//	@Param			type	path		string	true	"Application type"
//	@Param			index	path		int		true	"Instance index"
//	@Success		200		{string}	string	"configuration check result"
//	@Failure		400		{string}	string	"parsing error"
//	@Router			/checkInstance/{type}/{index} [post]
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

// handleCheckAll checks all the starr, media and downloader apps and returns the results.
//
//	@Summary		Check all instances
//	@Description	Checks all the starr, media and downloader apps and returns the results.
//	@Tags			Integrations
//	@Produce		application/json
//	@Success		200	{object}	checkapp.CheckAllOutput	"all check results"
//	@Router			/checkAllInstances [get]
func (c *Client) handleCheckAll(response http.ResponseWriter, request *http.Request) {
	input := &checkapp.CheckAllInput{
		Sonarr:       c.Config.Sonarr,
		Radarr:       c.Config.Radarr,
		Readarr:      c.Config.Readarr,
		Lidarr:       c.Config.Lidarr,
		Prowlarr:     c.Config.Prowlarr,
		NZBGet:       c.Config.NZBGet,
		Deluge:       c.Config.Deluge,
		Qbit:         c.Config.Qbit,
		Rtorrent:     c.Config.Rtorrent,
		Transmission: c.Config.Transmission,
		SabNZB:       c.Config.SabNZB,
	}

	if c.Config.Plex.URL != "" && c.Config.Plex.Token != "" {
		input.Plex = []apps.PlexConfig{c.Config.Plex}
	}

	if c.Config.Tautulli.URL != "" && c.Config.Tautulli.APIKey != "" {
		input.Tautulli = []apps.TautulliConfig{c.Config.Tautulli}
	}

	output := checkapp.CheckAll(request.Context(), input)
	if err := json.NewEncoder(response).Encode(output); err != nil {
		logs.Log.Errorf("Encoding check all instances: %v", err)
	}
}

type BrowseDir struct {
	Sep   string   `json:"sep"`   // Filepath separator.
	Path  string   `json:"path"`  // Current directory path.
	Mom   string   `json:"mom"`   // Parent directory path.
	Dirs  []string `json:"dirs"`  // Directories in the current directory.
	Files []string `json:"files"` // Files in the current directory.
	Error string   `json:"error"` // Error message.
}

// handleFileBrowser returns a list of files and folders in a path.
// part of the file browser javascript code.
//
//	@Summary		Browse file system
//	@Description	Returns a list of files and folders in the specified directory path.
//	@Tags			Files
//	@Produce		json
//	@Param			dir	query		string		false	"Directory path to browse"
//	@Success		200	{object}	BrowseDir	"directory contents"
//	@Failure		406	{string}	string		"error reading directory"
//	@Router			/browse [get]
func (c *Client) handleFileBrowser(response http.ResponseWriter, request *http.Request) {
	output, err := c.getFileBrowserOutput(request.Context(), mux.Vars(request)["dir"])
	if err != nil {
		http.Error(response, err.Error(), http.StatusNotAcceptable)
	} else if err = json.NewEncoder(response).Encode(&output); err != nil {
		logs.Log.Errorf("Encoding file browser directory: %v", err)
	}
}

//nolint:cyclop
func (c *Client) getFileBrowserOutput(ctx context.Context, dirPath string) (*BrowseDir, error) {
	output, err := c.getBrowsedDir(dirPath)
	if err != nil {
		return nil, err
	}

	if (output.Path == "/" || output.Path == "" || output.Path == `\`) && mnd.IsWindows {
		partitions, err := disk.PartitionsWithContext(ctx, false)
		if err != nil {
			logs.Log.Errorf("Getting disk partitions: %v", err)
		}

		for _, partition := range partitions {
			output.Dirs = append(output.Dirs, partition.Mountpoint)
		}

		output.Mom = ``
		output.Path = ``

		return output, nil
	}

	if output.Path == "" || output.Path == `\` {
		output.Path = "/"
	}

	dir, err := os.ReadDir(filepath.Join(output.Path, string(filepath.Separator)))
	if err != nil {
		return nil, fmt.Errorf("unable to read content of provided path: %w", err)
	}

	for _, file := range dir {
		if file.IsDir() {
			output.Dirs = append(output.Dirs, file.Name())
		} else {
			output.Files = append(output.Files, file.Name())
		}
	}

	return output, nil
}

func (c *Client) getBrowsedDir(dir string) (*BrowseDir, error) {
	if dir = configfile.ExpandHomedir(dir); dir == "~" {
		if mnd.IsWindows {
			dir = ""
		} else {
			dir = "/"
		}
	}

	output := &BrowseDir{
		Path:  dir,
		Dirs:  []string{},
		Files: []string{},
		Sep:   string(filepath.Separator),
		Mom:   filepath.Dir(dir),
	}

	if dir == "" {
		output.Mom = ""
		return output, nil
	}

	dirStat, err := os.Stat(dir)
	if err != nil {
		output.Error = "unable to read provided path: " + err.Error()
		output.Path = filepath.Dir(dir)
		output.Mom = filepath.Dir(output.Path)

		if dirStat, err = os.Stat(output.Path); err != nil {
			return nil, fmt.Errorf("unable to read provided path: %w", err)
		}
	}

	if !dirStat.IsDir() {
		output.Path = filepath.Dir(dir)
		output.Mom = filepath.Dir(output.Path)
	}

	output.Mom = filepath.Dir(output.Path)
	// weird windows thing when at drive root.
	if (output.Mom == output.Path+"." || output.Mom == output.Path) && output.Path != "/" {
		output.Mom = ""
	}

	return output, nil
}

// handleNewFile creates a new file at the specified path.
// part of the file browser javascript code.
//
//	@Summary		Create file
//	@Description	Creates a new file at the specified path.
//	@Tags			Files
//	@Produce		json
//	@Param			file	query		string		true	"File path to create"
//	@Param			new		query		string		true	"Must be true to create a file"
//	@Success		200		{object}	BrowseDir	"directory contents where file was created"
//	@Failure		406		{string}	string		"error reading directory"
//	@Failure		500		{string}	string		"error creating file"
//	@Router			/browse [get]
func (c *Client) handleNewFile(response http.ResponseWriter, request *http.Request) {
	filePath := mux.Vars(request)["file"]
	logs.Log.Printf("[user requested] Creating file: %s", filePath)

	f, err := os.Create(filePath)
	if err != nil {
		http.Error(response, "unable to create file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	f.Close()

	output, err := c.getFileBrowserOutput(request.Context(), filepath.Dir(filePath))
	if err != nil {
		http.Error(response, err.Error(), http.StatusNotAcceptable)
	} else if err = json.NewEncoder(response).Encode(&output); err != nil {
		logs.Log.Errorf("Encoding file browser directory: %v", err)
	}
}

// handleNewFolder creates a new folder at the specified path.
// part of the file browser javascript code.
//
//	@Summary		Create folder
//	@Description	Creates a new folder at the specified path.
//	@Tags			Files
//	@Produce		json
//	@Param			dir	query		string		true	"Folder path to create"
//	@Param			new	query		string		true	"Must be true to create a folder"
//	@Success		200	{object}	BrowseDir	"directory contents where file was created"
//	@Failure		406	{string}	string		"error reading directory"
//	@Failure		500	{string}	string		"error creating folder"
//	@Router			/browse [get]
func (c *Client) handleNewFolder(response http.ResponseWriter, request *http.Request) {
	dirPath := mux.Vars(request)["dir"]
	logs.Log.Printf("[user requested] Creating folder: %s", dirPath)

	if err := os.MkdirAll(dirPath, mnd.Mode0750); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := c.getFileBrowserOutput(request.Context(), dirPath)
	if err != nil {
		http.Error(response, err.Error(), http.StatusNotAcceptable)
	} else if err = json.NewEncoder(response).Encode(&output); err != nil {
		logs.Log.Errorf("Encoding file browser directory: %v", err)
	}
}

// handleCommandStats is for js getCmdStats.
//
//	@Summary		Get command statistics
//	@Description	Returns execution statistics for a specific command.
//	@Tags			Integrations
//	@Produce		json
//	@Param			path	path		string	true	"Command path identifier"	Enums(cmdstats, cmdargs)
//	@Param			hash	path		string	true	"Command hash"
//	@Success		200		{object}	any		"command statistics"
//	@Failure		400		{string}	string	"invalid command hash"
//	@Router			/ajax/{path}/{hash} [get]
//
//nolint:dupword
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

// Integrations is the data returned by the UI integrations endpoint.
type Integrations struct {
	Snapshot         *snapshot.Snapshot `json:"snapshot"`
	SnapshotAge      time.Time          `json:"snapshotAge"`
	Plex             *plex.PMSInfo      `json:"plex"`
	PlexAge          time.Time          `json:"plexAge"`
	Sessions         *plex.Sessions     `json:"sessions"`
	SessionsAge      time.Time          `json:"sessionsAge"`
	Dashboard        *dashboard.States  `json:"dashboard"`
	DashboardAge     time.Time          `json:"dashboardAge"`
	TautulliUsers    *tautulli.Users    `json:"tautulliUsers"`
	TautulliUsersAge time.Time          `json:"tautulliUsersAge"`
	Tautulli         *tautulli.Info     `json:"tautulli"`
	TautulliAge      time.Time          `json:"tautulliAge"`
	Lidarr           struct {
		Status    []*lidarr.SystemStatus `json:"status"`
		StatusAge []time.Time            `json:"statusAge"`
		Queue     []*lidarr.Queue        `json:"queue"`
		QueueAge  []time.Time            `json:"queueAge"`
	} `json:"lidarr"`
	Radarr struct {
		Status    []*radarr.SystemStatus `json:"status"`
		StatusAge []time.Time            `json:"statusAge"`
		Queue     []*radarr.Queue        `json:"queue"`
		QueueAge  []time.Time            `json:"queueAge"`
	} `json:"radarr"`
	Readarr struct {
		Status    []*readarr.SystemStatus `json:"status"`
		StatusAge []time.Time             `json:"statusAge"`
		Queue     []*readarr.Queue        `json:"queue"`
		QueueAge  []time.Time             `json:"queueAge"`
	} `json:"readarr"`
	Sonarr struct {
		Status    []*sonarr.SystemStatus `json:"status"`
		StatusAge []time.Time            `json:"statusAge"`
		Queue     []*sonarr.Queue        `json:"queue"`
		QueueAge  []time.Time            `json:"queueAge"`
	} `json:"sonarr"`
	Prowlarr struct {
		Status    []*prowlarr.SystemStatus `json:"status"`
		StatusAge []time.Time              `json:"statusAge"`
	} `json:"prowlarr"`
}

// handleIntegrations returns the current integrations statuses and data.
//
//	@Summary		Get integrations status
//	@Description	Returns current status and data for all configured integrations including apps, Plex, Tautulli, etc.
//	@Tags			Integrations
//	@Produce		json
//	@Success		200	{object}	Integrations	"integrations status data"
//	@Router			/integrations [get]
//
//nolint:cyclop,funlen
func (c *Client) handleIntegrations(response http.ResponseWriter, request *http.Request) {
	integrations := Integrations{}
	integrations.Lidarr.Status = make([]*lidarr.SystemStatus, len(c.apps.Lidarr))
	integrations.Lidarr.Queue = make([]*lidarr.Queue, len(c.apps.Lidarr))
	integrations.Radarr.Status = make([]*radarr.SystemStatus, len(c.apps.Radarr))
	integrations.Radarr.Queue = make([]*radarr.Queue, len(c.apps.Radarr))
	integrations.Readarr.Status = make([]*readarr.SystemStatus, len(c.apps.Readarr))
	integrations.Readarr.Queue = make([]*readarr.Queue, len(c.apps.Readarr))
	integrations.Sonarr.Status = make([]*sonarr.SystemStatus, len(c.apps.Sonarr))
	integrations.Sonarr.Queue = make([]*sonarr.Queue, len(c.apps.Sonarr))
	integrations.Prowlarr.Status = make([]*prowlarr.SystemStatus, len(c.apps.Prowlarr))
	integrations.Lidarr.StatusAge = make([]time.Time, len(c.apps.Lidarr))
	integrations.Radarr.StatusAge = make([]time.Time, len(c.apps.Radarr))
	integrations.Readarr.StatusAge = make([]time.Time, len(c.apps.Readarr))
	integrations.Sonarr.StatusAge = make([]time.Time, len(c.apps.Sonarr))
	integrations.Prowlarr.StatusAge = make([]time.Time, len(c.apps.Prowlarr))
	integrations.Lidarr.QueueAge = make([]time.Time, len(c.apps.Lidarr))
	integrations.Radarr.QueueAge = make([]time.Time, len(c.apps.Radarr))
	integrations.Readarr.QueueAge = make([]time.Time, len(c.apps.Readarr))
	integrations.Sonarr.QueueAge = make([]time.Time, len(c.apps.Sonarr))

	if item := data.Get("snapshot"); item != nil {
		integrations.SnapshotAge = item.Time
		integrations.Snapshot, _ = item.Data.(*snapshot.Snapshot)
	}

	if ps := data.Get("plexStatus"); ps != nil {
		integrations.PlexAge = ps.Time
		integrations.Plex, _ = ps.Data.(*plex.PMSInfo)
	}

	if item := data.Get("plexCurrentSessions"); item != nil {
		integrations.SessionsAge = item.Time
		integrations.Sessions, _ = item.Data.(*plex.Sessions)
	}

	if item := data.GetWithID("tautulliStatus", 1); item != nil {
		integrations.TautulliAge = item.Time
		integrations.Tautulli, _ = item.Data.(*tautulli.Info)
	}

	if item := data.Get("tautulliUsers"); item != nil {
		integrations.TautulliUsersAge = item.Time
		integrations.TautulliUsers, _ = item.Data.(*tautulli.Users)
	}

	if item := data.Get("dashboard"); item != nil {
		integrations.DashboardAge = item.Time
		integrations.Dashboard, _ = item.Data.(*dashboard.States)
	}

	for idx := range c.apps.Lidarr {
		if item := data.GetWithID("lidarrStatus", idx); item != nil {
			integrations.Lidarr.StatusAge[idx] = item.Time
			integrations.Lidarr.Status[idx], _ = item.Data.(*lidarr.SystemStatus)
		}

		if item := data.GetWithID("lidarr", idx); item != nil {
			integrations.Lidarr.QueueAge[idx] = item.Time
			integrations.Lidarr.Queue[idx], _ = item.Data.(*lidarr.Queue)
		}
	}

	for idx := range c.apps.Radarr {
		if item := data.GetWithID("radarrStatus", idx); item != nil {
			integrations.Radarr.StatusAge[idx] = item.Time
			integrations.Radarr.Status[idx], _ = item.Data.(*radarr.SystemStatus)
		}

		if item := data.GetWithID("radarr", idx); item != nil {
			integrations.Radarr.QueueAge[idx] = item.Time
			integrations.Radarr.Queue[idx], _ = item.Data.(*radarr.Queue)
		}
	}

	for idx := range c.apps.Readarr {
		if item := data.GetWithID("readarrStatus", idx); item != nil {
			integrations.Readarr.StatusAge[idx] = item.Time
			integrations.Readarr.Status[idx], _ = item.Data.(*readarr.SystemStatus)
		}

		if item := data.GetWithID("readarr", idx); item != nil {
			integrations.Readarr.QueueAge[idx] = item.Time
			integrations.Readarr.Queue[idx], _ = item.Data.(*readarr.Queue)
		}
	}

	for idx := range c.apps.Sonarr {
		if item := data.GetWithID("sonarrStatus", idx); item != nil {
			integrations.Sonarr.StatusAge[idx] = item.Time
			integrations.Sonarr.Status[idx], _ = item.Data.(*sonarr.SystemStatus)
		}

		if item := data.GetWithID("sonarr", idx); item != nil {
			integrations.Sonarr.QueueAge[idx] = item.Time
			integrations.Sonarr.Queue[idx], _ = item.Data.(*sonarr.Queue)
		}
	}

	for idx := range c.apps.Prowlarr {
		if item := data.GetWithID("prowlarrStatus", idx); item != nil {
			integrations.Prowlarr.StatusAge[idx] = item.Time
			integrations.Prowlarr.Status[idx], _ = item.Data.(*prowlarr.SystemStatus)
		}
	}

	if err := json.NewEncoder(response).Encode(integrations); err != nil {
		logs.Log.Errorf("Encoding integrations: %v", err)
	}
}

// handleRunCommand only handles commands with arguments.
// Commands without arguments are handled as an instance test.
//
//	@Summary		Run command
//	@Description	Executes a specific command with provided arguments.
//	@Tags			Integrations
//	@Accept			application/x-www-form-urlencoded
//	@Produce		text/plain
//	@Param			hash	path		string		true	"Command hash"
//	@Param			args	formData	[]string	false	"Command arguments"	collectionFormat(multi)
//	@Success		200		{string}	string		"success message"
//	@Failure		400		{string}	string		"invalid command hash"
//	@Router			/runCommand/{hash} [post]
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
//
//	@Summary		Get process list
//	@Description	Returns a list of currently running system processes.
//	@Tags			System
//	@Produce		text/plain
//	@Success		200	{string}	string	"process list data"
//	@Failure		500	{string}	string	"error getting process list"
//	@Router			/ps [get]
func (c *Client) handleProcessList(response http.ResponseWriter, request *http.Request) {
	if ps, err := getProcessList(request.Context()); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	} else if _, err = ps.WriteTo(response); err != nil {
		user, _ := c.getUserName(request)
		logs.Log.Errorf("[gui '%s' requested] Writing HTTP Response: %v", user, err)
	}
}

// handleStartFileWatcher starts a file watcher.
//
//	@Summary		Start file watcher
//	@Description	Starts monitoring a specific file for changes.
//	@Tags			Integrations
//	@Produce		text/plain
//	@Param			index	path		int		true	"File watcher index"
//	@Success		200		{string}	string	"success message"
//	@Failure		400		{string}	string	"invalid or unknown index"
//	@Failure		406		{string}	string	"watcher already running"
//	@Failure		500		{string}	string	"error starting watcher"
//	@Router			/startFileWatch/{index} [get]
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

// handleStopFileWatcher stops a file watcher.
//
//	@Summary		Stop file watcher
//	@Description	Stops monitoring a specific file for changes.
//	@Tags			Integrations
//	@Produce		text/plain
//	@Param			index	path		int		true	"File watcher index"
//	@Success		200		{string}	string	"success message"
//	@Failure		400		{string}	string	"invalid or unknown index"
//	@Failure		406		{string}	string	"watcher already stopped"
//	@Failure		500		{string}	string	"error stopping watcher"
//	@Router			/stopFileWatch/{index} [get]
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

// handleConfigPost handles the reconfig endpoint.
//
//	@Summary		Update configuration
//	@Description	Updates the application configuration with new settings and optionally triggers a reload.
//	@Tags			System
//	@Accept			json
//	@Produce		text/plain
//	@Param			noreload	query		string				false	"set to 'true' to skip reload"
//	@Param			config		body		configfile.Config	true	"Configuration data"
//	@Success		200			{string}	string				"success message"
//	@Failure		400			{string}	string				"invalid request or config"
//	@Failure		500			{string}	string				"error saving config"
//	@Router			/reconfig [post]
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

// handleAPIKey allows updating the API key from the GUI without a password.
// This method is only enabled if the API Key is not 36 characters long.
func (c *Client) handleAPIKey(respond http.ResponseWriter, request *http.Request) {
	err := website.TestApiKey(request.Context(), request.Header.Get("X-Api-Key"))
	if err != nil {
		http.Error(respond, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Lock()
	defer c.Unlock()

	c.Config.APIKey = request.Header.Get("X-Api-Key")

	err = c.saveNewConfig(request.Context(), c.Config)
	if err != nil {
		http.Error(respond, "Failed Writing Config File: "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer c.triggerConfigReload(website.EventGUI, "GUI Requested")

	username, _ := c.getUserName(request)
	if username == "" {
		username = "user"
	}

	// respond.
	logs.Log.Printf("[gui '%s' requested] Updated Configuration. Reloading in 5 seconds...", username)
	http.Error(respond, "Config Saved. Reloading in 5 seconds...", http.StatusOK)

	respond.WriteHeader(http.StatusOK)
	_, _ = respond.Write([]byte("API key set successfully, reloading!"))
}
