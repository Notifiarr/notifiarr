//go:build !linux && !windows

package logs

/* The purpose of this code is to log stderr (application panics) to a log file. */

import (
	"os"
	"syscall"
)

//nolint:gochecknoglobals
var stderr = os.Stderr.Fd()

func redirectStderr(file *os.File) {
	// This works on darwin and freebsd, maybe others.
	_ = syscall.Dup2(int(file.Fd()), int(stderr))
}
