//go:build !windows && !darwin && !linux

package ui

// Warning wraps dlgs.Warning.
func Warning(_ string) {}

// Error wraps dlgs.Error.
func Error(_ string) {}

// Info wraps dlgs.Info.
func Info(_ string) {}

// Entry wraps dlgs.Entry.
func Entry(_, val string) (string, bool, error) {
	return val, false, nil
}

// Question wraps dlgs.Question.
func Question(_ string, _ bool) (bool, error) {
	return true, nil
}
