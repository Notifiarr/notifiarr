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

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/logo/notifiarr.png"

// ErrUnsupported is just an error.
var ErrUnsupported = errors.New("unsupported OS")

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

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running cmd: %w", err)
	}

	return nil
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

func HasStartupLink() (string, bool) {
	return "", false
}

func DeleteStartupLink() (string, error) {
	return "", ErrUnsupported
}

func CreateStartupLink() (bool, string, error) {
	return false, "", ErrUnsupported
}
