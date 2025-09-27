package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/jxeng/shortcut"
	"github.com/ncruces/zenity"
	"golang.org/x/sys/windows"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "favicon.ico"

// HasGUI always returns true on Windows.
func HasGUI() bool {
	return true
}

func Toast(_ context.Context, msg string, v ...interface{}) error {
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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running cmd: %w", err)
	}

	return nil
}

// OpenCmd opens anything.
func OpenCmd(ctx context.Context, cmd ...string) error {
	return StartCmd(ctx, opener, append([]string{"/c", "start"}, cmd...)...)
}

// OpenURL opens URL Links.
func OpenURL(ctx context.Context, url string) error {
	return OpenCmd(ctx, strings.ReplaceAll(url, "&", "^&"))
}

// OpenLog opens Log Files.
func OpenLog(ctx context.Context, logFile string) error {
	return OpenCmd(ctx, "PowerShell", "Get-Content",
		"-Tail", "1000", "-Wait", "-Encoding", "utf8", "-Path", "'"+logFile+"'")
}

// OpenFile open Config Files.
func OpenFile(ctx context.Context, filePath string) error {
	return OpenCmd(ctx, "file://"+filePath)
}

const linkName = "Notifiarr.lnk"

func HasStartupLink() (string, bool) {
	path, err := windows.KnownFolderPath(windows.FOLDERID_Startup, 0)
	if err != nil {
		return "", false
	}

	if _, err = os.Stat(filepath.Join(path, linkName)); err != nil {
		return "", false
	}

	return filepath.Join(path, linkName), true
}

func CreateStartupLink(_ context.Context) (bool, string, error) {
	exe, err := os.Executable()
	if err != nil {
		return false, "", fmt.Errorf("finding executable: %w", err)
	}

	path, err := windows.KnownFolderPath(windows.FOLDERID_Startup, 0)
	if err != nil {
		return false, "", fmt.Errorf("getting startup folder: %w", err)
	}

	path = filepath.Join(path, linkName)
	loaded := false

	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(path) // Remove it so we can re-create it.
		loaded = true
	}

	err = shortcut.Create(shortcut.Shortcut{
		ShortcutPath:     path,
		Target:           exe,
		IconLocation:     GetPNG(),
		Description:      "Launches client for Notifiarr.com",
		WindowStyle:      "1",
		WorkingDirectory: filepath.Dir(exe),
	})
	if err != nil {
		return loaded, "", fmt.Errorf("creating startup shortcut: %w", err)
	}

	return loaded, path, nil
}

func DeleteStartupLink() (string, error) {
	link, has := HasStartupLink()
	if !has {
		return "", nil
	}

	if err := os.Remove(link); err != nil {
		return "", fmt.Errorf("removing shortcut: %w", err)
	}

	return link, nil
}
