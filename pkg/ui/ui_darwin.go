package ui

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/kardianos/osext"
)

// SystrayIcon is the icon in the menu bar.
const SystrayIcon = "files/macos.png"

var hasGUI = os.Getenv("USEGUI") == "true"

// HasGUI returns false on Linux, true on Windows and optional on macOS.
func HasGUI() bool {
	return hasGUI
}

// HideConsoleWindow doesn't work on maacOS.
func HideConsoleWindow() {}

// ShowConsoleWindow does nothing on OSes besides Windows.
func ShowConsoleWindow() {}

func Notify(msg string, v ...interface{}) error {
	if !hasGUI {
		return nil
	}

	app, err := exec.LookPath("terminal-notifier")
	if err != nil {
		if folder, err := osext.ExecutableFolder(); err == nil {
			app = filepath.Join(folder, app)
		}
	}

	if _, err := os.Stat(app); err != nil {
		return fmt.Errorf("cannot find terminal-notifier: %w", err)
	}

	err = StartCmd(app, "-title", mnd.Title, "-message", fmt.Sprintf(msg, v...), "-sender", "com.notifiarr.client")
	if err != nil {
		return fmt.Errorf("ui element failed: %w", err)
	}

	return nil
}

/*
// getPNG purposely returns an empty string when there is no verified file.
// This is used to give the toast notification an icon.
// Do not throw errors if the icon is missing, it'd nbd, just return empty "".
func getPNG() string {
	folder, err := osext.ExecutableFolder()
	if err != nil {
		return ""
	}

	pngPath := filepath.Join(folder, "notifiarr.png")
	if _, err := os.Stat(pngPath); err == nil {
		return pngPath // most code paths land here.
	}

	try := "/Applications/Notifiarr.app/Contents/MacOS/notifiarr.png"
	if _, err := os.Stat(try); err == nil {
		return try
	}

	data, err := bindata.Asset("files/favicon.png")
	if err != nil {
		return ""
	}

	if err := os.WriteFile(pngPath, data, mnd.Mode0600); err == nil {
		return pngPath
	}

	if err := os.WriteFile(try, data, mnd.Mode0600); err == nil {
		return try
	}

	if err := os.WriteFile("/tmp/notifiarr.png", data, mnd.Mode0600); err == nil {
		return "/tmp/notifiarr.png"
	}

	return ""
}
*/

// StartCmd starts a command.
func StartCmd(c string, v ...string) error {
	cmd := exec.Command(c, v...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard

	return cmd.Run()
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
