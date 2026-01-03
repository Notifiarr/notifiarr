//go:build !windows && !darwin && !linux && !freebsd

/* Not sure what OS this will get. */

package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

// SystrayIcon is the icon in the system tray or task bar.
const SystrayIcon = "notifiarr.png"

// ErrUnsupported is just an error.
var ErrUnsupported = errors.New("unsupported OS")

// HasGUI returns false on this gui-unsupported OS.
func HasGUI() bool {
	return false
}

// Toast does not do anything on this OS.
func Toast(_ context.Context, _ string, _ ...any) error {
	return nil
}

// StartCmd starts a command.
func StartCmd(ctx context.Context, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running cmd: %w", err)
	}

	return nil
}

// OpenCmd opens anything.
func OpenCmd(ctx context.Context, cmd ...string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnsupported, runtime.GOOS, cmd)
}

// OpenURL opens URL Links.
func OpenURL(ctx context.Context, url string) error {
	return OpenCmd(ctx, url)
}

// OpenLog opens Log Files.
func OpenLog(ctx context.Context, logFile string) error {
	return OpenCmd(ctx, logFile)
}

// OpenFile open Config Files.
func OpenFile(ctx context.Context, filePath string) error {
	return OpenCmd(ctx, filePath)
}

func HasStartupLink() (string, bool) {
	return "", false
}

func DeleteStartupLink() (string, error) {
	return "", ErrUnsupported
}

func CreateStartupLink(_ context.Context) (bool, string, error) {
	return false, "", ErrUnsupported
}
