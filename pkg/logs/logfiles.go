package logs

import (
	"encoding/base64"
	"expvar"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	homedir "github.com/mitchellh/go-homedir"
	"golift.io/datacounter"
	"golift.io/rotatorr"
	"golift.io/rotatorr/timerotator"
)

//nolint:gochecknoglobals
var (
	stdout  = logCounter(os.Stdout.Name(), os.Stdout)
	discard = logCounter("/dev/null", io.Discard)
)

// setDefaultLogPaths makes sure a GUI app has log files defined.
// These are enforced on GUI OSes (like macOS.app and Windows).
func (l *Logger) setDefaultLogPaths() {
	// Make sure log file paths exist if AppName is provided; this indicates GUI OS.
	if l.LogConfig.AppName != "" { // This only happens if ui.HasGUI() is true.
		base := ".notifiarr" // windows and macos
		if !mnd.IsWindows && !mnd.IsDarwin {
			base = ".config/notifiarr" // *nix desktop
		}

		if l.LogConfig.LogFile == "" {
			l.LogConfig.LogFile = filepath.Join("~", base, l.LogConfig.AppName+defExt)
		}

		if l.LogConfig.HTTPLog == "" {
			l.LogConfig.HTTPLog = filepath.Join("~", base, l.LogConfig.AppName+httpExt)
		}
	}
}

// setAppLogPath sets the log path for app log.
func (l *Logger) setAppLogPath() {
	// Regular log file.
	if l.LogConfig.LogFile == "" {
		return
	}

	if f, err := homedir.Expand(l.LogConfig.LogFile); err == nil {
		l.LogConfig.LogFile = f
	} else if l.LogConfig.AppName != "" {
		l.LogConfig.LogFile = l.LogConfig.AppName + defExt
	}

	if f, err := filepath.Abs(l.LogConfig.LogFile); err == nil {
		l.LogConfig.LogFile = f
	}

	// If a directory was provided, append a file name.
	if stat, _ := os.Stat(l.LogConfig.LogFile); stat != nil && stat.IsDir() {
		l.LogConfig.LogFile = filepath.Join(l.LogConfig.LogFile, mnd.Title+defExt)
	}
}

// setHTTPLogPath sets the log path for HTTP log.
func (l *Logger) setHTTPLogPath() {
	if l.LogConfig.HTTPLog == "" {
		return
	}

	if f, err := homedir.Expand(l.LogConfig.HTTPLog); err == nil {
		l.LogConfig.HTTPLog = f
	} else if l.LogConfig.AppName != "" {
		l.LogConfig.HTTPLog = l.LogConfig.AppName + httpExt
	}

	if f, err := filepath.Abs(l.LogConfig.HTTPLog); err == nil {
		l.LogConfig.HTTPLog = f
	}

	// If a directory was provided, append a file name.
	if stat, _ := os.Stat(l.LogConfig.HTTPLog); stat != nil && stat.IsDir() {
		l.LogConfig.HTTPLog = filepath.Join(l.LogConfig.HTTPLog, mnd.Title+httpExt)
	}
}

func (l *Logger) openLogFile() {
	rotate := &rotatorr.Config{
		Filepath: l.LogConfig.LogFile,                         // log file name.
		FileSize: int64(l.LogConfig.LogFileMb) * mnd.Megabyte, // mnd.Megabytes
		FileMode: l.LogConfig.FileMode.Mode(),                 // set file mode.
		Rotatorr: &timerotator.Layout{
			FileCount:  l.LogConfig.LogFiles, // number of files to keep.
			PostRotate: l.postLogRotate,      // method to run after rotating.
		},
	}

	switch { // only use MultiWriter if we have > 1 writer.
	case !l.LogConfig.Quiet && l.LogConfig.LogFile != "":
		l.app = rotatorr.NewMust(rotate)
		l.InfoLog.SetOutput(io.MultiWriter(logCounter(l.LogConfig.LogFile, l.app), stdout))
	case !l.LogConfig.Quiet && l.LogConfig.LogFile == "":
		l.InfoLog.SetOutput(stdout)
	case l.LogConfig.LogFile == "":
		l.InfoLog.SetOutput(discard) // default is "nothing"
	default:
		l.app = rotatorr.NewMust(rotate)
		l.InfoLog.SetOutput(logCounter(l.LogConfig.LogFile, l.app))
	}

	// Don't forget errors log, and do standard logger too.
	if l.LogConfig.Debug && l.LogConfig.DebugLog == "" {
		l.DebugLog.SetOutput(l.InfoLog.Writer())
	}

	l.ErrorLog.SetOutput(l.InfoLog.Writer())
	log.SetOutput(l.InfoLog.Writer())
	l.postLogRotate("", "")
}

