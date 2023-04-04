//go:build !windows && !darwin

package ui

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/favicon.png"

// HasGUI returns false on Linux, true on Windows and optional on macOS.
func HasGUI() bool {
	return false
}

func Notify(_ string, _ ...interface{}) error {
	return nil
}

// HideConsoleWindow doesn't work on most OSes.
func HideConsoleWindow() {}

// ShowConsoleWindow does nothing on OSes besides Windows.
func ShowConsoleWindow() {}

// StartCmd starts a command.
func StartCmd(c string, v ...string) error {
	cmd := exec.Command(c, v...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	return cmd.Start() //nolint:wrapcheck
}

// ErrUnsupported is just an error.
var ErrUnsupported = fmt.Errorf("unsupported OS")

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
