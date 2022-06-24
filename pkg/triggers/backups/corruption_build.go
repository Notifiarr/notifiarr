// https://pkg.go.dev/modernc.org/sqlite#hdr-Supported_platforms_and_architectures
//go:build (darwin && (amd64 || arm64)) || (freebsd && amd64) || (windows && amd64) || linux

package backups

// This driver does not work on all architectures.
// Missing platforms produce errors when working on sqlite databases.
import _ "modernc.org/sqlite" // database driver for sqlite3.
