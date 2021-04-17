package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// procInfo is derived from a pid.
type procInfo struct {
	cmdline string
	created time.Time
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

// start a loop through processes to find the one we care about.
func (s *Service) checkProccess() *result {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return &result{
			state:  StateUnknown,
			output: "process list error: " + err.Error(),
		}
	}

	return s.getProcessResults(ctx, processes)
}

func (s *Service) getProcessResults(ctx context.Context, processes []*process.Process) *result {
	var found []int32

	// Loop each process/pid, get the command line name, and check for a match.
	for _, proc := range processes {
		procinfo, err := getProcInfo(ctx, proc)
		if err != nil {
			continue
		}

		// Look for a match for our process.
		if strings.Contains(procinfo.cmdline, s.Value) ||
			(s.proc.checkRE != nil && s.proc.checkRE.FindString(procinfo.cmdline) != "") {
			found = append(found, proc.Pid)

			// log.Printf("pid: %d, age: %v, cmd: %v",
			// 	proc.Pid, time.Since(procinfo.created).Round(time.Second), procinfo.cmdline)

			if s.proc.running && time.Since(procinfo.created) > s.Interval.Duration {
				return &result{
					state: StateCritical,
					output: fmt.Sprintf("%s: process restarted since last check, age: %v, pid: %d, proc: %s",
						s.Value, time.Since(procinfo.created), proc.Pid, procinfo.cmdline),
				}
			}
		}
	}

	return s.checkProcessCounts(found)
}

//
func (s *Service) checkProcessCounts(pids []int32) *result {
	count := len(pids)
	r := &result{
		state: StateOK,
		output: fmt.Sprintf("%s: found %d processes; min %d, max: %d, pids: %v",
			s.Value, count, s.proc.countMin, s.proc.countMax, pids),
	}

	switch {
	case s.proc.countMax != 0 && count > s.proc.countMax:
		r.state = StateCritical
	case s.proc.countMin != 0 && count < s.proc.countMin:
		r.state = StateCritical
	case s.proc.running && count > 0:
		r.state = StateCritical
		r.output = fmt.Sprintf("%s: found %d processes; max 0 (should not be running), pids: %v", s.Value, count, pids)
	case !s.proc.running && count == 0:
		r.state = StateCritical
		r.output = s.Value + ": not running!"
	}

	return r
}

// getProcInfo returns age and cli args for a process.
func getProcInfo(ctx context.Context, p *process.Process) (*procInfo, error) {
	var (
		err      error
		procinfo procInfo
	)

	procinfo.cmdline, err = p.CmdlineWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("CmdlineWithContext: %w", err)
	}

	created, err := p.CreateTimeWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateTimeWithContext: %w", err)
	}

	procinfo.created = time.Unix(created/epochOffset, 0).Round(time.Millisecond)

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

	return &procinfo, nil
}
