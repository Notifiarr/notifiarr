// Package logs provides the low-level routines for directing log messages.
// It creates several logging channels for debug, info, errors, http, etc.
// These channels are directed to log files and/or stdout depending on how
// the application is configured. This package reads its configuration
// directly from a config file. In here you will find the log rotation
// config for rotatorr, panic redirection, and external logging methods.
package logs

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/logs/share"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	homedir "github.com/mitchellh/go-homedir"
	"golift.io/rotatorr"
	"golift.io/rotatorr/timerotator"
	"golift.io/version"
)

// Logger provides some methods with baked in assumptions.
type Logger struct {
	ErrorLog  *log.Logger // Shares a Writer with InfoLog.
	DebugLog  *log.Logger // Shares a Writer with InfoLog by default. Changeable.
	InfoLog   *log.Logger
	HTTPLog   *log.Logger
	web       *rotatorr.Logger
	app       *rotatorr.Logger
	debug     *rotatorr.Logger
	custom    *rotatorr.Logger // must not be set when web/app/debug are set.
	LogConfig *LogConfig
	mu        sync.RWMutex
}

// These are used for custom logs.
//
//nolint:gochecknoglobals
var (
	logFiles  = 1
	logFileMb = 100
	customLog = make(map[string]*rotatorr.Logger)
)

// Custom errors.
var (
	ErrCloseCustom = errors.New("cannot close custom logs directly")
)

// satisfy gomnd.
const (
	callDepth = 2 // log the line that called us.
	defExt    = ".log"
	httpExt   = ".http.log"
	errStr    = "Error"
	infoStr   = "Info"
	dbugStr   = "Debug"
)

// LogConfig allows sending logs to rotating files.
// Setting an AppName will force log creation even if LogFile and HTTPLog are empty.
type LogConfig struct {
	AppName   string   `json:"-"         toml:"-"           xml:"-"           yaml:"-"`
	LogFile   string   `json:"logFile"   toml:"log_file"    xml:"log_file"    yaml:"logFile"`
	DebugLog  string   `json:"debugLog"  toml:"debug_log"   xml:"debug_log"   yaml:"debugLog"`
	HTTPLog   string   `json:"httpLog"   toml:"http_log"    xml:"http_log"    yaml:"httpLog"`
	LogFiles  int      `json:"logFiles"  toml:"log_files"   xml:"log_files"   yaml:"logFiles"`
	LogFileMb int      `json:"logFileMb" toml:"log_file_mb" xml:"log_file_mb" yaml:"logFileMb"`
	FileMode  FileMode `json:"fileMode"  toml:"file_mode"   xml:"file_mode"   yaml:"fileMode"`
	Debug     bool     `json:"debug"     toml:"debug"       xml:"debug"       yaml:"debug"`
	Quiet     bool     `json:"quiet"     toml:"quiet"       xml:"quiet"       yaml:"quiet"`
	NoUploads bool     `json:"noUploads" toml:"no_uploads"  xml:"no_uploads"  yaml:"noUploads"`
}

// New returns a new Logger with debug off and sends everything to stdout.
func New() *Logger {
	return &Logger{
		DebugLog:  log.New(discard, "[DEBUG] ", log.LstdFlags),
		InfoLog:   log.New(stdout, "[INFO] ", log.LstdFlags),
		ErrorLog:  log.New(stdout, "[ERROR] ", log.LstdFlags),
		HTTPLog:   log.New(stdout, "", log.LstdFlags),
		LogConfig: &LogConfig{},
	}
}

// SetupLogging splits log writers into a file and/or stdout.
func (l *Logger) SetupLogging(config *LogConfig) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fileMode = config.FileMode.Mode()
	logFiles = config.LogFiles
	logFileMb = config.LogFileMb
	l.LogConfig = config
	config.Quiet = !hasConsoleWindow() || config.Quiet

	l.setDefaultLogPaths()
	l.setAppLogPath()
	l.setHTTPLogPath()
	l.openLogFile()
	l.openHTTPLog()
	l.openDebugLog()
}

