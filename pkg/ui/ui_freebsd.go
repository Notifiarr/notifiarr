package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gen2brain/beeep"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/logo/notifiarr.png"

// HasGUI returns true on FreeBSD if USEGUI env var is true.
func HasGUI() bool {
	return hasGUI
}

// Toast does not work properly on FreeBSD because we cross compile it without dbus. :(
func Toast(msg string, v ...interface{}) error {
	if !hasGUI {
		return nil
	}

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
	return OpenCmd(logFile)
}

// OpenFile open Config Files.
func OpenFile(filePath string) error {
	return OpenCmd(filePath)
}

func HasStartupLink() (string, bool) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}

	path := filepath.Join(dir, ".config", "autostart", "notifiarr.desktop")
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
		return "", fmt.Errorf("unlinking autostart: %w", err)
	}

	return link, nil
}

func CreateStartupLink() (bool, string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return false, "", fmt.Errorf("finding home dir: %w", err)
	}

	dir = filepath.Join(dir, ".config", "autostart")
	if err := os.MkdirAll(dir, mnd.Mode0755); err != nil {
		return false, "", fmt.Errorf("making autostart: %w", err)
	}

	path := filepath.Join(dir, "notifiarr.desktop")
	loaded := false

	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(path) // Remove it so we can re-create it.
		loaded = true
	}

	err = os.Symlink("/usr/local/share/applications/notifiarr.desktop", path)
	if err != nil {
		return loaded, "", fmt.Errorf("linking autostart: %w", err)
	}

	return loaded, path, nil
}
