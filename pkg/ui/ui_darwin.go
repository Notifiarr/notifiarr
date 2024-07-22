package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/kardianos/osext"
)

// SystrayIcon is the icon in the menu bar.
const SystrayIcon = "files/images/macos.png"

//nolint:gochecknoglobals
var hasGUI = os.Getenv("USEGUI") == "true"

// HasGUI returns true on macOS if USEGUI env var is true.
func HasGUI() bool {
	return hasGUI
}

func Toast(msg string, vars ...interface{}) error {
	if !hasGUI {
		return nil
	}

	// This finds terminal-notifier inside this app or in your PATH.
	app, err := osext.ExecutableFolder()
	if err != nil {
		return fmt.Errorf("cannot find application running directory: %w", err)
	} else if filepath.Base(app) == "MacOS" {
		app = filepath.Join(filepath.Dir(app), "Resources", "terminal-notifier.app", "Contents", "MacOS", "terminal-notifier")
	} else if app, err = exec.LookPath("terminal-notifier"); err != nil {
		list, _ := os.ReadDir(filepath.Dir(app))
		return fmt.Errorf("cannot locate terminal-notifier: %w, app folder: %s, ../: %s", err, app, list)
	}

	err = StartCmd(app, "-title", mnd.Title, "-message", fmt.Sprintf(msg, vars...), "-sender", "io.golift.notifiarr")
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
	return StartCmd("open", cmd...)
}

// OpenURL opens URL Links.
func OpenURL(url string) error {
	return OpenCmd(url)
}

// OpenLog opens Log Files.
func OpenLog(logFile string) error {
	return OpenCmd("-b", "com.apple.Console", logFile)
}

// OpenFile open Config Files.
func OpenFile(filePath string) error {
	return OpenCmd("-t", filePath)
}
