package snapshot

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/v3/disk"
)

var ErrNoDisks = fmt.Errorf("no disks found")

func (s *Snapshot) getDriveData(ctx context.Context, run bool, useSudo bool) (errs []error) {
	if !run {
		return nil
	}

	finder := getParts

	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		finder = getBlocks
	}

	disks, err := finder(ctx)
	if err != nil {
		return []error{err}
	}

	if len(disks) == 0 {
		return []error{ErrNoDisks}
	}

	devices := make(map[string]struct{})
	s.DriveAges = make(map[string]int)
	s.DriveTemps = make(map[string]int)
	s.DiskHealth = make(map[string]string)

	for _, disk := range disks {
		if _, ok := devices[disk]; ok {
			continue
		}

		errs = append(errs, s.getDiskData(ctx, disk, useSudo))
		errs = append(errs, s.getDiskHealth(ctx, disk, useSudo))
		devices[disk] = struct{}{}
	}

	return errs
}

// works well on mac and linux, probably windows too.
func getBlocks(ctx context.Context) ([]string, error) {
	block, err := ghw.Block()
	if err != nil {
		return nil, fmt.Errorf("unable to get block devices: %w", err)
	}

	list := []string{}

	for _, dev := range block.Disks {
		if runtime.GOOS != "windows" {
			list = append(list, path.Join("/dev", dev.Name))
		} else {
			list = append(list, dev.Name)
		}
	}

	return list, nil
}

// use this for everything else....
func getParts(ctx context.Context) ([]string, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("unable to get partitions: %w", err)
	}

	list := []string{}

	for _, part := range partitions {
		list = append(list, part.Device)
	}

	return list, nil
}

func (s *Snapshot) getDiskData(ctx context.Context, disk string, useSudo bool) error { //nolint: cyclop
	cmd, stdout, wg, err := readyCommand(ctx, useSudo, "smartctl", "-A", disk)
	if err != nil {
		return err
	}

	go func() {
		for stdout.Scan() {
			switch fields := strings.Fields(stdout.Text()); {
			case len(fields) > 1 && fields[0] == "Temperature:":
				s.DriveTemps[disk], _ = strconv.Atoi(fields[1])
			case len(fields) > 3 && fields[0]+fields[1]+fields[2] == "PowerOnHours:":
				s.DriveAges[disk], _ = strconv.Atoi(strings.ReplaceAll(fields[3], ",", ""))
			case len(fields) < 10: // nolint: gomnd
				continue
			case strings.HasPrefix(fields[1], "Airflow_Temp") ||
				strings.HasPrefix(fields[1], "Temperature_Cel"):
				s.DriveTemps[disk], _ = strconv.Atoi(fields[9])
			case strings.HasPrefix(fields[1], "Power_On_Hour"):
				s.DriveAges[disk], _ = strconv.Atoi(fields[9])
			}
		}
		wg.Done()
	}()

	return runCommand(cmd, wg)
}

func (s *Snapshot) getDiskHealth(ctx context.Context, disk string, useSudo bool) error {
	cmd, stdout, wg, err := readyCommand(ctx, useSudo, "smartctl", "-H", disk)
	if err != nil {
		return err
	}

	go func() {
		for stdout.Scan() {
			if text := stdout.Text(); strings.Contains(text, "self-assessment ") {
				s.DiskHealth[disk] = text[strings.LastIndex(text, " ")+1:]
			}
		}
		wg.Done()
	}()

	return runCommand(cmd, wg)
}