// This is only for the main log. To deal with stderr.
func (l *Logger) postLogRotate(fileName, newFile string) {
	l.postRotateCounter(fileName, newFile)

	if l.app != nil && l.app.File != nil {
		redirectStderr(l.app.File) // Log panics.
	}
}

func (l *Logger) openDebugLog() {
	if !l.LogConfig.Debug {
		// in case we're reloading without debug and had it before.
		l.DebugLog.SetOutput(io.Discard)
	}

	if !l.LogConfig.Debug || l.LogConfig.DebugLog == "" {
		return
	}

	if f, err := homedir.Expand(l.LogConfig.DebugLog); err == nil {
		l.LogConfig.DebugLog = f
	}

	if f, err := filepath.Abs(l.LogConfig.DebugLog); err == nil {
		l.LogConfig.DebugLog = f
	}

	if stat, err := os.Stat(l.LogConfig.DebugLog); err == nil {
		if stat.IsDir() {
			l.LogConfig.DebugLog = filepath.Join(l.LogConfig.DebugLog, mnd.Title+".debug"+defExt)
		}
	}

	rotateDebug := &rotatorr.Config{
		Filepath: l.LogConfig.DebugLog,                        // log file name.
		FileSize: int64(l.LogConfig.LogFileMb) * mnd.Megabyte, // mnd.Megabytes
		FileMode: l.LogConfig.FileMode.Mode(),                 // set file mode.
		Rotatorr: &timerotator.Layout{
			FileCount:  l.LogConfig.LogFiles, // number of files to keep.
			PostRotate: l.postRotateCounter,
		},
	}
	l.debug = rotatorr.NewMust(rotateDebug)

	if l.LogConfig.Quiet {
		l.DebugLog.SetOutput(logCounter(l.LogConfig.DebugLog, l.debug))
	} else {
		l.DebugLog.SetOutput(io.MultiWriter(logCounter(l.LogConfig.DebugLog, l.debug), stdout))
	}
}

func (l *Logger) openHTTPLog() {
	rotateHTTP := &rotatorr.Config{
		Filepath: l.LogConfig.HTTPLog,                         // log file name.
		FileSize: int64(l.LogConfig.LogFileMb) * mnd.Megabyte, // mnd.Megabytes
		FileMode: l.LogConfig.FileMode.Mode(),                 // set file mode.
		Rotatorr: &timerotator.Layout{
			FileCount:  l.LogConfig.LogFiles, // number of files to keep.
			PostRotate: l.postRotateCounter,
		},
	}

	switch { // only use MultiWriter if we have > 1 writer.
	case !l.LogConfig.Quiet && l.LogConfig.HTTPLog != "":
		l.web = rotatorr.NewMust(rotateHTTP)
		l.HTTPLog.SetOutput(io.MultiWriter(logCounter(l.LogConfig.HTTPLog, l.web), stdout))
	case !l.LogConfig.Quiet && l.LogConfig.HTTPLog == "":
		l.HTTPLog.SetOutput(stdout)
	case l.LogConfig.HTTPLog == "":
		l.HTTPLog.SetOutput(discard) // default is "nothing"
	default:
		l.web = rotatorr.NewMust(rotateHTTP)
		l.HTTPLog.SetOutput(logCounter(l.LogConfig.HTTPLog, l.web))
	}
}

