package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/client"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/ui"
)

func main() {
	setup()

	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
}

func setup() {
	ui.HideConsoleWindow()
	// setup log package in case we throw an error for main.go before logging is setup.
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("[ERROR] ")

	// Set time zone based on TZ env variable.
	if err := setTimeZone(os.Getenv("TZ")); err != nil {
		log.Print(err)
	}
}

func setTimeZone(tz string) (err error) {
	if tz == "" {
		return nil
	}

	if time.Local, err = time.LoadLocation(tz); err != nil {
		return fmt.Errorf("loading TZ location '%s': %w", tz, err)
	}

	return nil
}
