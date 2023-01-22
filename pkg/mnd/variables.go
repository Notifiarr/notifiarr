package mnd

import (
	"math/rand"
	"os"
	"time"
)

//nolint:gochecknoglobals
var (
	// IsSynology tells us if this we're running on a Synology.
	IsSynology bool
	// IsDocker tells us if this is our Docker container.
	IsDocker = os.Getpid() == 1
)

//nolint:gochecknoinits
func init() {
	// initialize global pseudo random generator that gets used in a few places.
	rand.Seed(time.Now().UnixNano())

	_, err := os.Stat(Synology)
	IsSynology = err == nil
}
