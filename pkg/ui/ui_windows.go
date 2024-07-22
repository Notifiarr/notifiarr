package ui

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gen2brain/beeep"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/favicon.ico"

// HasGUI always returns true on Windows.
func HasGUI() bool {
	return true
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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return cmd.Start() //nolint:wrapcheck
}

// OpenCmd opens anything.
func OpenCmd(cmd ...string) error {
	return StartCmd("cmd", append([]string{"/c", "start"}, cmd...)...)
}

// OpenURL opens URL Links.
func OpenURL(url string) error {
	return OpenCmd(strings.ReplaceAll(url, "&", "^&"))
}

// OpenLog opens Log Files.
func OpenLog(logFile string) error {
	return OpenCmd("PowerShell", "Get-Content", "-Tail", "1000", "-Wait", "-Encoding", "utf8", "-Path", "'"+logFile+"'")
}

// OpenFile open Config Files.
func OpenFile(filePath string) error {
	return OpenCmd("file://" + filePath)
}
