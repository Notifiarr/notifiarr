//go:build (darwin && (amd64 || arm64)) || (freebsd && amd64) || (netbsd && amd64) || windows || linux

package notifiarr

// This driver does not work on all architectures.
// Missing platforms produce errors when working on sqlite databases.
import _ "modernc.org/sqlite" // database driver for sqlite3.
