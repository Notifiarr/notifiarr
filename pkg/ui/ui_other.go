//go:build !windows && !darwin && !linux && !freebsd

/* Not sure what OS this will get. */

package ui

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

// ErrUnsupported is just an error.
var ErrUnsupported = errors.New("unsupported OS")

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/favicon.png"

// HasGUI returns false on this gui-unsupported OS.
func HasGUI() bool {
	return false
}

// Toast does not do anything on this OS.
func Toast(_ string, _ ...interface{}) error {
	return nil
}

// StartCmd starts a command.
func StartCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	return cmd.Start() //nolint:wrapcheck
}

// OpenCmd opens anything.
func OpenCmd(cmd ...string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnsupported, runtime.GOOS, cmd)
}

// OpenURL opens URL Links.
func OpenURL(url string) error {
	return OpenCmd(url)
}

// OpenLog opens Log Files.
func OpenLog(logFile string) error {
	return OpenCmd(logFile)
}

// OpenFile open Config Files.
func OpenFile(filePath string) error {
	return OpenCmd(filePath)
}
