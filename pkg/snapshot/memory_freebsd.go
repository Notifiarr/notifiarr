package snapshot

// This may work on all BSDs. Unknown.
// Tested on: FreeBSD 12.1-RELEASE r354233 GENERIC amd64 (VM on Slackware)

import (
	"context"

	"golang.org/x/sys/unix"
)

// GetMemoryUsage returns current host memory consumption.
func (s *Snapshot) GetMemoryUsage(ctx context.Context, run bool) error {
	if !run {
		return nil
	}

	pageSize, err := unix.SysctlUint32("hw.pagesize")
	if err != nil {
		return s.getMemoryUsageShared(ctx, run)
	}

	s.System.MemTotal, err = unix.SysctlUint64("hw.physmem")
	if err != nil {
		return s.getMemoryUsageShared(ctx, run)
	}

	// If the above two worked, these are unlikely to fail.
	var (
		memFree, _  = unix.SysctlUint32("vm.stats.vm.v_free_count")
		memUsed, _  = unix.SysctlUint32("vm.stats.vm.v_cache_count")
		inactive, _ = unix.SysctlUint32("vm.stats.vm.v_inactive_count")
	)

	s.System.MemFree = uint64(inactive+memUsed+memFree) * uint64(pageSize)
	s.System.MemUsed = s.System.MemTotal - s.System.MemFree

	return nil
}
