package snapshot

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// NvidiaConfig is our intput data.
type NvidiaConfig struct {
	SMIPath  string   `toml:"smi_path" xml:"smi_path" json:"smiPath"`
	BusIDs   []string `toml:"bus_ids" xml:"bus_id" json:"busIDs"`
	Disabled bool     `toml:"disabled" xml:"disabled" json:"disabled"`
}

// NvidiaOutput is what we send to the website.
type NvidiaOutput struct {
	Name        string
	Driver      string
	Pstate      string
	Vbios       string
	BusID       string
	Temperature int
	Utilization int
	MemTotal    int
	MemFree     int
}

// HasID returns true if the ID is requested, or no IDs are filtered.
func (n *NvidiaConfig) HasID(busID string) bool {
	if len(n.BusIDs) == 0 {
		return true
	}

	for _, id := range n.BusIDs {
		if id == busID {
			return true
		}
	}

	return false
}

// GetNvidia requires nvidia-smi executable and Nvidia drivers.
func (s *Snapshot) GetNvidia(ctx context.Context, config *NvidiaConfig) error {
	if config == nil || config.Disabled {
		return nil
	}

	var err error

	cmdPath := config.SMIPath
	if cmdPath != "" {
		if _, err = os.Stat(cmdPath); err != nil {
			return fmt.Errorf("unable to locate nvidia-smi at provided path '%s': %w", cmdPath, err)
		}
	} else if cmdPath, err = exec.LookPath("nvidia-smi"); err != nil {
		// do not throw an error if nvidia-smi is missing.
		// return fmt.Errorf("nvidia-smi missing! %w", err)
		return nil //nolint:nilerr
	}

	cmd := exec.CommandContext(
		ctx,
		cmdPath,
		"--format=csv,noheader",
		"--query-gpu=name,"+
			"driver_version,"+
			"pstate,"+
			"vbios_version,"+
			"pci.bus_id,"+
			"temperature.gpu,"+
			"utilization.gpu,"+
			"memory.total,"+
			"memory.free",
	)
	sysCallSettings(cmd)

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)

	go s.scanNvidiaSMIOutput(stdout, config, &wg)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v: %w: %s", cmd.Args, err, stderr)
	}

	return nil
}

func (s *Snapshot) scanNvidiaSMIOutput(stdout io.Reader, config *NvidiaConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	// Output (1 card per line):
	// NVIDIA GeForce GTX 1060 3GB, 465.19.01, P0, 86.06.3C.40.17, 00000000:04:00.0, 63, 2 %, 3017 MiB, 3017 MiB
	for scanner.Scan() {
		item := strings.Split(scanner.Text(), ",")
		if len(item) != reflect.TypeOf(config).NumField() {
			continue
		}

		busID := strings.TrimSpace(item[4])
		if !config.HasID(busID) {
			continue
		}

		output := NvidiaOutput{
			Name:   strings.TrimSpace(item[0]),
			Driver: strings.TrimSpace(item[1]),
			Pstate: strings.TrimSpace(item[2]),
			Vbios:  strings.TrimSpace(item[3]),
			BusID:  busID,
		}
		output.Temperature, _ = strconv.Atoi(strings.TrimSpace(item[5]))
		output.Utilization, _ = strconv.Atoi(strings.Fields(item[6])[0])
		output.MemTotal, _ = strconv.Atoi(strings.Fields(item[7])[0])
		output.MemFree, _ = strconv.Atoi(strings.Fields(item[8])[0])

		s.Nvidia = append(s.Nvidia, &output)
	}
}
