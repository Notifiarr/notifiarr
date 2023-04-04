package mnd

import (
	"fmt"
	"os"
)

//nolint:gochecknoglobals
var (
	// IsSynology tells us if this we're running on a Synology.
	IsSynology bool
	// IsDocker tells us if this is our Docker container.
	IsDocker = os.Getpid() == 1
)

// ErrDisabledInstance is returned when a request for a disabled instance is performed.
var ErrDisabledInstance = fmt.Errorf("instance is administratively disabled")

//nolint:gochecknoinits
func init() {
	_, err := os.Stat(Synology)
	IsSynology = err == nil
}
