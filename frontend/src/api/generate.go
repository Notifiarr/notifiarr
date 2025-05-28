// Package main is used to generate the backend data model as typescript interfaces.
//
//go:generate sh generate.sh
package main

import (
	"log"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/client"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"golift.io/cnfg"
	"golift.io/goty"
	"golift.io/goty/gotydoc"
)

const (
	outputFileName = "notifiarrConfig.ts"
	localPrefix    = "github.com/Notifiarr/notifiarr/"
)

func main() {
	vendorDir := os.Getenv("VENDOR_DIR")
	if vendorDir == "" {
		log.Fatal("env VENDOR_DIR is not set")
	}

	docs := gotydoc.New()
	goat := goty.NewGoty(&goty.Config{
		Docs: docs,
		Overrides: goty.Overrides{
			cnfg.Duration{}:                 {Type: "string"},
			reflect.TypeOf(time.Weekday(0)): {Comment: "The day of the week."},
		},
	})

	// Parse the weekday enums and then parse the config struct.
	log.Println("==> parsing enums")
	goat.Enums([]goty.Enum{
		{Name: "Sunday", Value: time.Sunday},
		{Name: "Monday", Value: time.Monday},
		{Name: "Tuesday", Value: time.Tuesday},
		{Name: "Wednesday", Value: time.Wednesday},
		{Name: "Thursday", Value: time.Thursday},
		{Name: "Friday", Value: time.Friday},
		{Name: "Saturday", Value: time.Saturday},
	})
	goat.Enums([]goty.Enum{
		{Name: "password", Value: configfile.AuthPassword},
		{Name: "header", Value: configfile.AuthHeader},
		{Name: "noauth", Value: configfile.AuthNone},
	})
	log.Println("==> parsing config structs")
	goat.Parse(client.Profile{}, client.ProfilePost{})

	log.Println("==> splitting packages")
	vendorPkgs, localPkgs := splitPkgs(goat.Pkgs())

	log.Printf("==> adding %d vendor packages", len(vendorPkgs))
	docs.AddMust(vendorDir, vendorPkgs...) // `go mod vendor`

	log.Printf("==> adding %d local packages", len(localPkgs))
	for _, pkg := range localPkgs {
		dir := strings.TrimPrefix(pkg, localPrefix)
		docs.AddPkgMust(path.Join("../../../", dir), pkg) // `git clone`
	}

	log.Println("==> writing output file")
	if err := goat.Write(outputFileName, true); err != nil {
		log.Fatal(err)
	}
}

// All the notifiarr packages are not in the vendor folder.
// This splits them out so we can add the vendor docs from the vendor folder,
// and the local docs from the local git checkout.
func splitPkgs(pkgs []string) ([]string, []string) {

	vendorPkgs := []string{}
	localPkgs := []string{}

	for _, pkg := range pkgs {
		if strings.HasPrefix(pkg, localPrefix) {
			localPkgs = append(localPkgs, pkg)
		} else {
			vendorPkgs = append(vendorPkgs, pkg)
		}
	}

	return vendorPkgs, localPkgs
}
