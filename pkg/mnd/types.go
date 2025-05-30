package mnd

import (
	"context"
	"log"
)

var Log Lagger

// Logger is an interface for our logs package. We use this to avoid an import cycle.
type Lagger interface {
	Print(v ...interface{})
	Printf(msg string, v ...interface{})
	Error(v ...interface{})
	Errorf(msg string, v ...interface{})
	ErrorfNoShare(msg string, v ...interface{})
	Debug(v ...interface{})
	Debugf(msg string, v ...interface{})
	NoUploads() bool
	GetInfoLog() *log.Logger
	GetLogFiles() map[string]string
	DebugEnabled() bool
	CapturePanic()
}

// InstancePinger is an interface for pinging a server instance.
// Used between apps and client.
type InstancePinger interface {
	PingContext(ctx context.Context) error
	Enabled
}

type Enabled interface {
	Enabled() bool
}
