package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/ncruces/zenity"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "notifiarr.png"

// HasGUI returns true on Linux if USEGUI env var is true.
func HasGUI() bool {
	return hasGUI
}

func Toast(_ context.Context, msg string, v ...any) error {
	if !hasGUI {
		return nil
	}

	err := zenity.Notify(fmt.Sprintf(msg, v...), zenity.Title(mnd.Title), zenity.Icon(GetPNG()))
	if err != nil {
		return fmt.Errorf("ui element failed: %w", err)
	}

	return nil
}

// StartCmd starts a command.
func StartCmd(ctx context.Context, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running cmd: %w", err)
	}

	return nil
}

// OpenCmd opens anything.
func OpenCmd(ctx context.Context, cmd ...string) error {
	return StartCmd(ctx, opener, cmd...)
}

// OpenURL opens URL Links.
func OpenURL(ctx context.Context, url string) error {
	return OpenCmd(ctx, url)
}

// OpenLog opens Log Files.
func OpenLog(ctx context.Context, logFile string) error {
	return OpenCmd(ctx, logFile)
}

// OpenFile open Config Files.
func OpenFile(ctx context.Context, filePath string) error {
	return OpenCmd(ctx, filePath)
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

func CreateStartupLink(_ context.Context) (bool, string, error) {
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

	err = os.Symlink("/usr/share/applications/notifiarr.desktop", path)
	if err != nil {
		return loaded, "", fmt.Errorf("linking autostart: %w", err)
	}

	return loaded, path, nil
}
