package ui

import (
	"io"
	"os/exec"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/favicon.png"

// HasGUI returns true on FreeBSD if USEGUI env var is true.
func HasGUI() bool {
	return hasGUI
}

// Toast does not work on FreeBSD. :(
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
	return StartCmd(opener, cmd...)
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
