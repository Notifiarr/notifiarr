// Package share is here so we can keep website cruft out of the logs package.
package share

import (
	"github.com/Notifiarr/notifiarr/pkg/triggers/filewatch"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

type Website interface {
	SendData(data *website.Request)
	HaveClientInfo() bool
}

// Config is setup by the configfile package.
var config Website //nolint:gochecknoglobals

func Setup(website Website) {
	config = website
}

// Log sends an error message to the website.
func Log(msg string) {
	if config == nil || !config.HaveClientInfo() {
		return
	}

	config.SendData(&website.Request{
		Payload:    &filewatch.Match{File: "client_error_log", Line: msg, Matches: []string{"[ERROR]"}},
		Route:      website.LogLineRoute,
		Event:      website.EventFile,
		LogPayload: true,
	})
}
