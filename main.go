package main

import (
	"log"
	"runtime/debug"

	"github.com/Notifiarr/notifiarr/pkg/client"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
)

func main() {
	ui.HideConsoleWindow()
	// setup log package in case we throw an error in main.go before logging is setup.
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("[ERROR] ")

	defer func() {
		if r := recover(); r != nil {
			ui.ShowConsoleWindow()
			log.Printf("Go Panic! %s\n%v\n%s", mnd.BugIssue, r, string(debug.Stack()))
		}
	}()

	if err := client.Start(); err != nil {
		_, _ = ui.Error(mnd.Title, err.Error())
		log.Fatal(err) // nolint:gocritic // defer does not need to run if we have an error.
	}
}
