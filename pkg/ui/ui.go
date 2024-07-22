package ui

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/kardianos/osext"
)

//nolint:gochecknoglobals
var (
	// pngPathCache caches the path to the application icon.
	// Do not use this variable directly. Call GetPNG()
	pngPathCache = ""
	// hasGUI is used on Linux and Darwin.
	hasGUI = os.Getenv("USEGUI") == "true" //nolint:unused,nolintlint
	opener = getOpener()
)

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

	folder, err := os.MkdirTemp("", "notifiarr")
	if err != nil {
		// No temp dir, try to put it next to the executable.
		if folder, err = osext.ExecutableFolder(); err != nil {
			return ""
		}
	}

	data, err := bindata.Asset("files/images/favicon.png")
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

	file, err := os.Stat(pngPath)
	if err != nil || file.Size() < minimumFileSize || file.Size() > maximumFileSize {
		// File does not exist, or not within 1% of correct size. Overwrite it.
		if err = os.WriteFile(pngPath, data, mnd.Mode0644); err != nil {
			return ""
		}
	}

	pngPathCache = pngPath // Save it for next time.

	// go log.Println("minmaxsize", minimumFileSize, maximumFileSize, file.Size(), len(data))

	return pngPath
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
