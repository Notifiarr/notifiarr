package snapshot

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/shirou/gopsutil/v4/disk"
)

func (s *Snapshot) getDisksUsage(ctx context.Context, run bool, allDrives bool) []error { //nolint:cyclop
	if !run {
		return nil
	}

	var (
		errs        []error
		getAllDisks = allDrives || mnd.IsDocker
	)

	partitions, err := disk.PartitionsWithContext(ctx, getAllDisks)
	if err != nil {
		errs = append(errs, fmt.Errorf("unable to get partitions: %w", err))
	}

	s.DiskUsage = make(map[string]*Partition)

	for idx := range partitions {
		usage, err := disk.UsageWithContext(ctx, partitions[idx].Mountpoint)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to get partition usage: %s: %w", partitions[idx].Mountpoint, err))
			continue
		}

		// skip tmpfs volumes
		if usage.Fstype == "tmpfs" ||
			// skip read only volumes with no device.
			(slices.Contains(partitions[idx].Opts, "ro") && slices.Contains(partitions[idx].Opts, "nodev")) ||
			// skip hidden volumes on macos.
			(mnd.IsDarwin && slices.Contains(partitions[idx].Opts, "nobrowse")) {
			continue
		}

		if usage.Total == 0 ||
			((mnd.IsDarwin || strings.HasSuffix(runtime.GOOS, "bsd")) &&
				!strings.HasPrefix(partitions[idx].Device, "/dev/")) {
			continue
		}

		if usage.Used == 0 && usage.Free > 0 && usage.Total > usage.Free {
			usage.Used = usage.Total - usage.Free
		}

		s.DiskUsage[partitions[idx].Device] = &Partition{
			Device:   partitions[idx].Mountpoint,
			Total:    usage.Total,
			Free:     usage.Free,
			Used:     usage.Used,
			FSType:   usage.Fstype,
			ReadOnly: slices.Contains(partitions[idx].Opts, "ro"),
			Opts:     partitions[idx].Opts,
		}
	}

	return errs
}

func (s *Snapshot) getQuota(ctx context.Context, run bool) error {
	if !run {
		return nil
	}

	cmd, stdout, waitg, err := readyCommand(ctx, false, "quota", "--no-wrap", "--show-mntpoint", "--human-readable")
	if err != nil {
		return err
	}

	s.Quotas = make(map[string]*Partition)

	go func() {
		for stdout.Scan() {
			fields := strings.Fields(stdout.Text())
			if len(fields) < 4 || fields[0][0] != '/' { // partitions tend to start with a slash.
				continue
			}

			// 	Filesystem                    mount space   quota     limit           grace  files   quota   limit   grace
			// /dev/mapper/ubuntu--vg-ubuntu--lv /  95216K  10485760K 11534336K       0      1k      0k      0k       0
			space := getQuotaSize(fields[2])
			quota := getQuotaSize(fields[3])
			s.Quotas[fields[0]] = &Partition{
				Device: fields[1],
				Total:  uint64(quota),
				Free:   uint64(quota - space),
				Used:   uint64(space),
			}
		}
		waitg.Done()
	}()

	if err := runCommand(cmd, waitg); err != nil {
		return fmt.Errorf("PLEASE REPORT THIS ERROR: %w", err)
	}

	return nil
}

func getQuotaSize(line string) int {
	size, _ := strconv.Atoi(strings.TrimRight(line, "KMGT"))

	switch line[len(line)-1] {
	case 'K':
		return size * mnd.Kilobyte
	case 'M':
		return size * mnd.Megabyte
	case 'G':
		return size * mnd.Megabyte * mnd.Kilobyte
	case 'T':
		return size * mnd.Megabyte * mnd.Megabyte
	default:
		return size
	}
}

// Does not work on windows at all. Linux and Solaris only.
func (s *Snapshot) getZFSPoolData(ctx context.Context, pools []string) error {
	if len(pools) == 0 {
		return nil
	}

	// # zpool list -pH
	// NAME   SIZE          ALLOC         FREE          CKPOINT EXPANDSZ FRAG CAP DEDUP HEALTH  ALTROOT
	// data   3985729650688 2223640698880 1762088951808 -       -        10   55  1.00  ONLINE  -
	// data2  996432412672  98463039488   897969373184  -       -        8    9   1.00  ONLINE  -
	// data3  996432412672  44307656704   952124755968  -       -        4    4   1.00  ONLINE  -
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
					s.ZFSPool[pool] = &Partition{Device: fields[4], FSType: "zfs", Opts: []string{fields[9]}}
					s.ZFSPool[pool].Total, _ = strconv.ParseUint(fields[1], mnd.Base10, mnd.Bits64)
					s.ZFSPool[pool].Free, _ = strconv.ParseUint(fields[3], mnd.Base10, mnd.Bits64)
				}
			}
		}
		waitg.Done()
	}()

	return runCommand(cmd, waitg)
}
