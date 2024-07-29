package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/logo/notifiarr.png"

// HasGUI returns true on macOS if USEGUI env var is true.
func HasGUI() bool {
	return hasGUI
}

func Toast(msg string, vars ...interface{}) error {
	if !hasGUI {
		return nil
	}

	// This finds terminal-notifier inside this app or in your PATH.
	app, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find application: %w", err)
	} else if app = filepath.Dir(app); filepath.Base(app) == "MacOS" {
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

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running cmd: %w", err)
	}

	return nil
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
	return OpenCmd("-b", "com.apple.Console", logFile)
}

// OpenFile open Config Files.
func OpenFile(filePath string) error {
	return OpenCmd("-t", filePath)
}

func HasStartupLink() (string, bool) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}

	path := filepath.Join(dir, "Library", "LaunchAgents", "io.golift.notifiarr.plist")
	if _, err := os.Stat(path); err != nil {
		return "", false
	}

	return path, true
}

func DeleteStartupLink() (string, error) {
	link, has := HasStartupLink()
	if !has {
		return "", nil
	}

	if err := os.Remove(link); err != nil {
		return "", fmt.Errorf("removing launch agent: %w", err)
	}

	return link, nil
}

func CreateStartupLink() (bool, string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return false, "", fmt.Errorf("finding home dir: %w", err)
	}

	dir = filepath.Join(dir, "Library", "LaunchAgents")
	if err := os.MkdirAll(dir, mnd.Mode0755); err != nil {
		return false, "", fmt.Errorf("making launch agent path: %w", err)
	}

	path := filepath.Join(dir, "io.golift.notifiarr.plist")

	file, err := os.OpenFile(path, os.O_TRUNC|os.O_RDWR|os.O_CREATE, mnd.Mode0644)
	if err != nil {
		return false, "", fmt.Errorf("creating launch agent: %w", err)
	}
	defer file.Close()

	err = template.Must(template.New("launchAgent").
		Funcs(template.FuncMap{
			"exe": func() string {
				exe, _ := os.Executable()
				return exe
			},
		}).Parse(bindata.Plist)).
		Execute(file, nil)
	if err != nil {
		return false, "", fmt.Errorf("writing launch agent: %w", err)
	}

	loaded := false
	if err := StartCmd("launchctl", "list", "io.golift.notifiarr"); err == nil {
		loaded = true
	}

	if err := StartCmd("launchctl", "load", "-w", path); err != nil {
		return loaded, path, fmt.Errorf("loading launch agent: %w", err)
	}

	return loaded, path, nil
}
