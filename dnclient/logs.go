package dnclient

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

// satisfy gomnd.
const (
	helpLink  = "GoLift Discord: https://golift.io/discord"
	callDepth = 2 // log the line that called us.
	megabyte  = 1024 * 1024
)

// Debugf writes Debug log lines... to stdout and/or a file.
func (l *Logger) Debugf(msg string, v ...interface{}) {
	if l.debug {
		err := l.Logger.Output(callDepth, "[DEBUG] "+fmt.Sprintf(msg, v...))
		if err != nil {
			fmt.Println("Logger Error:", err)
		}
	}
}

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

// SetupLogging splits log write into a file and/or stdout.
func (c *Client) SetupLogging() {
	if hasGUI() && c.Config.LogFile == "" {
		f, err := homedir.Expand(filepath.Join("~", ".dnclient", c.Flags.Name()+".log"))
		if err != nil {
			c.Config.LogFile = c.Flags.Name() + ".log"
		} else {
			c.Config.LogFile = f
		}
	}

	c.Logger = &Logger{
		debug:  c.Config.Debug,
		Logger: log.New(ioutil.Discard, "", log.LstdFlags),
	}

	if c.Config.Debug {
		c.Logger.Logger.SetFlags(log.Lshortfile | log.Lmicroseconds | log.Ldate)
	}

	rotate := &rotatorr.Config{
		Filepath: c.Config.LogFile,                                  // log file name.
		FileSize: int64(c.Config.LogFileMb) * megabyte,              // megabytes
		Rotatorr: &timerotator.Layout{FileCount: c.Config.LogFiles}, // number of files to keep.
	}

	switch { // only use MultiWriter if we have > 1 writer.
	case !c.Config.Quiet && c.Config.LogFile != "":
		c.Logger.Logger.SetOutput(io.MultiWriter(rotatorr.NewMust(rotate), os.Stdout))
	case !c.Config.Quiet && c.Config.LogFile == "":
		c.Logger.Logger.SetOutput(os.Stdout)
	case c.Config.LogFile == "":
		c.Logger.Logger.SetOutput(ioutil.Discard) // default is "nothing"
	default:
		c.Logger.Logger.SetOutput(rotatorr.NewMust(rotate))
	}
}
