// Package bindata provides file system assets to the running binary.
package bindata

import "embed"

var (
	//go:embed files
	Files embed.FS
	//go:embed templates
	Templates embed.FS
	//go:embed other/fortunes.txt
	Fortunes string
	//go:embed other/io.golift.notifiarr.plist
	Plist string
)