// Rotate rotates the log files. If called on a custom log, only rotates that log file.
func (l *Logger) Rotate() []error {
	if l.custom != nil {
		if _, err := l.custom.Rotate(); err != nil {
			return []error{fmt.Errorf("rotating Custom Log: %w", err)}
		}
	}

	var errors []error

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
func (l *Logger) Close() []error {
	if l.custom != nil {
		return []error{ErrCloseCustom}
	}

	var errors []error

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

// CapturePanic can be deferred in any go routine to log any panic that occurs.
func (l *Logger) CapturePanic() {
	if r := recover(); r != nil {
		l.ErrorLog.Output(callDepth, //nolint:errcheck
			fmt.Sprintf("Go Panic! %s\n%s-%s %s %v\n%s", mnd.BugIssue,
				version.Version, version.Revision, version.Branch, r, string(debug.Stack())))
		panic(r)
	}
}

// Debug writes log lines... to stdout and/or a file.
func (l *Logger) Debug(v ...interface{}) {
	l.writeMsg(fmt.Sprintln(v...), l.DebugLog, dbugStr, false)
}

// Debugf writes log lines... to stdout and/or a file.
func (l *Logger) Debugf(msg string, v ...interface{}) {
	l.writeMsg(fmt.Sprintf(msg, v...), l.DebugLog, dbugStr, false)
}

// Print writes log lines... to stdout and/or a file.
func (l *Logger) Print(v ...interface{}) {
	l.writeMsg(fmt.Sprintln(v...), l.InfoLog, infoStr, false)
}

// Printf writes log lines... to stdout and/or a file.
func (l *Logger) Printf(msg string, v ...interface{}) {
	l.writeMsg(fmt.Sprintf(msg, v...), l.InfoLog, infoStr, false)
}

// Error writes log lines... to stdout and/or a file.
func (l *Logger) Error(v ...interface{}) {
	l.writeMsg(fmt.Sprintln(v...), l.ErrorLog, errStr, true)
}

// Errorf writes log lines... to stdout and/or a file.
func (l *Logger) Errorf(msg string, v ...interface{}) {
	l.writeMsg(fmt.Sprintf(msg, v...), l.ErrorLog, errStr, true)
}

// ErrorfNoShare writes log lines... to stdout and/or a file.
func (l *Logger) ErrorfNoShare(msg string, v ...interface{}) {
	l.writeMsg(fmt.Sprintf(msg, v...), l.ErrorLog, errStr, false)
}

func (l *Logger) writeMsg(msg string, log *log.Logger, name string, shared bool) {
	if err := log.Output(callDepth, msg); err != nil {
		if errors.Is(err, rotatorr.ErrWriteTooLarge) {
			l.writeSplitMsg(msg, log, name)
			return
		}

		fmt.Println("Logger Error:", err) //nolint:forbidigo
	} else {
		mnd.LogFiles.Add(name+" Lines", 1)
	}

	if shared { // we share errors with the website.
		share.Log(msg)
	}
}

// writeSplitMsg splits the message in half and attempts to write each half.
// If the message is still too large, it'll be split again, and the process continues until it works.
func (l *Logger) writeSplitMsg(msg string, log *log.Logger, name string) {
	half := len(msg) / 2 //nolint:mnd // split messages in half, recursively as needed.
	part1 := msg[:half]
	part2 := "...continuing: " + msg[half:]

	mnd.LogFiles.Add(name+" Splits", 1)
	l.writeMsg(part1, log, name, false)
	l.writeMsg(part2, log, name, false)
}

// CustomLog allows the creation of ad-hoc rotating log files from other packages.
// This is not thread safe with Rotate(), so do not call them at the same time.
func CustomLog(filePath, logName string) *Logger {
	if filePath == "" || logName == "" {
		return &Logger{
			DebugLog: log.New(discard, "", 0),
			InfoLog:  log.New(discard, "", 0),
			ErrorLog: log.New(discard, "", 0),
			HTTPLog:  log.New(discard, "", 0),
		}
	}

	if f, err := homedir.Expand(filePath); err == nil {
		filePath = f
	}

	if f, err := filepath.Abs(filePath); err == nil {
		filePath = f
	}

	logger := &Logger{}
	customLog[logName] = rotatorr.NewMust(&rotatorr.Config{
		Filepath: filePath,                        // log file name.
		FileSize: int64(logFileMb) * mnd.Megabyte, // mnd.Megabytes
		FileMode: fileMode,                        // set file mode.
		Rotatorr: &timerotator.Layout{
			FileCount:  logFiles, // number of files to keep.
			PostRotate: logger.postRotateCounter,
		},
	})
	logs := logCounter(filePath, customLog[logName])
	logger.DebugLog = log.New(logs, "[DEBUG] ", log.LstdFlags)
	logger.InfoLog = log.New(logs, "[INFO] ", log.LstdFlags)
	logger.ErrorLog = log.New(logs, "[ERROR] ", log.LstdFlags)
	logger.HTTPLog = log.New(logs, "[HTTP] ", log.LstdFlags)
	logger.custom = customLog[logName]

	return logger
}

func (l *Logger) postRotateCounter(fileName, newFile string) {
	if fileName != "" {
		mnd.LogFiles.Add("Rotated: "+fileName, 1)
	}

	if newFile != "" && l != nil {
		go l.Printf("Rotated log file to: %s", newFile)
	}
}

func (l *Logger) GetInfoLog() *log.Logger {
	return l.InfoLog
}

func (l *Logger) GetErrorLog() *log.Logger {
	return l.ErrorLog
}

func (l *Logger) GetDebugLog() *log.Logger {
	return l.DebugLog
}

func (l *Logger) DebugEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.LogConfig.Debug
}
