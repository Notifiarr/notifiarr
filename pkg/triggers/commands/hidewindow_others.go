//go:build !windows

package commands

import (
	"os/exec"
)

func hideWindow(_ *exec.Cmd) {}
