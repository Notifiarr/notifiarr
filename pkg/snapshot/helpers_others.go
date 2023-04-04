//go:build !windows

package snapshot

import (
	"os/exec"
)

func sysCallSettings(_ *exec.Cmd) {}
