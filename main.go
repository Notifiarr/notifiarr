package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/dnclient"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// setup log package in case we throw an error for main.go before logging is setup.
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("[ERROR] ")

	// Set time zone based on TZ env variable.
	if err := setTimeZone(os.Getenv("TZ")); err != nil {
		log.Print(err)
	}

	return dnclient.Start()
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
