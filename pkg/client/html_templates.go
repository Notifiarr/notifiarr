package client

import (
	"bytes"
	"context"
	"encoding/json"
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

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/fsnotify/fsnotify"
	"github.com/hako/durafmt"
	"github.com/mitchellh/go-homedir"
	"github.com/shirou/gopsutil/v3/host"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golift.io/cnfg"
	"golift.io/version"
)

// loadAssetsTemplates watches for changs to template files, and loads them.
func (c *Client) loadAssetsTemplates(ctx context.Context) error {
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

	dirList, err := os.ReadDir(templates)
	if err != nil {
		return fmt.Errorf("cannot watch '%s' templates subfolders: %w", templates, err)
	}

	for _, item := range dirList {
		if !item.IsDir() {
			continue
		}

		// Watch each sub folder too.
		name := filepath.Join(templates, item.Name())
		if err := fsn.Add(name); err != nil {
			return fmt.Errorf("cannot watch '%s' templates subfolder: %w", name, err)
		}
	}

	go c.watchAssetsTemplates(ctx, fsn)

	return nil
}

func (c *Client) watchAssetsTemplates(ctx context.Context, fsn *fsnotify.Watcher) {
	defer c.CapturePanic()

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

			if err := c.StopWebServer(ctx); err != nil {
				panic("Stopping web server: " + err.Error())
			}

			if err := c.ParseGUITemplates(); err != nil {
				c.Errorf("fsnotify/parsing templates: %v", err)
			}

			c.StartWebServer()
		}
	}
}

