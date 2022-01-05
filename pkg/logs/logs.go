// Package logs provides the low-level routines for directing log messages.
// It creates several logging channels for debug, info, errors, http, etc.
// These channels are directed to log files and/or stdout depending on how
// the application is configured. This package reads its configuration
// directly from a config file. In here you will find the log roatation
// config for rotatorr, panic redirection, and external logging methods.
package logs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
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
	AppName   string   `json:"-"`
	LogFile   string   `json:"logFile" toml:"log_file" xml:"log_file" yaml:"logFile"`
	DebugLog  string   `json:"debugLog" toml:"debug_log" xml:"debug_log" yaml:"debugLog"`
	HTTPLog   string   `json:"httpLog" toml:"http_log" xml:"http_log" yaml:"httpLog"`
	LogFiles  int      `json:"logFiles" toml:"log_files" xml:"log_files" yaml:"logFiles"`
	LogFileMb int      `json:"logFileMb" toml:"log_file_mb" xml:"log_file_mb" yaml:"logFileMb"`
	FileMode  FileMode `json:"fileMode" toml:"file_mode" xml:"file_mode" yaml:"fileMode"`
	Debug     bool     `json:"debug" toml:"debug" xml:"debug" yaml:"debug"`
	Quiet     bool     `json:"quiet" toml:"quiet" xml:"quiet" yaml:"quiet"`
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

// SetupLogging splits log writers into a file and/or stdout.
func (l *Logger) SetupLogging(config *LogConfig) {
	l.logs = config
	l.setDefaultLogPaths()
	l.setLogPaths()
	l.openLogFile()
	l.openHTTPLog()
	l.openDebugLog()
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

// CapturePanic can be defered in any go routine to log any panic that occurs.
func (l *Logger) CapturePanic() {
	if r := recover(); r != nil {
		ui.ShowConsoleWindow()
		l.ErrorLog.Output(callDepth, //nolint:errcheck
			fmt.Sprintf("Go Panic! %s\n%v\n%s", mnd.BugIssue, r, string(debug.Stack())))
		panic(r)
	}
}

// Debug writes log lines... to stdout and/or a file.
func (l *Logger) Debug(v ...interface{}) {
	err := l.DebugLog.Output(callDepth, fmt.Sprintln(v...))
	if err != nil {
		fmt.Println("Logger Error:", err) //nolint:forbidigo
	}
}

// Debugf writes log lines... to stdout and/or a file.
func (l *Logger) Debugf(msg string, v ...interface{}) {
	err := l.DebugLog.Output(callDepth, fmt.Sprintf(msg, v...))
	if err != nil {
		fmt.Println("Logger Error:", err) //nolint:forbidigo
	}
}

// Print writes log lines... to stdout and/or a file.
func (l *Logger) Print(v ...interface{}) {
	err := l.InfoLog.Output(callDepth, fmt.Sprintln(v...))
	if err != nil {
		fmt.Println("Logger Error:", err) //nolint:forbidigo
	}
}

// Printf writes log lines... to stdout and/or a file.
func (l *Logger) Printf(msg string, v ...interface{}) {
	err := l.InfoLog.Output(callDepth, fmt.Sprintf(msg, v...))
	if err != nil {
		fmt.Println("Logger Error:", err) //nolint:forbidigo
	}
}

// Error writes log lines... to stdout and/or a file.
func (l *Logger) Error(v ...interface{}) {
	err := l.ErrorLog.Output(callDepth, fmt.Sprintln(v...))
	if err != nil {
		fmt.Println("Logger Error:", err) //nolint:forbidigo
	}
}

// Errorf writes log lines... to stdout and/or a file.
func (l *Logger) Errorf(msg string, v ...interface{}) {
	err := l.ErrorLog.Output(callDepth, fmt.Sprintf(msg, v...))
	if err != nil {
		fmt.Println("Logger Error:", err) //nolint:forbidigo
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

	if f, err := homedir.Expand(filePath); err == nil {
		filePath = f
	}

	if f, err := filepath.Abs(filePath); err == nil {
		filePath = f
	}

	customLog[logName] = rotatorr.NewMust(&rotatorr.Config{
		Filepath: filePath,                                 // log file name.
		FileSize: int64(logFileMb) * mnd.Megabyte,          // mnd.Megabytes
		FileMode: fileMode,                                 // set file mode.
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
