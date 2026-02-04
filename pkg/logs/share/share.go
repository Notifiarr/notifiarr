// Package share is here so we can keep website cruft out of the logs package.
package share

import (
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

type Website interface {
	SendData(req *website.Request)
}

var (
	// Config is setup by the configfile package.
	enabled bool
	locker  sync.RWMutex
)

func Enable() {
	locker.Lock()
	defer locker.Unlock()

	enabled = true
}

func Disable() {
	locker.Lock()
	defer locker.Unlock()

	enabled = false
}

// Match is what we send to the website.
type Match struct {
	File    string   `json:"file"`
	Matches []string `json:"matches"`
	Line    string   `json:"line"`
}

// Log sends an error message to the website.
func Log(reqID string, msg string) {
	locker.RLock()
	defer locker.RUnlock()

	if ci := data.Get("clientInfo"); ci == nil || !enabled {
		return
	}

	website.SendData(&website.Request{
		ReqID:      reqID,
		Payload:    &Match{File: "client_error_log", Line: msg, Matches: []string{"[ERROR]"}},
		Route:      website.LogLineRoute,
		Event:      website.EventFile,
		LogPayload: true,
	})
}
