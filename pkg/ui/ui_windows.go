package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gen2brain/beeep"
	"github.com/gonutz/w32"
	"github.com/kardianos/osext"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "files/images/favicon.ico"

// HasGUI always returns true on Windows.
func HasGUI() bool {
	return true
}

func Notify(msg string, v ...interface{}) error {
	err := beeep.Notify(mnd.Title, fmt.Sprintf(msg, v...), getPNG())
	if err != nil {
		return fmt.Errorf("ui element failed: %w", err)
	}

	return nil
}

// getPNG purposely returns an empty string when there is no verified file.
// This is used to give the toast notification an icon.
// Do not throw errors if the icon is missing, it'd nbd, just return empty "".
func getPNG() string {
	folder, err := osext.ExecutableFolder()
	if err != nil {
		return ""
	}

	data, err := bindata.Asset("files/favicon.png")
	if err != nil {
		return ""
	}

	const (
		percent99  = 0.99
		percent101 = 1.01
	)

	minimumFileSize := int64(float64(len(data)) * percent99)
	maximumFileSize := int64(float64(len(data)) * percent101)
	pngPath := filepath.Join(folder, "notifiarr.png")

	f, err := os.Stat(pngPath)
	if err != nil || f.Size() < minimumFileSize || f.Size() > maximumFileSize {
		// File does not exist, or not within 1% of correct size. Overwrite it.
		if err := os.WriteFile(pngPath, data, mnd.Mode0600); err != nil {
			return ""
		}
	}

	// go log.Println("minmaxsize", minimumFileSize, maximumFileSize, f.Size(), len(data))
	return pngPath
}

// HideConsoleWindow makes the console window vanish on startup.
func HideConsoleWindow() {
	if console := w32.GetConsoleWindow(); console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, w32.SW_HIDE)
		}
	}
}

// ShowConsoleWindow does nothing on OSes besides Windows.
func ShowConsoleWindow() {
	if console := w32.GetConsoleWindow(); console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, w32.SW_SHOW)
		}
	}
}

// StartCmd starts a command.
func StartCmd(c string, v ...string) error {
	cmd := exec.Command(c, v...)
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
