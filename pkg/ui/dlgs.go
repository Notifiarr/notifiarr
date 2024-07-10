//go:build windows || darwin

package ui

import (
	"time"

	"github.com/gen2brain/dlgs"
)

func now() string {
	return "\nNow: " + time.Now().Format("Mon Jan 2, 2006 @ 15:04:05 MST")
}

// Warning wraps dlgs.Warning.
func Warning(title, msg string) (bool, error) {
	if !HasGUI() {
		return true, nil
	}

	return dlgs.Warning(title, msg+now()) //nolint:wrapcheck
}

// Error wraps dlgs.Error.
func Error(title, msg string) (bool, error) {
	if !HasGUI() {
		return true, nil
	}

	return dlgs.Error(title, msg+now()) //nolint:wrapcheck
}

// Info wraps dlgs.Info.
func Info(title, msg string) (bool, error) {
	if !HasGUI() {
		return true, nil
	}

	return dlgs.Info(title, msg+now()) //nolint:wrapcheck
}

// Entry wraps dlgs.Entry.
func Entry(title, msg, val string) (string, bool, error) {
	if !HasGUI() {
		return val, true, nil
	}

	return dlgs.Entry(title, msg+now(), val) //nolint:wrapcheck
}

// Question wraps dlgs.Question.
func Question(title, text string, defaultCancel bool) (bool, error) {
	if !HasGUI() {
		return true, nil
	}

	return dlgs.Question(title, text+now(), defaultCancel) //nolint:wrapcheck
}
