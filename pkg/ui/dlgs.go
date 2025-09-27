//go:build windows || darwin || linux

package ui

import (
	"errors"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/ncruces/zenity"
	"golift.io/version"
)

func title(suffix ...string) string {
	return mnd.Title + " " + version.Version + "-" + version.Revision + " " + strings.Join(suffix, " ")
}

func now() string {
	return "\nNow: " + time.Now().Format("Mon Jan 2, 2006 @ 15:04:05 MST")
}

// Warning wraps dlgs.Warning.
func Warning(msg string) {
	if !HasGUI() {
		return
	}

	_ = zenity.Warning(msg+now(), zenity.Title(title()))
}

// Error wraps zenity.Error.
func Error(msg string) {
	if !HasGUI() {
		return
	}

	_ = zenity.Error(msg+now(), zenity.Title(title("ERROR")))
}

// Info wraps zenity.Info.
func Info(msg string) {
	if !HasGUI() {
		return
	}

	_ = zenity.Info(msg+now(), zenity.Title(title()))
}

// Entry wraps zenity.Entry. Returns the entry text and true if ok is clicked.
func Entry(msg, val string) (string, bool, error) {
	if !HasGUI() {
		return val, true, nil
	}

	entry, err := zenity.Entry(msg+now(), zenity.Title(title()), zenity.EntryText(val))

	return entry, !errors.Is(err, zenity.ErrCanceled), err //nolint:wrapcheck
}

// Question wraps zenity.Question. Returns true if yes or ok is clicked.
func Question(text string, defaultCancel bool) (bool, error) {
	if !HasGUI() {
		return true, nil
	}

	opts := []zenity.Option{zenity.Title(title())}
	if defaultCancel {
		opts = append(opts, zenity.DefaultCancel())
	}

	err := zenity.Question(text+now(), opts...)
	return !errors.Is(err, zenity.ErrCanceled), err //nolint:wrapcheck
}
