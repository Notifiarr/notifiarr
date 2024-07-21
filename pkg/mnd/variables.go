package mnd

import (
	"errors"
	"os"
	"strings"

	"github.com/gorilla/schema"
	"github.com/hako/durafmt"
	"golift.io/version"
)

//nolint:gochecknoglobals
var (
	// IsSynology tells us if this we're running on a Synology.
	IsSynology = isSynology()
	// IsDocker tells us if this is a Docker container.
	IsDocker        = isDocker()
	IsUnstable      = strings.HasPrefix(version.Branch, "unstable")
	DurafmtUnits, _ = durafmt.DefaultUnitsCoder.Decode("year,week,day,hour,min,sec,ms:ms,µs:µs")
	DurafmtShort, _ = durafmt.DefaultUnitsCoder.Decode("y:y,w:w,d:d,h:h,m:m,s:s,ms:ms,µs:µs")
	// Set a Decoder instance as a package global, because it caches
	// meta-data about structs, and an instance can be shared safely.
	ConfigPostDecoder = schema.NewDecoder()
)

// ErrDisabledInstance is returned when a request for a disabled instance is performed.
var ErrDisabledInstance = errors.New("instance is administratively disabled")

func isSynology() bool {
	_, err := os.Stat(Synology)
	return err == nil
}

func isDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return os.Getpid() == 1 || err == nil
}
