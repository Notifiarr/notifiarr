// +build !windows

package snapshot

import (
	"os/exec"
)

func sysCallSettings(cmd *exec.Cmd) {}
