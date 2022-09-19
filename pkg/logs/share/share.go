// Package share is here so we can keep website cruft out of the logs package.
package share

import (
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/triggers/filewatch"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

type Website interface {
	SendData(data *website.Request)
}

//nolint:gochecknoglobals
var (
	// Config is setup by the configfile package.
	config Website
	locker sync.RWMutex
)

func Setup(website Website) {
	locker.Lock()
	defer locker.Unlock()

	config = website
}

func StopLogs() {
	locker.Lock()
	defer locker.Unlock()

	config = nil
}

// Log sends an error message to the website.
func Log(msg string) {
	locker.RLock()
	defer locker.RUnlock()

	if config == nil || !website.HaveClientInfo() {
		return
	}

	config.SendData(&website.Request{
		Payload:    &filewatch.Match{File: "client_error_log", Line: msg, Matches: []string{"[ERROR]"}},
		Route:      website.LogLineRoute,
		Event:      website.EventFile,
		LogPayload: true,
	})
}
