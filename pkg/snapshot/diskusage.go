package snapshot

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/shirou/gopsutil/v3/disk"
)

func (s *Snapshot) getDisksUsage(ctx context.Context, run bool, allDrives bool) []error {
	if !run {
		return nil
	}

	getAllDisks := allDrives || mnd.IsDocker

	partitions, err := disk.PartitionsWithContext(ctx, getAllDisks)
	if err != nil {
		return []error{fmt.Errorf("unable to get partitions: %w", err)}
	}

	s.DiskUsage = make(map[string]*Partition)

	var errs []error

	for idx := range partitions {
		usage, err := disk.UsageWithContext(ctx, partitions[idx].Mountpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to get partition usage: %s: %w", partitions[idx].Mountpoint, err))
			continue
		}

		if usage.Total == 0 ||
			((runtime.GOOS == "darwin" || strings.HasSuffix(runtime.GOOS, "bsd")) &&
				!strings.HasPrefix(partitions[idx].Device, "/dev/")) {
			continue
		}

		s.DiskUsage[partitions[idx].Device] = &Partition{
			Device: partitions[idx].Mountpoint,
			Total:  usage.Total,
			Free:   usage.Free,
			Used:   usage.Used,
		}
	}

	return errs
}

// Does not work on windows at all. Linux and Solaris only.
func (s *Snapshot) getZFSPoolData(ctx context.Context, pools []string) error {
	if len(pools) == 0 {
		return nil
	}

	cmd, stdout, waitg, err := readyCommand(ctx, false, "zpool", "list", "-pH")
	if err != nil {
		return err
	}

	s.ZFSPool = make(map[string]*Partition)

	go func() {
		for stdout.Scan() {
			fields := strings.Fields(stdout.Text())

			for _, pool := range pools {
				if len(fields) > 3 && strings.EqualFold(fields[0], pool) {
					s.ZFSPool[pool] = &Partition{Device: fields[4]}
					s.ZFSPool[pool].Total, _ = strconv.ParseUint(fields[1], mnd.Base10, mnd.Bits64)
					s.ZFSPool[pool].Free, _ = strconv.ParseUint(fields[3], mnd.Base10, mnd.Bits64)
				}
			}
		}
		waitg.Done()
	}()

	return runCommand(cmd, waitg)
}
