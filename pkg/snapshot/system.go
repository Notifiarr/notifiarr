package snapshot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"sort"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// Processes allows us to sort a process list.
type Processes []*Process

// Process is a PID's basic info.
type Process struct {
	Name       string  `json:"name"`
	Pid        int32   `json:"pid"`
	MemPercent float32 `json:"memPercent"`
	CPUPercent float64 `json:"cpuPercent"`
}

// GetCPUSample gets a CPU percentage sample, CPU Times and Load Average.
func (s *Snapshot) GetCPUSample(ctx context.Context) error {
	times, err := cpu.TimesWithContext(ctx, false) // percpu, true/false
	if err != nil {
		return fmt.Errorf("unable to get cpu times: %w", err)
	}

	s.System.AvgStat, err = load.AvgWithContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get load avg: %w", err)
	}

	cpus, err := cpu.PercentWithContext(ctx, time.Second, false) // percpu, true/false
	if err != nil {
		return fmt.Errorf("unable to get cpu usage: %w", err)
	}

	s.System.CPUTime = times[0]
	s.System.CPU = cpus[0]

	return nil
}

func (s *Snapshot) getMemoryUsageShared(ctx context.Context) error {
	memory, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get memory usage: %w", err)
	}

	s.System.MemFree = memory.Free
	s.System.MemUsed = memory.Used
	s.System.MemTotal = memory.Total

	return nil
}

// GetLocalData collects current username, logged in user and host info.
func (s *Snapshot) GetLocalData(ctx context.Context) (errs []error) {
	u, err := user.Current()
	if err != nil {
		s.System.Username = "uid:" + strconv.Itoa(os.Getuid())
	} else {
		s.System.Username = u.Username
	}

	if err := s.GetUsers(ctx); err != nil &&
		!errors.Is(err, os.ErrNotExist) && !errors.Is(err, os.ErrPermission) {
		errs = append(errs, err)
	}

	if s.System.InfoStat, err = host.InfoWithContext(ctx); err != nil {
		errs = append(errs, fmt.Errorf("getting sysinfo/uptime: %w", err))
	}

	return errs
}

// GetProcesses collects 'count' processes by CPU usage.
func (s *Snapshot) GetProcesses(ctx context.Context, count int) (errs []error) {
	if count < 1 {
		return nil
	}

	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return []error{fmt.Errorf("process list: %w", err)}
	}

	s.Processes = make(Processes, len(procs))

	for idx, proc := range procs {
		s.Processes[idx] = &Process{Pid: proc.Pid}

		if s.Processes[idx].Name, err = proc.NameWithContext(ctx); err != nil {
			errs = append(errs, fmt.Errorf("pid %d, no name: %w", proc.Pid, err))
		}

		// This for loop primes the second run of PercentWithContext.
		// Then sleep a moment, and gather the cpu samples for all PIDs across that moment.
		_, _ = proc.PercentWithContext(ctx, 0)
	}

	time.Sleep(4 * time.Second) //nolint:gomnd

	for idx, proc := range procs {
		s.Processes[idx].CPUPercent, err = proc.PercentWithContext(ctx, 0)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, fmt.Errorf("pid %d, cpu percent: %w", proc.Pid, err))
		}

		s.Processes[idx].MemPercent, err = proc.MemoryPercentWithContext(ctx)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, fmt.Errorf("pid %d, mem percent: %w", proc.Pid, err))
		}
	}

	sort.Sort(s.Processes)
	s.Processes.Shrink(count)

	return errs
}

// Len allows us to sort Processes.
func (s Processes) Len() int {
	return len(s)
}

// Swap allows us to sort Processes.
func (s Processes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less allows us to sort Processes.
func (s Processes) Less(i, j int) bool {
	return s[i].CPUPercent > s[j].CPUPercent
}

// Shrink a process list.
func (s *Processes) Shrink(size int) {
	if s == nil {
		return
	}

	if len(*s) > size {
		*s = (*s)[:size]
	}
}
