package main

//go:generate go-bindata -pkg bindata -modtime 1587356420 -o bindata/bindata.go init/windows/application.ico

import (
	"log"
	"os"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/dnclient"
)

// Keep it simple.
func main() {
	// Set time zone based on TZ env variable.
	setTimeZone(os.Getenv("TZ"))

	if err := dnclient.Start(); err != nil {
		log.Fatal("[ERROR]", err)
	}
}

func setTimeZone(tz string) {
	if tz == "" {
		return
	}

	var err error

	if time.Local, err = time.LoadLocation(tz); err != nil {
		log.Printf("[ERROR] Loading TZ Location '%s': %v", tz, err)
	}
}
