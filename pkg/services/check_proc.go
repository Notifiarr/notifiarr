package services

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/shirou/gopsutil/v4/process"
)

// ProcInfo is derived from a pid.
type ProcInfo struct {
	CmdLine string
	Created time.Time
	PID     int32
	/* // these are possibly available..
		cmdlineslice []string
		cwd          string
		exe          string
		meminfo      *process.MemoryInfoStat
		memperc      float32
		name         string
	/**/
}

// procExpect is setup for each 'process' service from input data on initialization.
type procExpect struct {
	checkRE  *regexp.Regexp
	countMin int
	countMax int
	restarts bool
	running  bool
}

const epochOffset = 1000

// GetAllProcesses returns all running process on the host.
func GetAllProcesses(ctx context.Context) ([]*ProcInfo, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting process list: %w", err)
	}

	procs := []*ProcInfo{}

	for _, proc := range processes {
		procinfo, errs := getProcInfo(ctx, proc)
		if len(errs) != 0 {
			if procinfo == nil || procinfo.CmdLine == "" {
				// log.Printf("DEBUG: [pid: %d] proc errs: %v", proc.Pid, errs)
				continue
			}
			// log.Printf("DEBUG: [pid: %d] proc errs: %v, cmd: %s", proc.Pid, errs, procinfo.CmdLine)
		} //nolint:wsl

		procinfo.PID = proc.Pid
		procs = append(procs, procinfo)
	}

	return procs, nil
}

// start a loop through processes to find the one we care about.
func (s *Service) checkProccess(ctx context.Context) *result {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout.Duration)
	defer cancel()

	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return &result{
			state:  StateUnknown,
			output: &Output{str: "process list error: " + err.Error()},
		}
	}

	return s.getProcessResults(ctx, processes)
}

func (s *Service) getProcessResults(ctx context.Context, processes []*process.Process) *result {
	var (
		found []int32
		ages  []time.Time
	)

	// Loop each process/pid, get the command line name, and check for a match.
	for _, proc := range processes {
		procinfo, _ := getProcInfo(ctx, proc)
		if procinfo.CmdLine == "" {
			continue
		}

		// Look for a match for our process.
		if strings.Contains(procinfo.CmdLine, s.Value) ||
			(s.svc.proc.checkRE != nil && s.svc.proc.checkRE.FindString(procinfo.CmdLine) != "") {
			found = append(found, proc.Pid)
			ages = append(ages, procinfo.Created)

			// log.Printf("pid: %d, age: %v, cmd: %v",
			// 	proc.Pid, time.Since(procinfo.Created).Round(time.Second), procinfo.CmdLine)

			if !procinfo.Created.IsZero() && s.svc.proc.restarts && time.Since(procinfo.Created) < s.Interval.Duration {
				return &result{
					state: StateCritical,
					output: &Output{str: fmt.Sprintf("%s: process restarted since last check, age: %v, pid: %d, proc: %s",
						s.Value, time.Since(procinfo.Created), proc.Pid, procinfo.CmdLine)},
				}
			}
		}
	}

	return s.checkProcessCounts(found, ages)
}

// checkProcessCounts validates process check thresholds.
func (s *Service) checkProcessCounts(pids []int32, ages []time.Time) *result {
	min, max, age, pid := s.getProcessStrings(pids, ages)

	switch count := len(pids); {
	case !s.svc.proc.running && count == 0: // not running!
		fallthrough
	case s.svc.proc.countMax != 0 && count > s.svc.proc.countMax: // too many running!
		fallthrough
	case count < s.svc.proc.countMin: // not enough running!
		return &result{
			state:  StateCritical,
			output: &Output{str: fmt.Sprintf("%s: found %d processes; %s%s%s%s", s.Value, count, min, max, age, pid)},
		}
	case s.svc.proc.running && count > 0: // running but should not be!
		return &result{
			state:  StateCritical,
			output: &Output{str: fmt.Sprintf("%s: found %d processes; expected: 0%s%s", s.Value, count, age, pid)},
		}
	default: // running within thresholds!
		return &result{
			state:  StateOK,
			output: &Output{str: fmt.Sprintf("%s: found %d processes; %s%s%s%s", s.Value, count, min, max, age, pid)},
		}
	}
}

// getProcessStrings compiles output strings for a process service check.
func (s *Service) getProcessStrings(pids []int32, ages []time.Time) (string, string, string, string) {
	var (
		min           = "min: 1"
		max, age, pid string
	)

	if s.svc.proc.countMin > 0 { // min always exists.
		min = fmt.Sprintf("min: %d", s.svc.proc.countMin)
	}

	if s.svc.proc.countMax > 0 {
		max = fmt.Sprintf(", max: %d", s.svc.proc.countMax)
	}

	if len(ages) == 1 && !ages[0].IsZero() {
		age = mnd.DurationAge(ages[0])
	}

	for _, activePid := range pids {
		if pid == "" {
			pid = ", pids: "
		} else {
			pid += ";"
		}

		pid += strconv.Itoa(int(activePid))
	}

	return min, max, age, pid
}

// getProcInfo returns age and cli args for a process.
func getProcInfo(ctx context.Context, proc *process.Process) (*ProcInfo, []error) {
	var (
		err      error
		errs     []error
		procinfo ProcInfo
	)

	procinfo.CmdLine, err = proc.CmdlineWithContext(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("CmdlineWithContext: %w", err))
	}

	if procinfo.CmdLine == "" {
		if procinfo.CmdLine, err = proc.NameWithContext(ctx); err != nil {
			errs = append(errs, fmt.Errorf("NameWithContext: %w", err))
		}
	}

	// FreeBSD doesn't have create time.
	if !mnd.IsFreeBSD {
		created, err := proc.CreateTimeWithContext(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("CreateTimeWithContext: %w", err))
		} else {
			procinfo.Created = time.Unix(created/epochOffset, 0).Round(time.Millisecond)
		}
	}

	/*
			procinfo.cmdlineslice, err = p.CmdlineSliceWithContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("CmdlineSliceWithContext: %w", err)
			}
		  procinfo.name, err = p.NameWithContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("NameWithContext: %w", err)
			}
			procinfo.exe, err = p.ExeWithContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("ExeWithContext: %w", err)
			}
			procinfo.meminfo, err = p.MemoryInfoWithContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("MemoryInfoWithContext: %w", err)
			}
			procinfo.memperc, err = p.MemoryPercentWithContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("MemoryPercentWithContext: %w", err)
			}
	*/

	return &procinfo, errs
}
