package snapshot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
)

func (s *Snapshot) getDriveData(ctx context.Context, run bool, useSudo bool) (errs []error) {
	if !run {
		return nil
	}

	partitions, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return []error{fmt.Errorf("unable to get partitions: %w", err)}
	}

	devices := make(map[string]struct{})
	s.DriveAges = make(map[string]int)
	s.DriveTemps = make(map[string]int)
	s.DiskHealth = make(map[string]string)

	for _, partition := range partitions {
		log.Printf("[TEMPORARY] partition %s =-> %s", partition.Mountpoint, partition.Device)
		if _, ok := devices[partition.Device]; ok {
			continue
		}

		errs = append(errs, s.getDiskData(ctx, partition.Device, useSudo))
		errs = append(errs, s.getDiskHealth(ctx, partition.Device, useSudo))
		devices[partition.Device] = struct{}{}
	}

	return errs
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
