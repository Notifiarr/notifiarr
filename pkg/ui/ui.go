package ui

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/kardianos/osext"
)

// pngPathCache caches the path to the application icon.
// Do not use this variable directly. Call GetPNG()
var pngPathCache = "" //nolint:gochecknoglobals

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

	file, err := os.Stat(pngPath)
	if err != nil || file.Size() < minimumFileSize || file.Size() > maximumFileSize {
		// File does not exist, or not within 1% of correct size. Overwrite it.
		if err := os.WriteFile(pngPath, data, mnd.Mode0644); err != nil {
			return ""
		}
	}

	pngPathCache = pngPath // Save it for next time.

	// TODO: comment this debug log, and remove the TODO.
	go log.Println("minmaxsize", minimumFileSize, maximumFileSize, file.Size(), len(data))

	return pngPath
}
