// Package bindata provides file system assets to the running binary.
package bindata

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
)

//nolint:gochecknoglobals
var (
	//go:embed other/io.golift.notifiarr.plist
	Plist string
	//go:embed other/fortunes.txt.gz
	fortunes []byte
	Fortunes = decompress(fortunes)
)

func decompress(data []byte) string {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return ""
	}

	data, err = io.ReadAll(gz)
	if err != nil {
		return ""
	}

	return string(data)
}
