package main

//go:generate go-bindata -pkg dnclient  -o dnclient/bindata.go init/windows/application.ico

import (
	"log"

	"github.com/Go-Lift-TV/discordnotifier-client/dnclient"
)

func main() {
	if err := dnclient.Start(); err != nil {
		log.Fatal("[ERROR]", err)
	}
}
