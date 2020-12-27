package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/dnclient"
)

func main() {
	// setup log package in case we throw an error for main.go before logging is setup.
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("[ERROR] ")

	// Set time zone based on TZ env variable.
	if err := setTimeZone(os.Getenv("TZ")); err != nil {
		log.Print(err) // do not exit
	}

	if err := dnclient.Start(); err != nil {
		log.Fatal(err)
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
