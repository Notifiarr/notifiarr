// Package logs provides the low-level routines for directing log messages.
// It creates several logging channels for debug, info, errors, http, etc.
// These channels are directed to log files and/or stdout depending on how
// the application is configured. This package reads its configuration
// directly from a config file. In here you will find the log roatation
// config for rotatorr, panic redirection, and external logging methods.
package logs

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	homedir "github.com/mitchellh/go-homedir"
	"golift.io/rotatorr"
	"golift.io/rotatorr/timerotator"
)

// Logger provides some methods with baked in assumptions.
type Logger struct {
	ErrorLog *log.Logger // Shares a Writer with InfoLog.
	DebugLog *log.Logger // Shares a Writer with InfoLog by default. Changeable.
	InfoLog  *log.Logger
	HTTPLog  *log.Logger
	web      *rotatorr.Logger
	app      *rotatorr.Logger
	debug    *rotatorr.Logger
	custom   *rotatorr.Logger // must not be set when web/app/debug are set.
	logs     *LogConfig
}

// These are used for custom logs.
// nolint:gochecknoglobals
var (
	logFiles  = 1
	logFileMb = 100
	fileMode  = uint32(rotatorr.FileMode)
	customLog = make(map[string]*rotatorr.Logger)
)

// Custom errors.
var (
	ErrCloseCustom = fmt.Errorf("cannot close custom logs directly")
)

// satisfy gomnd.
const (
	callDepth = 2 // log the line that called us.
	defExt    = ".log"
	httpExt   = ".http.log"
)

// LogConfig allows sending logs to rotating files.
// Setting an AppName will force log creation even if LogFile and HTTPLog are empty.
type LogConfig struct {
	AppName   string `json:"-"`
	LogFile   string `json:"log_file" toml:"log_file" xml:"log_file" yaml:"log_file"`
	DebugLog  string `json:"debug_log" toml:"debug_log" xml:"debug_log" yaml:"debug_log"`
	HTTPLog   string `json:"http_log" toml:"http_log" xml:"http_log" yaml:"http_log"`
	LogFiles  int    `json:"log_files" toml:"log_files" xml:"log_files" yaml:"log_files"`
	LogFileMb int    `json:"log_file_mb" toml:"log_file_mb" xml:"log_file_mb" yaml:"log_file_mb"`
	FileMode  uint32 `json:"file_mode" toml:"file_mode" xml:"file_mode" yaml:"file_mode"`
	Debug     bool   `json:"debug" toml:"debug" xml:"debug" yaml:"debug"`
	Quiet     bool   `json:"quiet" toml:"quiet" xml:"quiet" yaml:"quiet"`
}

// New returns a new Logger with debug off and sends everything to stdout.
func New() *Logger {
	return &Logger{
		DebugLog: log.New(ioutil.Discard, "[DEBUG] ", log.LstdFlags),
		InfoLog:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		ErrorLog: log.New(os.Stdout, "[ERROR] ", log.LstdFlags),
		HTTPLog:  log.New(os.Stdout, "", log.LstdFlags),
		logs:     &LogConfig{},
	}
}

// Rotate rotates the log files. If called on a custom log, only rotates that log file.
func (l *Logger) Rotate() (errors []error) {
	if l.custom != nil {
		if _, err := l.custom.Rotate(); err != nil {
			return []error{fmt.Errorf("rotating cCustom Log: %w", err)}
		}
	}

	for name, logger := range map[string]*rotatorr.Logger{
		"HTTP":  l.web,
		"App":   l.app,
		"Debug": l.debug,
	} {
		if logger != nil {
			if _, err := logger.Rotate(); err != nil {
				errors = append(errors, fmt.Errorf("closing %s Log: %w", name, err))
			}
		}
	}

	for name, logger := range customLog {
		if _, err := logger.Rotate(); err != nil {
			errors = append(errors, fmt.Errorf("rotating %s Log: %w", name, err))
		}
	}

	return errors
}

// Close closes all open log files. Does not work on custom logs.
func (l *Logger) Close() (errors []error) {
	if l.custom != nil {
		return []error{ErrCloseCustom}
	}

	for name, logger := range map[string]*rotatorr.Logger{
		"HTTP":  l.web,
		"App":   l.app,
		"Debug": l.debug,
	} {
		if logger != nil {
			if err := logger.Close(); err != nil {
				errors = append(errors, fmt.Errorf("closing %s Log: %w", name, err))
			}
		}
	}

	l.web = nil
	l.app = nil
	l.debug = nil

	for name, logger := range customLog {
		if err := logger.Close(); err != nil {
			errors = append(errors, fmt.Errorf("closing %s Log: %w", name, err))
		}

		delete(customLog, name)
	}

	return errors
}

