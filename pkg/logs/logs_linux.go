package logs

/* The purpose of this code is to log stderr (application panics) to a log file. */

import (
	"os"
	"syscall"
)

var stderr = os.Stderr.Fd() //nolint:gochecknoglobals

func redirectStderr(file *os.File) {
	_ = syscall.Dup3(int(file.Fd()), int(stderr), 0)
}
