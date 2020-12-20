package dnclient

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"golift.io/rotatorr"
	"golift.io/rotatorr/timerotator"
)

// satisfy gomnd.
const (
	helpLink  = "GoLift Discord: https://golift.io/discord" // prints on start and on exit.
	callDepth = 2                                           // log the line that called us.
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

// setupLogging splits log write into a file and/or stdout.
func (c *Client) setupLogging() {
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

// logStartupInfo prints info about our startup config.
func (c *Client) logStartupInfo() {
	c.Printf("==> %s <==", helpLink)
	c.Print("==> Startup Settings <==")
	c.logSonarr()
	c.logRadarr()
	c.logLidarr()
	c.logReadarr()
	c.Print(" => Debug / Quiet:", c.Config.Debug, "/", c.Config.Quiet)

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		c.Print(" => Web HTTPS Listen:", "https://"+c.Config.BindAddr+path.Join("/", c.Config.WebRoot))
		c.Print(" => Web Cert & Key Files:", c.Config.SSLCrtFile+", "+c.Config.SSLKeyFile)
	} else {
		c.Print(" => Web HTTP Listen:", "http://"+c.Config.BindAddr+path.Join("/", c.Config.WebRoot))
	}

	if c.Config.LogFile != "" {
		msg := "no rotation"
		if c.Config.LogFiles > 0 {
			msg = fmt.Sprintf("%d @ %dMb", c.Config.LogFiles, c.Config.LogFileMb)
		}

		c.Printf(" => Log File: %s (%s)", c.Config.LogFile, msg)
	}
}

func (c *Client) logExitInfo() error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	c.Printf("[%s] Need help? %s\n=====> Exiting! Caught Signal: %v", c.Flags.Name(), helpLink, <-sig)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return c.server.Shutdown(ctx)
}