// Debug writes log lines... to stdout and/or a file.
func (l *Logger) Debug(v ...interface{}) {
	err := l.DebugLog.Output(callDepth, fmt.Sprintln(v...))
	if err != nil {
		fmt.Println("Logger Error:", err)
	}
}

// Debugf writes log lines... to stdout and/or a file.
func (l *Logger) Debugf(msg string, v ...interface{}) {
	err := l.DebugLog.Output(callDepth, fmt.Sprintf(msg, v...))
	if err != nil {
		fmt.Println("Logger Error:", err)
	}
}

// Print writes log lines... to stdout and/or a file.
func (l *Logger) Print(v ...interface{}) {
	err := l.InfoLog.Output(callDepth, fmt.Sprintln(v...))
	if err != nil {
		fmt.Println("Logger Error:", err)
	}
}

// Printf writes log lines... to stdout and/or a file.
func (l *Logger) Printf(msg string, v ...interface{}) {
	err := l.InfoLog.Output(callDepth, fmt.Sprintf(msg, v...))
	if err != nil {
		fmt.Println("Logger Error:", err)
	}
}

// Error writes log lines... to stdout and/or a file.
func (l *Logger) Error(v ...interface{}) {
	err := l.ErrorLog.Output(callDepth, fmt.Sprintln(v...))
	if err != nil {
		fmt.Println("Logger Error:", err)
	}
}

// Errorf writes log lines... to stdout and/or a file.
func (l *Logger) Errorf(msg string, v ...interface{}) {
	err := l.ErrorLog.Output(callDepth, fmt.Sprintf(msg, v...))
	if err != nil {
		fmt.Println("Logger Error:", err)
	}
}

// SetupLogging splits log writers into a file and/or stdout.
func (l *Logger) SetupLogging(config *LogConfig) {
	logFiles = config.LogFiles
	logFileMb = config.LogFileMb
	fileMode = config.FileMode
	l.logs = config
	l.setDefaultLogPaths()
	l.setLogPaths()
	l.openLogFile()
	l.openHTTPLog()
	l.openDebugLog()
}

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
		FileMode: os.FileMode(l.logs.FileMode),
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
		Filepath: l.logs.DebugLog,                        // log file name.
		FileSize: int64(l.logs.LogFileMb) * mnd.Megabyte, // mnd.Megabytes
		FileMode: os.FileMode(l.logs.FileMode),
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
		Filepath: l.logs.HTTPLog,                         // log file name.
		FileSize: int64(l.logs.LogFileMb) * mnd.Megabyte, // mnd.Megabytes
		FileMode: os.FileMode(l.logs.FileMode),
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

// CustomLog allows the creation of ad-hoc rotating log files from other packages.
// This is not thread safe with Rotate(), so do not call them at the same time.
func CustomLog(filePath, logName string) *Logger {
	if filePath == "" || logName == "" {
		return &Logger{
			DebugLog: log.New(ioutil.Discard, "", 0),
			InfoLog:  log.New(ioutil.Discard, "", 0),
			ErrorLog: log.New(ioutil.Discard, "", 0),
			HTTPLog:  log.New(ioutil.Discard, "", 0),
		}
	}

	f, err := homedir.Expand(filePath)
	if err == nil {
		filePath = f
	}

	if f, err = filepath.Abs(filePath); err == nil {
		filePath = f
	}

	customLog[logName] = rotatorr.NewMust(&rotatorr.Config{
		Filepath: filePath,                        // log file name.
		FileSize: int64(logFileMb) * mnd.Megabyte, // mnd.Megabytes
		FileMode: os.FileMode(fileMode),
		Rotatorr: &timerotator.Layout{FileCount: logFiles}, // number of files to keep.
	})

	return &Logger{
		custom:   customLog[logName],
		DebugLog: log.New(customLog[logName], "[DEBUG] ", log.LstdFlags),
		InfoLog:  log.New(customLog[logName], "[INFO] ", log.LstdFlags),
		ErrorLog: log.New(customLog[logName], "[ERROR] ", log.LstdFlags),
		HTTPLog:  log.New(customLog[logName], "[HTTP] ", log.LstdFlags),
	}
}
