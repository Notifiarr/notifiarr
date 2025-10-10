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

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/checkapp"
	"github.com/Notifiarr/notifiarr/pkg/client"
	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler"
	"golift.io/cnfg"
	"golift.io/goty"
	"golift.io/goty/gotydoc"
)

const (
	outputFileName = "notifiarrConfig.ts"
	localPrefix    = "github.com/Notifiarr/notifiarr/"
)

//nolint:funlen
func main() {
	vendorDir := os.Getenv("VENDOR_DIR")
	if vendorDir == "" {
		log.Fatal("env VENDOR_DIR is not set")
	}

	docs := gotydoc.New()
	goat := goty.NewGoty(&goty.Config{
		GlobalOverrides: goty.Override{
			KeepUnderscores: true,
		},
		Docs: docs,
		Overrides: goty.Overrides{
			cnfg.Duration{}:                 {Type: "string"},
			reflect.TypeOf(time.Weekday(0)): {Comment: "The day of the week."},
		},
	})

	// Parse the weekday enums and then parse the config struct.
	log.Println("==> parsing enums")
	goat.Enums([]goty.Enum{
		{Name: "Sunday", Value: scheduler.Sunday},
		{Name: "Monday", Value: scheduler.Monday},
		{Name: "Tuesday", Value: scheduler.Tuesday},
		{Name: "Wednesday", Value: scheduler.Wednesday},
		{Name: "Thursday", Value: scheduler.Thursday},
		{Name: "Friday", Value: scheduler.Friday},
		{Name: "Saturday", Value: scheduler.Saturday},
	})
	goat.Enums([]goty.Enum{
		{Name: "password", Value: configfile.AuthPassword},
		{Name: "header", Value: configfile.AuthHeader},
		{Name: "noauth", Value: configfile.AuthNone},
	})
	goat.Enums([]goty.Enum{
		{Name: "DeadCron", Value: scheduler.DeadCron},
		{Name: "Minutely", Value: scheduler.Minutely},
		{Name: "Hourly", Value: scheduler.Hourly},
		{Name: "Daily", Value: scheduler.Daily},
		{Name: "Weekly", Value: scheduler.Weekly},
		{Name: "Monthly", Value: scheduler.Monthly},
	})
	log.Println("==> parsing config structs")
	goat.Parse(
		client.Integrations{},
		client.Profile{},
		client.ProfilePost{},
		commands.Stats{},
		apps.ApiResponse{},
		checkapp.CheckAllOutput{},
		client.BrowseDir{},
	)

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
	var (
		vendorPkgs []string
		localPkgs  []string
	)

	for _, pkg := range pkgs {
		if strings.HasPrefix(pkg, localPrefix) {
			localPkgs = append(localPkgs, pkg)
		} else if strings.Contains(pkg, ".") {
			vendorPkgs = append(vendorPkgs, pkg)
		}
	}

	return vendorPkgs, localPkgs
}
