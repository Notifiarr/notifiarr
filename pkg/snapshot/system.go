package snapshot

import (
	"context"
	"fmt"
	"os/user"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func (s *Snapshot) GetCPUSample(ctx context.Context, run bool) error {
	if !run {
		return nil
	}

	cpus, err := cpu.PercentWithContext(ctx, time.Second, false) // percpu, true/false
	if err != nil {
		return fmt.Errorf("unable to get cpu usage: %w", err)
	}

	s.System.CPU = cpus[0]

	return nil
}

func (s *Snapshot) getMemoryUsageShared(ctx context.Context, run bool) error {
	if !run {
		return nil
	}

	memory, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get memory usage: %w", err)
	}

	s.System.MemFree = memory.Free
	s.System.MemUsed = memory.Used
	s.System.MemTotal = memory.Total

	return nil
}

func (s *Snapshot) GetLocalData(ctx context.Context, run bool) (errs []error) {
	u, err := user.Current()
	if err != nil {
		errs = append(errs, fmt.Errorf("getting username: %w", err))
	} else {
		s.System.Username = u.Username
	}

	if !run {
		return errs
	}

	if s.System.InfoStat, err = host.InfoWithContext(ctx); err != nil {
		errs = append(errs, fmt.Errorf("getting sysinfo/uptime: %w", err))
	}

	return errs
}
