package snapshot

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/shirou/gopsutil/v4/host"
)

var ErrNotSynology = errors.New("the running host is not a Synology")

// Synology is the data we care about from the config file.
//
//nolint:tagliatelle
type Synology struct {
	Build   string            `json:"last_admin_login_build"` // 254263
	Manager string            `json:"manager"`                // Synology DiskStation
	Vendor  string            `json:"vender"`                 // Synology Inc.
	Model   string            `json:"upnpmodelname"`          // DS1517+
	Version string            `json:"udc_check_state"`        // 6.2.3
	HA      map[string]string `json:"ha"`
}

/*
 "platform": "Synology Inc.",
 "platformFamily": "Synology DiskStation DS1517+",
 "platformVersion": "6.2.3-254263",
*/

// GetSynology checks if the app is running on a Synology, and gets system info.
func GetSynology(snapshot bool) (*Synology, error) { //nolint:cyclop
	if !mnd.IsSynology {
		return nil, ErrNotSynology
	}

	synoHA, err := getSynoHAStats(snapshot)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(mnd.Synology)
	if err != nil {
		return nil, fmt.Errorf("opening synology conf: %w", err)
	}
	defer file.Close()

	// Start reading from the file with a reader.
	var (
		reader = bufio.NewReader(file)
		syn    = &Synology{HA: synoHA}
	)

	for {
		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("reading synology conf: %w", err)
		}

		lsplit := strings.Split(line, "=")
		if len(lsplit) < 2 { //nolint:mnd
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

	return syn, nil
}

// getSynoHAStats uses `sudo synoha` to pull high-availability disk statuses.
// This package is not installed on most systems,
// this function ends at the LookPath in that case.
func getSynoHAStats(run bool) (map[string]string, error) {
	if !run {
		return nil, nil //nolint:nilnil
	}

	synoha, err := exec.LookPath("synoha")
	if err != nil {
		// If the tool doesn't exist, bail without an error.
		return nil, nil //nolint:nilerr,nilnil // Callers must handle this.
	}

	cmds := []string{ // Add more if you need more.
		"status", "local-role", "local-name", "lnode-status", "local-status",
		"remote-name", "remote-role", "rnode-status", "remote-status", "remote-ip",
	}

	output := make(map[string]string) // This is the returned value.
	cmdout := bytes.Buffer{}          // Reuse buffer for every command.

	for _, arg := range cmds {
		cmdout.Reset()

		cmd := exec.Command("sudo", synoha, "--"+arg)
		cmd.Stderr = io.Discard // Do not care about error output.
		cmd.Stdout = &cmdout    // Use buffer for command output.

		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("synoha failed: %w", err)
		}

		// Put this command's output into the output map using the arg-name as key.
		output[arg] = strings.TrimSpace(cmdout.String()) // Remove final newline from output.
	}

	return output, nil
}

// SetInfo writes synology data INTO the provided InfoStat.
func (s *Synology) SetInfo(hostInfo *host.InfoStat) {
	if hostInfo.Platform == "" && s.Vendor != "" {
		hostInfo.Platform = s.Vendor
	}

	if hostInfo.PlatformFamily == "" && s.Manager != "" {
		hostInfo.PlatformFamily = s.Manager + " " + s.Model
	}

	if hostInfo.PlatformVersion == "" && s.Version != "" {
		hostInfo.PlatformVersion = s.Version + "-" + s.Build
	}
}
