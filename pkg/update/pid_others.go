//go:build !windows

package update

import (
	"os"
	"syscall"
	"time"
)

const parentWaitPollTime = 500 * time.Millisecond

// waitForPidToExit blocks until pid exits, or parentWaitTimeout elapses.
// Unix FindProcess always succeeds, so existence is probed with Signal(0).
func waitForPidToExit(pid int) {
	if pid <= 0 {
		return
	}

	deadline := time.Now().Add(parentWaitTimeout)

	for time.Now().Before(deadline) {
		if !isPidRunning(pid) {
			return
		}

		time.Sleep(parentWaitPollTime)
	}
}

func isPidRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	return process.Signal(syscall.Signal(0)) == nil
}
