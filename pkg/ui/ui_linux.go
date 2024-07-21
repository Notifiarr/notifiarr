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
	if env := os.Getenv("FILE_OPENER"); env != "" {
		return env
	} else if path, err := exec.LookPath("xdg-open"); err != nil {
		return path
	} else if path, err := exec.LookPath("gnome-open"); err != nil {
		return path
	} else if path, err := exec.LookPath("slopen"); err != nil {
		return path
	}

	return "xdg-open" // Is there a better default?
}

// HasGUI tries to determine if the app was invoked as a GUI app.
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
func StartCmd(c string, v ...string) error {
	cmd := exec.Command(c, v...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	return cmd.Run() //nolint:wrapcheck
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
