//go:build windows || darwin || linux

package ui

import (
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gen2brain/dlgs"
	"golift.io/version"
)

func title(suffix ...string) string {
	return mnd.Title + " " + version.Version + "-" + version.Revision + " " + strings.Join(suffix, " ")
}

func now() string {
	return "\nNow: " + time.Now().Format("Mon Jan 2, 2006 @ 15:04:05 MST")
}

// Warning wraps dlgs.Warning.
func Warning(msg string) (bool, error) {
	if !HasGUI() {
		return true, nil
	}

	return dlgs.Warning(title(), msg+now()) //nolint:wrapcheck
}

// Error wraps dlgs.Error.
func Error(msg string) (bool, error) {
	if !HasGUI() {
		return true, nil
	}

	return dlgs.Error(title("ERROR"), msg+now()) //nolint:wrapcheck
}

// Info wraps dlgs.Info.
func Info(msg string) (bool, error) {
	if !HasGUI() {
		return true, nil
	}

	return dlgs.Info(title(), msg+now()) //nolint:wrapcheck
}

// Entry wraps dlgs.Entry.
func Entry(msg, val string) (string, bool, error) {
	if !HasGUI() {
		return val, true, nil
	}

	return dlgs.Entry(title(), msg+now(), val) //nolint:wrapcheck
}

// Question wraps dlgs.Question.
func Question(text string, defaultCancel bool) (bool, error) {
	if !HasGUI() {
		return true, nil
	}

	return dlgs.Question(title(), text+now(), defaultCancel) //nolint:wrapcheck
}
