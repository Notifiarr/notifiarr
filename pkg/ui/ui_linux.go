package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/favicon.png"

//nolint:gochecknoglobals
var hasGUI = os.Getenv("USEGUI") == "true"

// HasGUI tries to determine if the app was invoked as a GUI app.
func HasGUI() bool {
	return hasGUI
}

func Notify(_ string, _ ...interface{}) error {
	return nil
}

// StartCmd starts a command.
func StartCmd(c string, v ...string) error {
	cmd := exec.Command(c, v...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	return cmd.Start() //nolint:wrapcheck
}

// ErrUnsupported is just an error.
var ErrUnsupported = errors.New("unsupported OS, good with xwindows? help us")

// OpenCmd opens anything.
func OpenCmd(cmd ...string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnsupported, runtime.GOOS, cmd)
}

// OpenURL opens URL Links.
func OpenURL(url string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnsupported, runtime.GOOS, url)
}

// OpenLog opens Log Files.
func OpenLog(logFile string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnsupported, runtime.GOOS, logFile)
}

// OpenFile open Config Files.
func OpenFile(filePath string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnsupported, runtime.GOOS, filePath)
}