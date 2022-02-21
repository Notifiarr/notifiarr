package client

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/fsnotify/fsnotify"
	"github.com/hako/durafmt"
	"github.com/mitchellh/go-homedir"
	"golift.io/version"
)

// loadAssetsTemplates watches for changs to template files, and loads them.
func (c *Client) loadAssetsTemplates() error {
	if err := c.ParseGUITemplates(); err != nil {
		return err
	}

	if c.Flags.Assets == "" {
		return nil
	}

	fsn, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("fsnotify.NewWatcher: %w", err)
	}

	templates := filepath.Join(c.Flags.Assets, "templates")
	if err := fsn.Add(templates); err != nil {
		return fmt.Errorf("cannot watch '%s' templates path: %w", templates, err)
	}

	go c.watchAssetsTemplates(fsn)

	return nil
}

func (c *Client) watchAssetsTemplates(fsn *fsnotify.Watcher) {
	for {
		select {
		case err := <-fsn.Errors:
			c.Errorf("fsnotify: %v", err)
		case event, ok := <-fsn.Events:
			if !ok {
				return
			}

			if (event.Op&fsnotify.Write != fsnotify.Write && event.Op&fsnotify.Create != fsnotify.Create) ||
				!strings.HasSuffix(event.Name, ".html") {
				continue
			}

			c.Debugf("Got event: %s on %s, reloading HTML templates!", event.Op, event.Name)

			if err := c.StopWebServer(); err != nil {
				panic("Stopping web server: " + err.Error())
			}

			if err := c.ParseGUITemplates(); err != nil {
				c.Errorf("fsnotify/parsing templates: %v", err)
			}

			c.StartWebServer()
		}
	}
}

func (c *Client) getFuncMap() template.FuncMap {
	return template.FuncMap{
		// returns the username the logged in.
		"username": func() string { u, _ := c.getUserPass(); return u },
		// returns the current time.
		"now": time.Now,
		// returns an integer divided by a million.
		"megabyte": func(size int64) string { return fmt.Sprintf("%.2f", float64(size)/float64(mnd.Megabyte)) },
		// returns the URL base.
		"base": func() string { return strings.TrimSuffix(c.Config.URLBase, "/") },
		// returns the files url base.
		"files": func() string { return path.Join(c.Config.URLBase, "files") },
		// adds 1 an integer, to deal with instance IDs for humans.
		"instance": func(idx int) int { return idx + 1 },
		// returns true if the environment variable has a value.
		"locked":   func(env string) bool { return os.Getenv(env) != "" },
		"contains": strings.Contains,
		"since": func(t time.Time) string {
			return strings.ReplaceAll(durafmt.Parse(time.Since(t).Round(time.Second)).
				LimitFirstN(3). //nolint:gomnd
				Format(durafmt.Units{
					Year:   durafmt.Unit{Singular: "y", Plural: "y"},
					Week:   durafmt.Unit{Singular: "w", Plural: "w"},
					Day:    durafmt.Unit{Singular: "d", Plural: "d"},
					Hour:   durafmt.Unit{Singular: "h", Plural: "h"},
					Minute: durafmt.Unit{Singular: "m", Plural: "m"},
					Second: durafmt.Unit{Singular: "s", Plural: "s"},
				}), " ", "")
		},
		"min": func(s string) string {
			for _, pieces := range strings.Split(s, ",") {
				if split := strings.Split(pieces, ":"); len(split) >= 2 && split[0] == "count" {
					return split[1]
				}
			}
			return "0"
		},
		"max": func(s string) string {
			for _, pieces := range strings.Split(s, ",") {
				if split := strings.Split(pieces, ":"); len(split) > 2 && split[0] == "count" {
					return split[2]
				}
			}
			return "0"
		},
	}
}

// ParseGUITemplates parses the baked-in templates, and overrides them if a template directory is provided.
func (c *Client) ParseGUITemplates() (err error) {
	// Index and 404 do not have template files, but they can be customized.
	index := "<p>" + c.Flags.Name() + `: <strong>working</strong></p>`
	c.templat = template.Must(template.New("index.html").Parse(index)).Funcs(c.getFuncMap())

	// Parse all our compiled-in templates.
	for _, name := range bindata.AssetNames() {
		if strings.HasPrefix(name, "templates/") {
			c.templat = template.Must(c.templat.New(path.Base(name)).Parse(bindata.MustAssetString(name)))
		}
	}

	if c.Flags.Assets == "" {
		return nil
	}

	templates := filepath.Join(c.Flags.Assets, "templates", "*.html")
	c.Printf("==> Parsing and watching HTML templates @ %s", templates)

	c.templat, err = c.templat.ParseGlob(templates)
	if err != nil {
		return fmt.Errorf("parsing custom template: %w", err)
	}

	return nil
}

