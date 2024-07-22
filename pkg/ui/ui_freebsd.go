package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gen2brain/beeep"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/favicon.png"

//nolint:gochecknoglobals
var (
	hasGUI = os.Getenv("USEGUI") == "true"
	opener = getOpener()
)

func getOpener() string {
	if path := os.Getenv("FILE_OPENER"); path != "" {
		return path
	} else if path, _ := exec.LookPath("xdg-open"); path != "" {
		return path
	} else if path, _ = exec.LookPath("gnome-open"); path != "" {
		return path
	} else if path, _ = exec.LookPath("slopen"); path != "" {
		return path
	}

	return "xdg-open" // Is there a better default?
}

// HasGUI returns true on Linux if USEGUI env var is true.
func HasGUI() bool {
	return hasGUI
}

func Toast(msg string, v ...interface{}) error {
	err := beeep.Notify(mnd.Title, fmt.Sprintf(msg, v...), GetPNG())
	if err != nil {
		return fmt.Errorf("ui element failed: %w", err)
	}

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