func (c *Client) getFuncMap() template.FuncMap { //nolint:funlen
	title := cases.Title(language.AmericanEnglish)

	return template.FuncMap{
		"lower": strings.ToLower,
		"cutindex": func(str, delim, def string, idx int) string {
			split := strings.Split(str, delim)
			if idx >= len(split) {
				return def
			}
			return split[idx]
		},
		"cache":   data.Get,
		"cacheID": data.GetWithID,
		"tojson": func(input any) string {
			output, _ := json.MarshalIndent(input, "", " ")
			return string(output)
		},
		"dateFmt": func(date time.Time) string {
			if ci := clientinfo.Get(); ci != nil {
				return ci.User.DateFormat.Format(date)
			}

			return date.String()
		},
		"title":       title.String,
		"plexmedia":   plex.GetMediaTranscode,
		"todaysemoji": mnd.TodaysEmoji,
		"fortune":     Fortune,
		// returns the current time.
		"now": time.Now,
		// returns an integer divided by a million.
		"megabyte": megabyte,
		// returns the URL base.
		"base": func() string { return path.Join(c.Config.URLBase, "ui") + "/" },
		// returns the files url base.
		"files": func() string { return path.Join(c.Config.URLBase, "files") },
		// adds 1 an integer, to deal with instance IDs for humans.
		"instance": func(idx int) int { return idx + 1 },
		// returns true if the environment variable has a value.
		"locked":   func(env string) bool { return os.Getenv(env) != "" },
		"contains": strings.Contains,
		"since":    since,
		"percent": func(i, j float64) int64 {
			return int64(i / j * 100) //nolint:gomnd
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
		"one259": func() (num []float64) { // 1 to 59
			for i := float64(1); i < 60; i++ {
				num = append(num, i)
			}
			return num
		},
		"add": func(i, j float64) float64 {
			return i + j
		},
		"intervaloptions": intervaloptions,
	}
}

type option struct {
	Val string // value (machine)
	Op  string // option (human)
	Sel bool   // selected
}

func intervaloptions(current cnfg.Duration) []*option { //nolint:funlen
	times := []time.Duration{
		-1 * time.Second, // Disable Service Checks
		30 * time.Second,
		45 * time.Second,
		1 * time.Minute,
		1*time.Minute + 15*time.Second,
		1*time.Minute + 30*time.Second,
		1*time.Minute + 45*time.Second,
		2 * time.Minute,
		2*time.Minute + 15*time.Second,
		2*time.Minute + 30*time.Second,
		2*time.Minute + 45*time.Second,
		3 * time.Minute,
		3*time.Minute + 15*time.Second,
		3*time.Minute + 30*time.Second,
		3*time.Minute + 45*time.Second,
		4 * time.Minute,
		4*time.Minute + 15*time.Second,
		4*time.Minute + 30*time.Second,
		4*time.Minute + 45*time.Second,
		5 * time.Minute,
		5*time.Minute + 15*time.Second,
		5*time.Minute + 30*time.Second,
		5*time.Minute + 45*time.Second,
		6 * time.Minute,
		6*time.Minute + 30*time.Second,
		7 * time.Minute,
		7*time.Minute + 30*time.Second,
		8 * time.Minute,
		8*time.Minute + 30*time.Second,
		9 * time.Minute,
		9*time.Minute + 30*time.Second,
		10 * time.Minute,
		11 * time.Minute,
		12 * time.Minute,
		13 * time.Minute,
		14 * time.Minute,
		15 * time.Minute,
		20 * time.Minute,
		30 * time.Minute,
		35 * time.Minute,
		40 * time.Minute,
		45 * time.Minute,
		50 * time.Minute,
		1 * time.Hour,
		1*time.Hour + 30*time.Minute,
		2 * time.Hour,
	}
	output := []*option{}

	if current.Duration == 0 {
		output = append(output, &option{
			Val: current.String(),
			Op:  "select...",
			Sel: true,
		})
	}

	for idx, dur := range times {
		if idx != 0 && current.Duration < dur && current.Duration > times[idx-1] && current.Duration != 0 {
			// This adds the current selected value in case it does not match one of the predefined options.
			output = append(output, &option{
				Val: current.String(),
				Op:  durShort(current.Duration),
				Sel: true,
			})
		}

		if dur < 0 {
			// We should only have 1 less than 0.
			output = append(output, &option{
				Val: cnfg.Duration{Duration: dur}.String(),
				Op:  "Disabled",
				Sel: current.Duration < 0,
			})
		} else {
			output = append(output, &option{
				Val: cnfg.Duration{Duration: dur}.String(),
				Op:  durShort(dur),
				Sel: current.Duration == dur,
			})
		}
	}

	return output
}

// durShort is gross but gets the job done.
func durShort(dur time.Duration) string {
	output := cnfg.Duration{Duration: dur}.String()
	output = strings.ReplaceAll(output, "m", " min")
	output = strings.ReplaceAll(output, "s", " sec")
	output = strings.ReplaceAll(output, "h", " hour")

	s := ""
	if dur.Hours() != 1 {
		s = "s"
	}

	if dur.Minutes() != 0 {
		output = strings.ReplaceAll(output, "hour", "hour"+s+" ")
	}

	s = ""
	if dur.Minutes() != 1 {
		s = "s"
	}

	if dur.Seconds() > 60 && int(dur.Seconds())%60 != 0 || int(dur.Hours()) > 0 {
		output = strings.ReplaceAll(output, "min", "min ")
	} else {
		output = strings.ReplaceAll(output, "min", "minute"+s)
	}

	if dur.Minutes() < 1 {
		output = strings.ReplaceAll(output, "sec", " seconds")
	}

	return output
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
	case val > mnd.Megabyte*mnd.Megabyte*mnd.Kilobyte*1000: // 2^60
		return fmt.Sprintf("%.2f EiB", float64(val)/float64(mnd.Megabyte*mnd.Megabyte*mnd.Megabyte))
	case val > mnd.Megabyte*mnd.Megabyte*1000: // 2^50
		return fmt.Sprintf("%.2f PiB", float64(val)/float64(mnd.Megabyte*mnd.Megabyte*mnd.Kilobyte))
	case val > mnd.Megabyte*mnd.Kilobyte*1000: // 2^40
		return fmt.Sprintf("%.2f TiB", float64(val)/float64(mnd.Megabyte*mnd.Megabyte))
	case val > mnd.Megabyte*1000: // 2^30
		return fmt.Sprintf("%.2f GiB", float64(val)/float64(mnd.Megabyte*mnd.Kilobyte))
	case val > mnd.Kilobyte*1000: // 2^20
		return fmt.Sprintf("%.1f MiB", float64(val)/float64(mnd.Megabyte))
	default: // 2^10
		return fmt.Sprintf("%.1f KiB", float64(val)/float64(mnd.Kilobyte))
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
func (c *Client) ParseGUITemplates() error {
	// Index and 404 do not have template files, but they can be customized.
	index := "<p>" + c.Flags.Name() + `: <strong>working</strong></p>`
	c.template = template.Must(template.New("index.html").Parse(index)).Funcs(c.getFuncMap())

	var err error

	// Parse all our compiled-in templates.
	for _, name := range bindata.AssetNames() {
		if strings.HasPrefix(name, "templates/") {
			trim := strings.TrimPrefix(name, "templates/")
			if c.template, err = c.template.New(trim).Parse(bindata.MustAssetString(name)); err != nil {
				return fmt.Errorf("bug parsing internal template: %w", err)
			}
		}
	}

	if c.Flags.Assets != "" {
		return c.parseCustomTemplates()
	}

	return nil
}

func (c *Client) parseCustomTemplates() error {
	templatePath := filepath.Join(c.Flags.Assets, "templates")

	c.Printf("==> Parsing and watching HTML templates @ %s", templatePath)

	return filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error { //nolint:wrapcheck
		if err != nil {
			return fmt.Errorf("walking custom template path: %w", err)
		}

		if info.IsDir() {
			return nil // cannot parse directories.
		}

		// Convert windows paths to unix paths for template names.
		trim := strings.TrimPrefix(strings.ReplaceAll(strings.TrimPrefix(path, templatePath), `\`, "/"), "/")
		c.Debugf("Parsing Template File '%s' to %s", path, trim)

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading custom template: %w", err)
		}

		c.template, err = c.template.New(trim).Parse(string(data))
		if err != nil {
			return fmt.Errorf("parsing custom template: %w", err)
		}

		return nil
	})
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
	ClientInfo  *clientinfo.ClientInfo         `json:"clientInfo"`
	Expvar      mnd.AllData                    `json:"expvar"`
	HostInfo    *host.InfoStat                 `json:"hostInfo"`
	Disks       map[string]*snapshot.Partition `json:"disks"`
	Headers     map[string][]string            `json:"headers"`
	ProxyAllow  bool                           `json:"proxyAllow"`
	UpstreamIP  string                         `json:"upstreamIp"`
}

func (c *Client) renderTemplate(
	ctx context.Context,
	response io.Writer,
	req *http.Request,
	templateName,
	msg string,
) {
	clientInfo := clientinfo.Get()
	if clientInfo == nil {
		clientInfo = &clientinfo.ClientInfo{}
	}

	binary, _ := os.Executable()
	userName, dynamic := c.getUserName(req)
	hostInfo, _ := c.website.GetHostInfo(ctx)
	backupPath := filepath.Join(filepath.Dir(c.Flags.ConfigFile), "backups", filepath.Base(c.Flags.ConfigFile))

	err := c.template.ExecuteTemplate(response, templateName, &templateData{
		ProxyAllow:  c.Config.Allow.Contains(req.RemoteAddr),
		UpstreamIP:  strings.Trim(req.RemoteAddr[:strings.LastIndex(req.RemoteAddr, ":")], "[]"),
		Config:      c.Config,
		Flags:       c.Flags,
		Username:    userName,
		Dynamic:     dynamic,
		Webauth:     c.webauth,
		Msg:         msg,
		LogFiles:    c.Logger.GetAllLogFilePaths(),
		ConfigFiles: logs.GetFilePaths(c.Flags.ConfigFile, backupPath),
		ClientInfo:  clientInfo,
		Disks:       c.getDisks(ctx),
		Headers:     req.Header,
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
			"uid":       os.Getuid(),
			"gid":       os.Getgid(),
		},
		Expvar:   mnd.GetAllData(),
		HostInfo: hostInfo,
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

func (c *Client) setUserPass(ctx context.Context, authType, username, password string) error {
	c.Lock()
	defer c.Unlock()

	current := c.Config.UIPassword

	var err error

	switch authType {
	default:
		err = c.Config.UIPassword.Set(username + ":" + password)
	case "header":
		err = c.Config.UIPassword.SetHeader(username)
	case "nopass":
		err = c.Config.UIPassword.SetNoAuth(username)
	}

	if err != nil {
		c.Config.UIPassword = current
		return fmt.Errorf("saving new auth settings: %w", err)
	}

	if err := c.saveNewConfig(ctx, c.Config); err != nil {
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
	output.Grow(count * 150) //nolint:gomnd

	for {
		location-- // read 1 byte
		if _, err := fileHandle.Seek(location, io.SeekEnd); err != nil {
			return nil, fmt.Errorf("seeking open file: %w", err)
		}

		if _, err := fileHandle.Read(char); err != nil {
			return nil, fmt.Errorf("reading open file: %w", err)
		}

		if location != -1 && (char[0] == 10) { //nolint:gomnd
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
	output.Grow(count * 150) //nolint:gomnd

	for ; ; location++ {
		if _, err := fileHandle.Seek(location, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seeking open file: %w", err)
		}

		if _, err := fileHandle.Read(char); err != nil {
			return nil, fmt.Errorf("reading open file: %w", err)
		}

		if char[0] == 10 { //nolint:gomnd
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

func (c *Client) getDisks(ctx context.Context) map[string]*snapshot.Partition {
	output := make(map[string]*snapshot.Partition)
	snapcnfg := &snapshot.Config{
		Plugins:   &snapshot.Plugins{},
		DiskUsage: true,
		AllDrives: true,
		ZFSPools:  c.Config.Snapshot.ZFSPools,
		UseSudo:   c.Config.Snapshot.UseSudo,
		//		Raid:      c.Config.Snapshot.Raid,
	}
	snapshot, _, _ := snapcnfg.GetSnapshot(ctx)

	for k, v := range snapshot.DiskUsage {
		output[k] = v
	}

	for k, v := range snapshot.Quotas {
		output["Quota: "+k] = v
	}

	for k, v := range snapshot.ZFSPool {
		output[k] = v
	}

	return output
}
