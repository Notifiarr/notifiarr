package logs

import (
	"encoding/base64"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
	if l.logs.AppName != "" {
		if l.logs.LogFile == "" {
			l.logs.LogFile = filepath.Join("~", ".notifiarr", l.logs.AppName+defExt)
		}

		if l.logs.HTTPLog == "" {
			l.logs.HTTPLog = filepath.Join("~", ".notifiarr", l.logs.AppName+httpExt)
		}
	}
}

// setLogPaths sets the log paths for app and http logs.
func (l *Logger) setLogPaths() {
	// Regular log file.
	if l.logs.LogFile != "" {
		if f, err := homedir.Expand(l.logs.LogFile); err == nil {
			l.logs.LogFile = f
		} else if l.logs.AppName != "" {
			l.logs.LogFile = l.logs.AppName + defExt
		}

		if f, err := filepath.Abs(l.logs.LogFile); err == nil {
			l.logs.LogFile = f
		}
	}

	// HTTP log file.
	if l.logs.HTTPLog != "" {
		if f, err := homedir.Expand(l.logs.HTTPLog); err == nil {
			l.logs.HTTPLog = f
		} else if l.logs.AppName != "" {
			l.logs.HTTPLog = l.logs.AppName + httpExt
		}

		if f, err := filepath.Abs(l.logs.HTTPLog); err == nil {
			l.logs.HTTPLog = f
		}
	}
}

func (l *Logger) openLogFile() {
	rotate := &rotatorr.Config{
		Filepath: l.logs.LogFile,                         // log file name.
		FileSize: int64(l.logs.LogFileMb) * mnd.Megabyte, // mnd.Megabytes
		FileMode: l.logs.FileMode.Mode(),                 // set file mode.
		Rotatorr: &timerotator.Layout{
			FileCount:  l.logs.LogFiles, // number of files to keep.
			PostRotate: l.postLogRotate, // method to run after rotating.
		},
	}

	switch { // only use MultiWriter if we have > 1 writer.
	case !l.logs.Quiet && l.logs.LogFile != "":
		l.app = rotatorr.NewMust(rotate)
		l.InfoLog.SetOutput(io.MultiWriter(l.app, os.Stdout))
	case !l.logs.Quiet && l.logs.LogFile == "":
		l.InfoLog.SetOutput(os.Stdout)
	case l.logs.LogFile == "":
		l.InfoLog.SetOutput(ioutil.Discard) // default is "nothing"
	default:
		l.app = rotatorr.NewMust(rotate)
		l.InfoLog.SetOutput(l.app)
	}

	// Don't forget errors log, and do standard logger too.
	if l.logs.Debug && l.logs.DebugLog == "" {
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
	if !l.logs.Debug {
		// in case we're reloading without debug and had it before.
		l.DebugLog.SetOutput(io.Discard)
	}

	if !l.logs.Debug || l.logs.DebugLog == "" {
		return
	}

	if f, err := homedir.Expand(l.logs.DebugLog); err == nil {
		l.logs.DebugLog = f
	}

	if f, err := filepath.Abs(l.logs.DebugLog); err == nil {
		l.logs.DebugLog = f
	}

	rotateDebug := &rotatorr.Config{
		Filepath: l.logs.DebugLog,                                 // log file name.
		FileSize: int64(l.logs.LogFileMb) * mnd.Megabyte,          // mnd.Megabytes
		FileMode: l.logs.FileMode.Mode(),                          // set file mode.
		Rotatorr: &timerotator.Layout{FileCount: l.logs.LogFiles}, // number of files to keep.
	}
	l.debug = rotatorr.NewMust(rotateDebug)

	if l.logs.Quiet {
		l.DebugLog.SetOutput(l.debug)
	} else {
		l.DebugLog.SetOutput(io.MultiWriter(l.debug, os.Stdout))
	}
}

func (l *Logger) openHTTPLog() {
	rotateHTTP := &rotatorr.Config{
		Filepath: l.logs.HTTPLog,                                  // log file name.
		FileSize: int64(l.logs.LogFileMb) * mnd.Megabyte,          // mnd.Megabytes
		FileMode: l.logs.FileMode.Mode(),                          // set file mode.
		Rotatorr: &timerotator.Layout{FileCount: l.logs.LogFiles}, // number of files to keep.
	}

	switch { // only use MultiWriter if we have > 1 writer.
	case !l.logs.Quiet && l.logs.HTTPLog != "":
		l.web = rotatorr.NewMust(rotateHTTP)
		l.HTTPLog.SetOutput(io.MultiWriter(l.web, os.Stdout))
	case !l.logs.Quiet && l.logs.HTTPLog == "":
		l.HTTPLog.SetOutput(os.Stdout)
	case l.logs.HTTPLog == "":
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
	Path string
	Size int64
	Time time.Time
	Mode fs.FileMode
	Used bool
	User string
}

// GetAllLogFilePaths searches the disk for log file names.
func (l *Logger) GetAllLogFilePaths() *LogFileInfos {
	logFiles := []string{
		l.logs.LogFile,
		l.logs.HTTPLog,
		l.logs.DebugLog,
	}

	for cust := range customLog {
		logFiles = append(logFiles, cust)
	}

	contain := make(map[string]struct{})
	dirs := make(map[string]struct{})

	for _, logFilePath := range logFiles {
		files, err := filepath.Glob(filepath.Join(filepath.Dir(logFilePath), "*.log"))
		if err != nil {
			continue
		}

		dirs[filepath.Dir(logFilePath)] = struct{}{}

		for _, filePath := range files {
			contain[filePath] = struct{}{}
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
			Path: filePath,
			Size: fileInfo.Size(),
			Time: fileInfo.ModTime().Round(time.Second),
			Mode: fileInfo.Mode(),
			Used: used,
			User: getFileOwner(fileInfo),
		})
		output.Size += fileInfo.Size()
	}

	return output
}

func map2list(input map[string]struct{}) []string {
	output := []string{}
	for name := range input {
		output = append(output, name)
	}

	return output
}
