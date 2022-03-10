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
	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/fsnotify/fsnotify"
	"github.com/hako/durafmt"
	"github.com/mitchellh/go-homedir"
	"github.com/shirou/gopsutil/v3/host"
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
		// returns the current time.
		"now": time.Now,
		// returns an integer divided by a million.
		"megabyte": megabyte,
		// returns the URL base.
		"base": func() string { return strings.TrimSuffix(c.Config.URLBase, "/") },
		// returns the files url base.
		"files": func() string { return path.Join(c.Config.URLBase, "files") },
		// adds 1 an integer, to deal with instance IDs for humans.
		"instance": func(idx int) int { return idx + 1 },
		// returns true if the environment variable has a value.
		"locked":   func(env string) bool { return os.Getenv(env) != "" },
		"contains": strings.Contains,
		"since":    since,
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

func megabyte(size interface{}) string {
	val := int64(0)

	switch valtype := size.(type) {
	case int64:
		val = valtype
	case uint64:
		val = int64(valtype)
	case int:
		val = int64(valtype)
	}

	switch {
	case val > mnd.Megabyte*mnd.Kilobyte*1000: // lul
		return fmt.Sprintf("%.2f Tb", float64(val)/float64(mnd.Megabyte*mnd.Megabyte))
	case val > mnd.Megabyte*1000:
		return fmt.Sprintf("%.2f Gb", float64(val)/float64(mnd.Megabyte*mnd.Kilobyte))
	case val > mnd.Kilobyte*1000:
		return fmt.Sprintf("%.1f Mb", float64(val)/float64(mnd.Megabyte))
	default:
		return fmt.Sprintf("%.1f Kb", float64(val)/float64(mnd.Kilobyte))
	}
}

func since(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}

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
	Config      *configfile.Config             `json:"config"`
	Flags       *configfile.Flags              `json:"flags"`
	Username    string                         `json:"username"`
	Dynamic     bool                           `json:"dynamic"`
	Webauth     bool                           `json:"webauth"`
	Msg         string                         `json:"msg,omitempty"`
	Version     map[string]interface{}         `json:"version"`
	LogFiles    *logs.LogFileInfos             `json:"logFileInfo"`
	ConfigFiles *logs.LogFileInfos             `json:"configFileInfo"`
	ClientInfo  *notifiarr.ClientInfo          `json:"clientInfo"`
	Expvar      exp.AllData                    `json:"expvar"`
	HostInfo    *host.InfoStat                 `json:"hostInfo"`
	Disks       map[string]*snapshot.Partition `json:"disks"`
}

func (c *Client) renderTemplate(response io.Writer, req *http.Request,
	templateName, msg string) {
	clientInfo, _ := c.website.GetClientInfo()
	if clientInfo == nil {
		clientInfo = &notifiarr.ClientInfo{}
	}

	binary, _ := os.Executable()
	userName, dynamic := c.getUserName(req)

	err := c.templat.ExecuteTemplate(response, templateName, &templateData{
		Config:      c.Config,
		Flags:       c.Flags,
		Username:    userName,
		Dynamic:     dynamic,
		Webauth:     c.webauth,
		Msg:         msg,
		LogFiles:    c.Logger.GetAllLogFilePaths(),
		ConfigFiles: logs.GetFilePaths(c.Flags.ConfigFile),
		ClientInfo:  clientInfo,
		Disks:       c.getDisks(),
		Version: map[string]interface{}{
			"started":   version.Started.Round(time.Second),
			"program":   c.Flags.Name(),
			"version":   version.Version,
			"revision":  version.Revision,
			"branch":    version.Branch,
			"buildUser": version.BuildUser,
			"buildDate": version.BuildDate,
			"goVersion": version.GoVersion,
			"os":        runtime.GOOS,
			"arch":      runtime.GOARCH,
			"binary":    binary,
			"environ":   environ(),
			"docker":    mnd.IsDocker,
		},
		Expvar:   exp.GetAllData(),
		HostInfo: c.website.HostInfoNoError(),
	})
	if err != nil {
		c.Errorf("Sending HTTP Response: %v", err)
	}
}

func environ() map[string]string {
	out := make(map[string]string)

	for _, v := range os.Environ() {
		if s := strings.SplitN(v, "=", 2); len(s) == 2 && s[0] != "" { //nolint:gomnd
			out[s[0]] = s[1]
		}
	}

	return out
}

func (c *Client) setUserPass(username, password string) error {
	c.Lock()
	defer c.Unlock()

	current := c.Config.UIPassword

	err := c.Config.UIPassword.Set(username + ":" + password)
	if err != nil {
		c.Config.UIPassword = current
		return fmt.Errorf("saving username and password: %w", err)
	}

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
			found++ // we have a line

			if found <= skip {
				// skip writing new lines until we get to our first line.
				continue
			}
		}

		if found >= skip {
			output.WriteByte(char[0])
		}

		if found >= count+skip || // we found enough lines.
			location >= fileSize-1 { // end of file.
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

func (c *Client) getDisks() map[string]*snapshot.Partition {
	output := make(map[string]*snapshot.Partition)
	snapcnfg := &snapshot.Config{
		Plugins:   &snapshot.Plugins{},
		DiskUsage: true,
		AllDrives: true,
		ZFSPools:  c.Config.Snapshot.ZFSPools,
		UseSudo:   c.Config.Snapshot.UseSudo,
		//		Raid:      c.Config.Snapshot.Raid,
	}
	snapshot, _, _ := snapcnfg.GetSnapshot()

	for k, v := range snapshot.DiskUsage {
		output[k] = v
	}

	for k, v := range snapshot.ZFSPool {
		output[k] = v
	}

	return output
}
