// Package share is here so we can keep website cruft out of the logs package.
package share

import "github.com/Notifiarr/notifiarr/pkg/notifiarr"

type Website interface {
	QueueData(data *notifiarr.SendRequest)
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

	config.QueueData(&notifiarr.SendRequest{
		Payload:    &notifiarr.Match{File: "client_error_log", Line: msg},
		Route:      notifiarr.LogLineRoute,
		Event:      notifiarr.EventFile,
		LogPayload: true,
	})
}
