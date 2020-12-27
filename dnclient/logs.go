package dnclient

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Go-Lift-TV/discordnotifier-client/ui"
	homedir "github.com/mitchellh/go-homedir"
	"golift.io/rotatorr"
	"golift.io/rotatorr/timerotator"
)

// Logger provides a struct we can pass into other packages.
type Logger struct {
	Errors    *log.Logger // Shares a Writer with Logger.
	Logger    *log.Logger
	Requests  *log.Logger
	webrotate *rotatorr.Logger
	logrotate *rotatorr.Logger
}

// satisfy gomnd.
const (
	helpLink  = "GoLift Discord: https://golift.io/discord"
	callDepth = 2 // log the line that called us.
	megabyte  = 1024 * 1024
)

// Print writes log lines... to stdout and/or a file.
func (l *Logger) Print(v ...interface{}) {
	err := l.Logger.Output(callDepth, fmt.Sprintln(v...))
	if err != nil {
		fmt.Println("Logger Error:", err)
	}
}

// Printf writes log lines... to stdout and/or a file.
func (l *Logger) Printf(msg string, v ...interface{}) {
	err := l.Logger.Output(callDepth, fmt.Sprintf(msg, v...))
	if err != nil {
		fmt.Println("Logger Error:", err)
	}
}

// Errorf writes log lines... to stdout and/or a file.
func (l *Logger) Errorf(msg string, v ...interface{}) {
	err := l.Errors.Output(callDepth, fmt.Sprintf(msg, v...))
	if err != nil {
		fmt.Println("Logger Error:", err)
	}
}

// SetupLogging splits log write into a file and/or stdout.
func (c *Client) SetupLogging() {
	if ui.HasGUI() && c.Config.LogFile == "" {
		f, err := homedir.Expand(filepath.Join("~", ".dnclient", c.Flags.Name()+".log"))
		if err != nil {
			c.Config.LogFile = c.Flags.Name() + ".log"
		} else {
			c.Config.LogFile = f
		}
	}

	if ui.HasGUI() && c.Config.HTTPLog == "" {
		f, err := homedir.Expand(filepath.Join("~", ".dnclient", c.Flags.Name()+".http.log"))
		if err != nil {
			c.Config.HTTPLog = c.Flags.Name() + ".error.log"
		} else {
			c.Config.HTTPLog = f
		}
	}

	rotate := &rotatorr.Config{
		Filepath: c.Config.LogFile,                                  // log file name.
		FileSize: int64(c.Config.LogFileMb) * megabyte,              // megabytes
		Rotatorr: &timerotator.Layout{FileCount: c.Config.LogFiles}, // number of files to keep.
	}

	switch { // only use MultiWriter if we have > 1 writer.
	case !c.Config.Quiet && c.Config.LogFile != "":
		c.Logger.logrotate = rotatorr.NewMust(rotate)
		writer := io.MultiWriter(c.Logger.logrotate, os.Stdout)
		c.Logger.Logger.SetOutput(writer)
		c.Logger.Errors.SetOutput(writer)
	case !c.Config.Quiet && c.Config.LogFile == "":
		c.Logger.Logger.SetOutput(os.Stdout)
		c.Logger.Errors.SetOutput(os.Stdout)
	case c.Config.LogFile == "":
		c.Logger.Logger.SetOutput(ioutil.Discard) // default is "nothing"
		c.Logger.Errors.SetOutput(ioutil.Discard) // default is "nothing"
	default:
		c.Logger.logrotate = rotatorr.NewMust(rotate)
		c.Logger.Logger.SetOutput(c.Logger.logrotate)
		c.Logger.Errors.SetOutput(c.Logger.logrotate)
	}

	rotateHTTP := &rotatorr.Config{
		Filepath: c.Config.HTTPLog,                                  // log file name.
		FileSize: int64(c.Config.LogFileMb) * megabyte,              // megabytes
		Rotatorr: &timerotator.Layout{FileCount: c.Config.LogFiles}, // number of files to keep.
	}

	switch { // only use MultiWriter if we have > 1 writer.
	case !c.Config.Quiet && c.Config.HTTPLog != "":
		c.Logger.webrotate = rotatorr.NewMust(rotateHTTP)
		c.Logger.Requests.SetOutput(io.MultiWriter(c.Logger.webrotate, os.Stdout))
	case !c.Config.Quiet && c.Config.HTTPLog == "":
		c.Logger.Requests.SetOutput(os.Stdout)
	case c.Config.HTTPLog == "":
		c.Logger.Requests.SetOutput(ioutil.Discard) // default is "nothing"
	default:
		c.Logger.webrotate = rotatorr.NewMust(rotateHTTP)
		c.Logger.Requests.SetOutput(c.Logger.webrotate)
	}
}
