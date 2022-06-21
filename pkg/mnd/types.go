package mnd

import "log"

// Logger is an interface for our logs package. We use this to avoid an import cycle.
type Logger interface {
	Print(v ...interface{})
	Printf(msg string, v ...interface{})
	Error(v ...interface{})
	Errorf(msg string, v ...interface{})
	ErrorfNoShare(msg string, v ...interface{})
	Debug(v ...interface{})
	Debugf(msg string, v ...interface{})
	GetInfoLog() *log.Logger
	DebugEnabled() bool
	CapturePanic()
}
