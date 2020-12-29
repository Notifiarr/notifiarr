package logs

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"golift.io/rotatorr"
	"golift.io/rotatorr/timerotator"
)

// Logger provides some methods with baked in assumptions.
type Logger struct {
	ErrorLog *log.Logger // Shares a Writer with Logger.
	DebugLog *log.Logger // Shares a Writer with Logger.
	InfoLog  *log.Logger
	HTTPLog  *log.Logger
	web      *rotatorr.Logger
	app      *rotatorr.Logger
	logs     *Logs
}

// satisfy gomnd.
const (
	callDepth = 2 // log the line that called us.
	megabyte  = 1024 * 1024
	defExt    = ".log"
	httpExt   = ".http.log"
)

// Logs allows sending logs to rotating files.
// Setting an AppName will force log creation even if LogFile and HTTPLog are empty.
type Logs struct {
	AppName   string `json:"-"`
	LogFile   string `json:"log_file" toml:"log_file" xml:"log_file" yaml:"log_file"`
	HTTPLog   string `json:"http_log" toml:"http_log" xml:"http_log" yaml:"http_log"`
	LogFiles  int    `json:"log_files" toml:"log_files" xml:"log_files" yaml:"log_files"`
	LogFileMb int    `json:"log_file_mb" toml:"log_file_mb" xml:"log_file_mb" yaml:"log_file_mb"`
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
		logs:     &Logs{},
	}
}

// Rotate rotates the log files.
func (l *Logger) Rotate() (errors []error) {
	if l.web != nil {
		if _, err := l.web.Rotate(); err != nil {
			errors = append(errors, fmt.Errorf("rotating HTTP Log: %w", err))
		}
	}

	if l.app != nil {
		if _, err := l.app.Rotate(); err != nil {
			errors = append(errors, fmt.Errorf("rotating App Log: %w", err))
		}
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
func (l *Logger) SetupLogging(config *Logs) {
	l.logs = config
	l.setLogPaths()
	l.openLogFile()
	l.openHTTPLog()
}

func (l *Logger) setLogPaths() {
	// Make sure log file paths exist if AppName is provided.
	if l.logs.AppName != "" {
		if l.logs.LogFile == "" {
			l.logs.LogFile = filepath.Join("~", ".dnclient", l.logs.AppName+defExt)
		}

		if l.logs.HTTPLog == "" {
			l.logs.HTTPLog = filepath.Join("~", ".dnclient", l.logs.AppName+httpExt)
		}
	}

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
		Filepath: l.logs.LogFile,                                  // log file name.
		FileSize: int64(l.logs.LogFileMb) * megabyte,              // megabytes
		Rotatorr: &timerotator.Layout{FileCount: l.logs.LogFiles}, // number of files to keep.
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
	if l.logs.Debug {
		l.DebugLog.SetOutput(l.InfoLog.Writer())
	}

	l.ErrorLog.SetOutput(l.InfoLog.Writer())
	log.SetOutput(l.InfoLog.Writer())
}

func (l *Logger) openHTTPLog() {
	rotateHTTP := &rotatorr.Config{
		Filepath: l.logs.HTTPLog,                                  // log file name.
		FileSize: int64(l.logs.LogFileMb) * megabyte,              // megabytes
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
