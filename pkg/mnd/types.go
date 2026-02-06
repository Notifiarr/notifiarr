package mnd

import (
	"context"
)

// Log is created here to avoid an import cycle between website and logs.
var Log Logger

// Logger is an interface for our logs package. We use this to avoid an import cycle.
type Logger interface {
	Trace(reqID string, v ...any) string
	Print(reqID string, v ...any)
	Printf(reqID string, msg string, v ...any)
	Error(reqID string, v ...any)
	Errorf(reqID string, msg string, v ...any)
	ErrorfNoShare(reqID string, msg string, v ...any)
	Debug(reqID string, v ...any)
	Debugf(reqID string, msg string, v ...any)
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
