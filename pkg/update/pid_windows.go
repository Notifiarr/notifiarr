package update

import (
	"time"

	"golang.org/x/sys/windows"
)

// waitForPidToExit blocks until pid exits, or parentWaitTimeout elapses.
// OpenProcess(SYNCHRONIZE) plus WaitForSingleObject is signaled when the
// process actually exits, even if the kernel object is still around.
func waitForPidToExit(pid int) {
	if pid <= 0 {
		return
	}

	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return
	}

	defer func() { _ = windows.CloseHandle(handle) }()

	_, _ = windows.WaitForSingleObject(handle, uint32(parentWaitTimeout/time.Millisecond))
}
