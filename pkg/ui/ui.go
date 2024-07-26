package ui

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

//nolint:gochecknoglobals
var (
	asset = bindata.Files.ReadFile
	// pngPathCache caches the path to the application icon.
	// Do not use this variable directly. Call GetPNG()
	pngPathCache = ""
	// hasGUI is used on Linux and Darwin.
	hasGUI = os.Getenv("USEGUI") == "true" //nolint:unused,nolintlint
	opener = getOpener()
)

// ToastIcon is the icon in the menu bar.
const ToastIcon = "files/images/logo/notifiarr.png"

// GetPNG purposely returns an empty string when there is no verified file.
// This is used to give the toast notification an icon.
// Do not throw errors if the icon is missing, it's nbd, just return empty "".
func GetPNG() string {
	if pngPathCache != "" {
		return pngPathCache
	} else if pngPathCache == "-" { // Used as signal that we can't write a file.
		return ""
	}

	pngPathCache = "-" // We only do this once.

	if path := getPNGnotWindows(); path != "" {
		pngPathCache = path
		return pngPathCache
	}

	folder, err := os.MkdirTemp("", "notifiarr")
	if err != nil {
		// No temp dir, try to put it next to the executable.
		if folder, err = os.Executable(); err != nil {
			return ""
		}

		folder = filepath.Dir(folder)
	}

	data, err := asset(ToastIcon)
	if err != nil {
		return ""
	}

	pngPath := filepath.Join(folder, "notifiarr.png")
	if _, err := os.Stat(pngPath); err != nil {
		if err = os.WriteFile(pngPath, data, mnd.Mode0644); err != nil {
			return ""
		}
	}

	pngPathCache = pngPath // Save it for next time.

	// go log.Println("minmaxsize", minimumFileSize, maximumFileSize, file.Size(), len(data))

	return pngPath
}

func getPNGnotWindows() string {
	if mnd.IsWindows {
		return ""
	}

	for _, path := range []string{
		getMacOSResources("Images"),
		"/usr/share/doc/notifiarr/",
		"/usr/local/share/doc/notifiarr/",
		"/opt/homebrew/share/doc/notifiarr/",
	} {
		if path == "" {
			continue
		}

		if _, err := os.Stat(filepath.Join(path, "notifiarr.png")); err == nil {
			return path
		}
	}

	return ""
}

// This returns the absolute path to the mac app Resources folder. Or an empty string.
func getMacOSResources(subFolder string) string {
	if !mnd.IsDarwin {
		return ""
	}

	output, err := os.Executable()
	if err != nil {
		return ""
	} else if output = filepath.Dir(output); filepath.Base(output) == "MacOS" {
		return filepath.Join(filepath.Dir(output), "Resources", subFolder)
	}

	return ""
}

// getOpener returns the app that can open a file or url in a GUI.
func getOpener() string {
	if mnd.IsWindows {
		return "cmd"
	} else if mnd.IsDarwin {
		return "open"
	}

	if path := os.Getenv("FILE_OPENER"); path != "" {
		return path
	} else if path, _ := exec.LookPath("xdg-open"); path != "" {
		return path
	} else if path, _ = exec.LookPath("gnome-open"); path != "" {
		return path
	} else if path, _ = exec.LookPath("slopen"); path != "" {
		return path
	}

	return "xdg-open" // Is there a better default?
}
