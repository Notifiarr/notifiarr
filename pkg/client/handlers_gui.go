package client

import (
	"archive/zip"
	"context"
	"encoding/json"
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
	"github.com/gorilla/mux"
	"golift.io/version"
)

type templateData struct {
	Config   *configfile.Config `json:"config"`
	Flags    *configfile.Flags  `json:"flags"`
	Username string             `json:"username"`
	Data     url.Values         `json:"data,omitempty"`
	Msg      string             `json:"msg,omitempty"`
	Version  map[string]string  `json:"version"`
	LogFiles *logs.LogFileInfos `json:"logFileInfo"`
}

// userNameValue is used a context value key.
type userNameValue string

const userNameStr userNameValue = "username"

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
	validUsername, validPassword := "admin", c.Config.UIPassword
	if spl := strings.SplitN(validPassword, ":", 2); len(spl) == 2 { //nolint:gomnd
		validUsername = spl[0]
		validPassword = spl[1]
	}

	switch providedUsername := request.FormValue("name"); {
	case len(validPassword) < 16: // nolint:gomnd
		c.loginPage(response, request, "Invalid Password Configured")
	case c.getUserName(request) != "":
		http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
	case request.Method == http.MethodGet:
		c.loginPage(response, request, "")
	case providedUsername == validUsername && validPassword == request.FormValue("password"):
		c.setSession(providedUsername, response)
		http.Redirect(response, request, c.Config.URLBase, http.StatusFound)
	default: // Start over.
		c.loginPage(response, request, "Invalid Password")
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

func (c *Client) getLogDeleteHandler(response http.ResponseWriter, req *http.Request) {
	logID := mux.Vars(req)["id"]
	logs := c.Logger.GetAllLogFilePaths()

	for _, logFile := range logs.Logs {
		if logFile.ID != logID {
			continue
		}

		if err := os.Remove(logFile.Path); err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}

		if _, err := response.Write([]byte("ok")); err != nil {
			c.Errorf("Writing HTTP Response: %v", err)
		}

		break
	}

	// return new list of files here.
}

func (c *Client) getLogDownloadHandler(response http.ResponseWriter, req *http.Request) {
	logID := mux.Vars(req)["id"]
	logs := c.Logger.GetAllLogFilePaths()

	for _, logFile := range logs.Logs {
		if logFile.ID != logID {
			continue
		}

		zipWriter := zip.NewWriter(response)
		defer zipWriter.Close()

		logOpen, err := os.Open(logFile.Path)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
		defer logOpen.Close()

		newZippedFile, err := zipWriter.Create(logFile.Name)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}

		response.Header().Set("Content-Disposition", "attachment; filename="+logFile.Name+".zip")
		response.Header().Set("Content-Type", "application/zip")

		if _, err := io.Copy(newZippedFile, logOpen); err != nil {
			c.Errorf("Sending Zipped Log File: %v", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (c *Client) getLogHandler(response http.ResponseWriter, req *http.Request) {
	logID := mux.Vars(req)["id"]
	logs := c.Logger.GetAllLogFilePaths()
	skip, _ := strconv.Atoi(mux.Vars(req)["skip"])

	count, _ := strconv.Atoi(mux.Vars(req)["lines"])
	if count == 0 {
		count = 500
	}

	for _, logFile := range logs.Logs {
		if logFile.ID != logID {
			continue
		}

		lines, err := getLastLinesInFile(logFile.Path, count, skip)
		if err != nil {
			c.Errorf("Handling Log File Request: %v", err)
			http.Error(response, err.Error(), http.StatusInternalServerError)
		} else if logFile.Size == 0 {
			http.Error(response, "the file is empty", http.StatusInternalServerError)
		} else if _, err = response.Write(lines); err != nil {
			c.Errorf("Writing HTTP Response: %v", err)
		}

		return
	}
}

// getSettingsHandler returns all settings in a json blob. Useful for ajax requests.
func (c *Client) getSettingsHandler(response http.ResponseWriter, req *http.Request) {
	var err error

	response.Header().Set("content-type", "application/json")

	switch config := mux.Vars(req)["config"]; config {
	default:
		item := getFieldName(config, *c.Config)
		if item == nil {
			http.Error(response, `{"error": "no config item: `+config+`"}`, http.StatusBadRequest)
			return
		}

		err = json.NewEncoder(response).Encode(map[string]interface{}{config: item})
	case "flags":
		err = json.NewEncoder(response).Encode(map[string]interface{}{config: c.Flags})
	case "config":
		err = json.NewEncoder(response).Encode(map[string]interface{}{config: c.Config})
	case "username":
		err = json.NewEncoder(response).Encode(map[string]string{config: c.getUserName(req)})
	case "version":
		err = json.NewEncoder(response).Encode(map[string]string{
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
		})
	case "all":
		err = json.NewEncoder(response).Encode(&templateData{
			Config:   c.Config,
			Flags:    c.Flags,
			Username: c.getUserName(req),
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
	}

	if err != nil {
		c.Errorf("Sending HTTP JSON Response: %v", err)
		http.Error(response, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
	}
}

// getFieldName allows pulling a config item by json tag name.
func getFieldName(key string, config interface{}) interface{} {
	sType := reflect.TypeOf(config)
	sVal := reflect.ValueOf(config)

	if sType.Kind() == reflect.Ptr {
		sType = reflect.TypeOf(config).Elem()
		sVal = reflect.ValueOf(config).Elem()
	}

	if sType.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < sType.NumField(); i++ { //nolint:varnamelen
		loopType := reflect.TypeOf(sType.Field(i))
		//  Loop into exported anonymous structs.
		if loopType.Kind() == reflect.Struct && sType.Field(i).Anonymous && sType.Field(i).IsExported() {
			if item := getFieldName(key, sVal.Field(i).Interface()); item != nil {
				return item
			}
		}

		// See if this item has a json tag equal to our requested key.
		v := strings.Split(sType.Field(i).Tag.Get("json"), ",")[0]
		if v == key {
			return sVal.Field(i).Interface()
		}
	}

	return nil
}

func (c *Client) renderHTTPtemplate(w io.Writer, req *http.Request, tmpl string, msg string) {
	err := c.templat.ExecuteTemplate(w, tmpl, &templateData{
		Config:   c.Config,
		Flags:    c.Flags,
		Username: c.getUserName(req),
		Data:     req.PostForm,
		Msg:      msg,
		LogFiles: c.Logger.GetAllLogFilePaths(),
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

func (c *Client) loginPage(response http.ResponseWriter, request *http.Request, msg string) {
	response.Header().Add("content-type", "text/html")

	if request.Method != http.MethodGet {
		response.WriteHeader(http.StatusUnauthorized)
	}

	c.renderHTTPtemplate(response, request, "index.html", msg)
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
