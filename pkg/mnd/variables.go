package mnd

import "os"

//nolint:gochecknoglobals
var (
	// IsSynology tells us if this we're running on a Synology.
	IsSynology bool
	// IsDocker tells us if this is our Docker container.
	IsDocker = os.Getenv(DockerV) == "true"
)

//nolint:gochecknoinits
func init() {
	_, err := os.Stat(Synology)
	IsSynology = err == nil
}
