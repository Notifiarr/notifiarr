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

func (s *Snapshot) getDisksUsage(ctx context.Context, run bool) []error {
	if !run {
		return nil
	}

	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return []error{fmt.Errorf("unable to get partitions: %w", err)}
	}

	s.DiskUsage = make(map[string]*Partition)

	var errs []error

	for i := range partitions {
		u, err := disk.UsageWithContext(ctx, partitions[i].Mountpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to get partition usage: %s: %w", partitions[i].Mountpoint, err))
			continue
		}

		if u.Total == 0 ||
			((runtime.GOOS == "darwin" || strings.HasSuffix(runtime.GOOS, "bsd")) &&
				!strings.HasPrefix(partitions[i].Device, "/dev/")) {
			continue
		}

		s.DiskUsage[partitions[i].Device] = &Partition{
			Device: partitions[i].Mountpoint,
			Total:  u.Total,
			Free:   u.Free,
		}
	}

	return errs
}

// Does not work on windows at all. Linux and Solaris only.
func (s *Snapshot) getZFSPoolData(ctx context.Context, pools []string) error {
	if len(pools) == 0 {
		return nil
	}

	cmd, stdout, wg, err := readyCommand(ctx, false, "zpool", "list", "-pH")
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
		wg.Done()
	}()

	return runCommand(cmd, wg)
}