// LogFileInfos holds metadata about files.
type LogFileInfos struct {
	Dirs []string
	Size int64
	List []*LogFileInfo
}

// LogFileInfo is returned by GetAllLogFilePaths.
type LogFileInfo struct {
	ID   string
	Name string
	Path string
	Size int64
	Time time.Time
	Mode fs.FileMode
	Used bool
	User string
}

// GetActiveLogFilePaths returns the configured log file paths.
func (l *LogConfig) GetActiveLogFilePaths() []string {
	logFiles := []string{
		l.LogFile,
		l.HTTPLog,
		l.DebugLog,
	}

	for cust := range customLog {
		if customLog[cust] != nil && customLog[cust].File != nil {
			if name := customLog[cust].File.Name(); name != "" {
				logFiles = append(logFiles, name)
			}
		}
	}

	return logFiles
}

// GetAllLogFilePaths searches the disk for log file names.
func (l *Logger) GetAllLogFilePaths() *LogFileInfos {
	return GetFilePaths(l.LogConfig.GetActiveLogFilePaths()...)
}

// GetFilePaths is a helper function that returns data about similar files in
// a folder with the provided file(s). This is useful to find "all the log files"
// or "all the .conf files" in a folder. Simply pass in 1 or more file paths, and
// any files in the same folder with the same extension will be returned.
func GetFilePaths(files ...string) *LogFileInfos { //nolint:cyclop
	contain := make(map[string]struct{})
	dirs := make(map[string]struct{})

	for _, logFilePath := range files {
		dirExpanded, err := homedir.Expand(logFilePath)
		if err != nil {
			dirExpanded = logFilePath
		}

		ext := filepath.Ext(logFilePath)
		if ext == "" {
			continue
		}

		files, err := filepath.Glob(filepath.Join(filepath.Dir(dirExpanded), "*"+ext))
		if err != nil {
			continue
		}

		for _, filePath := range files {
			contain[filePath] = struct{}{}
			dirs[filepath.Dir(filePath)] = struct{}{}
		}
	}

	output := &LogFileInfos{List: []*LogFileInfo{}, Dirs: map2list(dirs)}

	var used bool

	for filePath := range contain {
		fileInfo, err := os.Stat(filePath)
		if err != nil || fileInfo.IsDir() {
			continue
		}

		used = false

		for _, name := range files {
			if name == filePath {
				used = true
			}
		}
		// fileDate := strings.TrimPrefix(strings.TrimSuffix(filePath, ".log"), strings.TrimSuffix(logFilePath, ".log"))
		// parsedDate, _ := time.Parse(timerotator.FormatDefault, fileDate)
		output.List = append(output.List, &LogFileInfo{
			ID:   strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(filePath)), "="),
			Name: fileInfo.Name(),
			Path: filePath,
			Size: fileInfo.Size(),
			Time: fileInfo.ModTime().Round(time.Second),
			Mode: fileInfo.Mode(),
			Used: used,
			User: getFileOwner(fileInfo),
		})
		output.Size += fileInfo.Size()
	}

	sort.Sort(output)

	return output
}

func (l *LogFileInfos) Len() int {
	return len(l.List)
}

func (l *LogFileInfos) Swap(i, j int) {
	l.List[i], l.List[j] = l.List[j], l.List[i]
}

func (l *LogFileInfos) Less(i, j int) bool {
	return l.List[i].Time.After(l.List[j].Time)
}

func map2list(input map[string]struct{}) []string {
	output := []string{}
	for name := range input {
		output = append(output, name)
	}

	return output
}

func logCounter(filename string, writer io.Writer) io.Writer {
	counter := datacounter.NewWriterCounter(writer)

	mnd.LogFiles.Set("Lines Written: "+filename, expvar.Func(
		func() interface{} { return int64(counter.Writes()) },
	))

	mnd.LogFiles.Set("Bytes Written: "+filename, expvar.Func(
		func() interface{} { return int64(counter.Count()) },
	))

	return counter
}
