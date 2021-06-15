package snapshot

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const SynologyConf = "/etc/synoinfo.conf"

type Synology struct {
	Build   string `json:"last_admin_login_build"` // 254263
	Manager string `json:"manager"`                // Synology DiskStation
	Vendor  string `json:"vender"`                 // Synology Inc.
	Model   string `json:"upnpmodelname"`          // DS1517+
	Version string `json:"udc_check_state"`        // 6.2.3
}

/*
 "platform": "Synology Inc.",
 "platformFamily": "Synology DiskStation DS1517+",
 "platformVersion": "6.2.3-254263",
*/

// GetSynology checks if the app is running on a Synology, and gets system info.
func (s *Snapshot) GetSynology(run bool) error { //nolint:cyclop
	if !run || !s.synology {
		return nil
	}

	file, err := os.Open(SynologyConf)
	if err != nil {
		return fmt.Errorf("opening synology conf: %w", err)
	}
	defer file.Close()

	// Start reading from the file with a reader.
	var (
		reader = bufio.NewReader(file)
		syn    = &Synology{}
	)

	for {
		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("reading synology conf: %w", err)
		}

		lsplit := strings.Split(line, "=")
		if len(lsplit) < 2 { //nolint:gomnd
			continue
		}

		switch lsplit[0] {
		case "last_admin_login_build":
			syn.Build = strings.Trim(lsplit[1], "\n\"")
		case "manager":
			syn.Manager = strings.Trim(lsplit[1], "\n\"")
		case "vender":
			syn.Vendor = strings.Trim(lsplit[1], "\n\"")
		case "upnpmodelname":
			syn.Model = strings.Trim(lsplit[1], "\n\"")
		case "udc_check_state":
			syn.Version = strings.Trim(lsplit[1], "\n\"")
		}
	}

	s.setSynology(syn)

	return nil
}

func (s *Snapshot) setSynology(syn *Synology) {
	if s.System.InfoStat.Platform == "" && syn.Vendor != "" {
		s.System.InfoStat.Platform = syn.Vendor
	}

	if s.System.InfoStat.PlatformFamily == "" && syn.Manager != "" {
		s.System.InfoStat.PlatformFamily = syn.Manager + " " + syn.Model
	}

	if s.System.InfoStat.PlatformVersion == "" && syn.Version != "" {
		s.System.InfoStat.PlatformVersion = syn.Version + "-" + syn.Build
	}
}
