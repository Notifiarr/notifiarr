package mnd

import (
	"context"
)

// Log is created here to avoid an import cycle between website and logs.
var Log Logger

// Logger is an interface for our logs package. We use this to avoid an import cycle.
type Logger interface {
	Trace(id string, v ...any) string
	Print(v ...any)
	Printf(msg string, v ...any)
	Error(v ...any)
	Errorf(msg string, v ...any)
	ErrorfNoShare(msg string, v ...any)
	Debug(v ...any)
	Debugf(msg string, v ...any)
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
