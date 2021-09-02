package ui

import (
	"fmt"
	"io/ioutil"
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
const SystrayIcon = "files/favicon.ico"

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

	const minimumFileSize = 100 // arbitrary

	pngPath := filepath.Join(folder, "notifiarr.png")
	if f, err := os.Stat(pngPath); err == nil && f.Size() > minimumFileSize {
		return pngPath // most code paths land here.
	} else if !os.IsNotExist(err) || (f != nil && f.Size() < minimumFileSize) {
		return ""
	}

	data, err := bindata.Asset("files/favicon.png")
	if err != nil {
		return ""
	}

	if err := os.WriteFile(pngPath, data, mnd.Mode0600); err != nil {
		return ""
	}

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
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard
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
	return OpenCmd("PowerShell", "Get-Content", "-Tail", "1000", "-Wait", "-Encoding", "utf8", "-Path", logFile)
}

// OpenFile open Config Files.
func OpenFile(filePath string) error {
	return OpenCmd("file://" + filePath)
}
