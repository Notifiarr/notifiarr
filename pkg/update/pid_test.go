package update //nolint:testpackage

import (
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func startPauseProcess(t *testing.T) *exec.Cmd {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "sleep", "60")
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(t.Context(), "ping", "-n", "60", "127.0.0.1")
	}

	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	return cmd
}

func TestWaitForPidToExitInvalid(t *testing.T) {
	t.Parallel()

	start := time.Now()
	waitForPidToExit(-1)
	waitForPidToExit(0)
	require.Less(t, time.Since(start), time.Second)
}

func TestWaitForPidToExitAlreadyGone(t *testing.T) {
	t.Parallel()

	cmd := exec.CommandContext(t.Context(), "true")
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(t.Context(), "cmd", "/c", "exit", "0")
	}

	require.NoError(t, cmd.Run())

	start := time.Now()
	waitForPidToExit(cmd.Process.Pid)
	require.Less(t, time.Since(start), time.Second)
}

func TestWaitForPidToExitAfterKill(t *testing.T) {
	t.Parallel()

	cmd := startPauseProcess(t)
	done := make(chan struct{})

	go func() {
		waitForPidToExit(cmd.Process.Pid)
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("waitForPidToExit returned while the process was still running")
	case <-time.After(200 * time.Millisecond):
	}

	require.NoError(t, cmd.Process.Kill())
	_ = cmd.Wait()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("waitForPidToExit did not return after the process exited")
	}
}
