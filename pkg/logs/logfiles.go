package logs

import (
	"encoding/base64"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	homedir "github.com/mitchellh/go-homedir"
	"golift.io/rotatorr"
	"golift.io/rotatorr/timerotator"
)

// setDefaultLogPaths makes sure a GUI app has log files defined.
// These are enforced on GUI OSes (like macOS.app and Windows).
func (l *Logger) setDefaultLogPaths() {
	// Make sure log file paths exist if AppName is provided; this indicates GUI OS.
	if l.LogConfig.AppName != "" {
		if l.LogConfig.LogFile == "" {
			l.LogConfig.LogFile = filepath.Join("~", ".notifiarr", l.LogConfig.AppName+defExt)
		}

		if l.LogConfig.HTTPLog == "" {
			l.LogConfig.HTTPLog = filepath.Join("~", ".notifiarr", l.LogConfig.AppName+httpExt)
		}
	}
}

// setLogPaths sets the log paths for app and http logs.
func (l *Logger) setLogPaths() {
	// Regular log file.
	if l.LogConfig.LogFile != "" {
		if f, err := homedir.Expand(l.LogConfig.LogFile); err == nil {
			l.LogConfig.LogFile = f
		} else if l.LogConfig.AppName != "" {
			l.LogConfig.LogFile = l.LogConfig.AppName + defExt
		}

		if f, err := filepath.Abs(l.LogConfig.LogFile); err == nil {
			l.LogConfig.LogFile = f
		}
	}

	// HTTP log file.
	if l.LogConfig.HTTPLog != "" {
		if f, err := homedir.Expand(l.LogConfig.HTTPLog); err == nil {
			l.LogConfig.HTTPLog = f
		} else if l.LogConfig.AppName != "" {
			l.LogConfig.HTTPLog = l.LogConfig.AppName + httpExt
		}

		if f, err := filepath.Abs(l.LogConfig.HTTPLog); err == nil {
			l.LogConfig.HTTPLog = f
		}
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
		l.InfoLog.SetOutput(io.MultiWriter(l.app, os.Stdout))
	case !l.LogConfig.Quiet && l.LogConfig.LogFile == "":
		l.InfoLog.SetOutput(os.Stdout)
	case l.LogConfig.LogFile == "":
		l.InfoLog.SetOutput(ioutil.Discard) // default is "nothing"
	default:
		l.app = rotatorr.NewMust(rotate)
		l.InfoLog.SetOutput(l.app)
	}

	// Don't forget errors log, and do standard logger too.
	if l.LogConfig.Debug && l.LogConfig.DebugLog == "" {
		l.DebugLog.SetOutput(l.InfoLog.Writer())
	}

	l.ErrorLog.SetOutput(l.InfoLog.Writer())
	log.SetOutput(l.InfoLog.Writer())
	l.postLogRotate("", "")
}

func (l *Logger) postLogRotate(_, newFile string) {
	if newFile != "" {
		go l.Printf("Rotated log file to: %s", newFile)
	}

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

	rotateDebug := &rotatorr.Config{
		Filepath: l.LogConfig.DebugLog,                                 // log file name.
		FileSize: int64(l.LogConfig.LogFileMb) * mnd.Megabyte,          // mnd.Megabytes
		FileMode: l.LogConfig.FileMode.Mode(),                          // set file mode.
		Rotatorr: &timerotator.Layout{FileCount: l.LogConfig.LogFiles}, // number of files to keep.
	}
	l.debug = rotatorr.NewMust(rotateDebug)

	if l.LogConfig.Quiet {
		l.DebugLog.SetOutput(l.debug)
	} else {
		l.DebugLog.SetOutput(io.MultiWriter(l.debug, os.Stdout))
	}
}

func (l *Logger) openHTTPLog() {
	rotateHTTP := &rotatorr.Config{
		Filepath: l.LogConfig.HTTPLog,                                  // log file name.
		FileSize: int64(l.LogConfig.LogFileMb) * mnd.Megabyte,          // mnd.Megabytes
		FileMode: l.LogConfig.FileMode.Mode(),                          // set file mode.
		Rotatorr: &timerotator.Layout{FileCount: l.LogConfig.LogFiles}, // number of files to keep.
	}

	switch { // only use MultiWriter if we have > 1 writer.
	case !l.LogConfig.Quiet && l.LogConfig.HTTPLog != "":
		l.web = rotatorr.NewMust(rotateHTTP)
		l.HTTPLog.SetOutput(io.MultiWriter(l.web, os.Stdout))
	case !l.LogConfig.Quiet && l.LogConfig.HTTPLog == "":
		l.HTTPLog.SetOutput(os.Stdout)
	case l.LogConfig.HTTPLog == "":
		l.HTTPLog.SetOutput(ioutil.Discard) // default is "nothing"
	default:
		l.web = rotatorr.NewMust(rotateHTTP)
		l.HTTPLog.SetOutput(l.web)
	}
}

type LogFileInfos struct {
	Dirs []string
	Size int64
	Logs []*LogFileInfo
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

// GetAllLogFilePaths searches the disk for log file names.
func (l *Logger) GetAllLogFilePaths() *LogFileInfos { //nolint:funlen,cyclop
	logFiles := []string{
		l.LogConfig.LogFile,
		l.LogConfig.HTTPLog,
		l.LogConfig.DebugLog,
	}

	for cust := range customLog {
		logFiles = append(logFiles, cust)
	}

	contain := make(map[string]struct{})
	dirs := make(map[string]struct{})

	for _, logFilePath := range logFiles {
		dirExpanded, err := homedir.Expand(logFilePath)
		if err != nil {
			dirExpanded = logFilePath
		}

		files, err := filepath.Glob(filepath.Join(filepath.Dir(dirExpanded), "*.log"))
		if err != nil {
			continue
		}

		for _, filePath := range files {
			contain[filePath] = struct{}{}
			dirs[filepath.Dir(filePath)] = struct{}{}
		}
	}

	output := &LogFileInfos{Logs: []*LogFileInfo{}, Dirs: map2list(dirs)}

	for filePath := range contain {
		fileInfo, err := os.Stat(filePath)
		if err != nil || fileInfo.IsDir() {
			continue
		}

		used := false

		for _, name := range logFiles {
			if name == filePath {
				used = true
			}
		}
		// fileDate := strings.TrimPrefix(strings.TrimSuffix(filePath, ".log"), strings.TrimSuffix(logFilePath, ".log"))
		// parsedDate, _ := time.Parse(timerotator.FormatDefault, fileDate)
		output.Logs = append(output.Logs, &LogFileInfo{
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
	return len(l.Logs)
}

func (l *LogFileInfos) Swap(i, j int) {
	l.Logs[i], l.Logs[j] = l.Logs[j], l.Logs[i]
}

func (l *LogFileInfos) Less(i, j int) bool {
	return l.Logs[i].Name < l.Logs[j].Name
}

func map2list(input map[string]struct{}) []string {
	output := []string{}
	for name := range input {
		output = append(output, name)
	}

	return output
}