type templateData struct {
	Config      *configfile.Config    `json:"config"`
	Flags       *configfile.Flags     `json:"flags"`
	Username    string                `json:"username"`
	Msg         string                `json:"msg,omitempty"`
	Version     map[string]string     `json:"version"`
	LogFiles    *logs.LogFileInfos    `json:"logFileInfo"`
	ConfigFiles *logs.LogFileInfos    `json:"configFileInfo"`
	ClientInfo  *notifiarr.ClientInfo `json:"clientInfo"`
}

func (c *Client) renderTemplate(response io.Writer, req *http.Request,
	templateName, msg string) {
	clientInfo, _ := c.website.GetClientInfo(notifiarr.EventUser)
	if clientInfo == nil {
		clientInfo = &notifiarr.ClientInfo{}
	}

	err := c.templat.ExecuteTemplate(response, templateName, &templateData{
		Config:      c.Config,
		Flags:       c.Flags,
		Username:    c.getUserName(req),
		Msg:         msg,
		LogFiles:    c.Logger.GetAllLogFilePaths(),
		ConfigFiles: logs.GetFilePaths(c.Flags.ConfigFile),
		ClientInfo:  clientInfo,
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

// haveCustomFile searches known locatinos for a file. Returns the file's path.
func (c *Client) haveCustomFile(fileName string) string {
	cwd, _ := os.Getwd()
	exe, _ := os.Executable()

	paths := map[string][]string{
		mnd.Windows: {
			`~/notifiarr`,
			cwd,
			filepath.Dir(exe),
			`C:\ProgramData\notifiarr`,
		},
		"darwin": {
			"~/.notifiarr",
			"/usr/local/etc/notifiarr",
		},
		"default": {
			`~/notifiarr`,
			"/config",
			"/etc/notifiarr",
		},
	}

	findIn := paths[runtime.GOOS]
	if len(findIn) == 0 {
		findIn = paths["default"]
	}

	for _, find := range findIn {
		if find == "" {
			continue
		}

		custom, err := homedir.Expand(filepath.Join(find, fileName))
		if err != nil {
			custom = filepath.Join(find, fileName)
		}

		custom2, err := filepath.Abs(custom)
		if err == nil {
			custom = custom2
		}

		if _, err = os.Stat(custom); err == nil {
			return custom
		}
	}

	return ""
}

// getLinesFromFile makes it easy to tail or head a file. Sorta.
func getLinesFromFile(filepath, sort string, count, skip int) ([]byte, error) {
	fileHandle, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer fileHandle.Close()

	stat, err := fileHandle.Stat()
	if err != nil {
		return nil, fmt.Errorf("stating open file: %w", err)
	}

	switch sort {
	default:
		fallthrough
	case "tail", "tails":
		return readFileTail(fileHandle, stat.Size(), count, skip)
	case "head", "heads":
		return readFileHead(fileHandle, stat.Size(), count, skip)
	}
}

func readFileTail(fileHandle *os.File, fileSize int64, count, skip int) ([]byte, error) { //nolint:cyclop
	var (
		output   bytes.Buffer
		location int64
		filesize = fileSize
		char     = make([]byte, 1)
		found    int
	)

	// This is a magic number.
	// We assume 150 characters per line to optimize the buffer.
	output.Grow(count * 150) // nolint:gomnd

	for {
		location-- // read 1 byte
		if _, err := fileHandle.Seek(location, io.SeekEnd); err != nil {
			return nil, fmt.Errorf("seeking open file: %w", err)
		}

		if _, err := fileHandle.Read(char); err != nil {
			return nil, fmt.Errorf("reading open file: %w", err)
		}

		if location != -1 && (char[0] == 10) { // nolint:gomnd
			found++ // we found a line
		}

		if skip == 0 || found >= skip {
			output.WriteByte(char[0])
		}

		if found >= count+skip || // we found enough lines.
			location == -filesize { // beginning of file.
			out := revBytes(output)
			if len(out) > 0 && out[0] == '\n' {
				return out[1:], nil // strip off the /n
			}

			return out, nil
		}
	}
}

func readFileHead(fileHandle *os.File, fileSize int64, count, skip int) ([]byte, error) {
	var (
		output   bytes.Buffer
		location int64
		char     = make([]byte, 1)
		found    int
	)

	// This is a magic number.
	// We assume 150 characters per line to optimize the buffer.
	output.Grow(count * 150) // nolint:gomnd

	for ; ; location++ {
		if _, err := fileHandle.Seek(location, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking open file: %w", err)
		}

		if _, err := fileHandle.Read(char); err != nil {
			return nil, fmt.Errorf("reading open file: %w", err)
		}

		if char[0] == 10 { // nolint:gomnd
			found++ // we found a line
		}

		if skip == 0 || found > skip {
			output.WriteByte(char[0])
		}

		if found >= count+skip || // we found enough lines.
			location >= fileSize-1 { // beginning of file.
			return output.Bytes(), nil
		}
	}
}

// revBytes returns a bytes buffer reversed.
func revBytes(output bytes.Buffer) []byte {
	data := output.Bytes()
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	return data
}
