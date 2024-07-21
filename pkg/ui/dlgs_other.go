//go:build !windows && !darwin && !linux

package ui

// Warning wraps dlgs.Warning.
func Warning(_ string) (bool, error) {
	return true, nil
}

// Error wraps dlgs.Error.
func Error(_ string) (bool, error) {
	return true, nil
}

// Info wraps dlgs.Info.
func Info(_ string) (bool, error) {
	return true, nil
}

// Entry wraps dlgs.Entry.
func Entry(_, val string) (string, bool, error) {
	return val, false, nil
}

// Question wraps dlgs.Question.
func Question(_ string, _ bool) (bool, error) {
	return true, nil
}
