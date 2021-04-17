package services

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hako/durafmt"
	"github.com/shirou/gopsutil/v3/process"
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

func GetAllProcesses() ([]*ProcInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting process list: %w", err)
	}

	p := []*ProcInfo{}

	for _, proc := range processes {
		procinfo, err := getProcInfo(ctx, proc)
		if err != nil {
			continue
		}

		procinfo.PID = proc.Pid
		p = append(p, procinfo)
	}

	return p, nil
}

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
	var (
		found []int32
		ages  []time.Time
	)

	// Loop each process/pid, get the command line name, and check for a match.
	for _, proc := range processes {
		procinfo, err := getProcInfo(ctx, proc)
		if err != nil {
			continue
		}

		// Look for a match for our process.
		if strings.Contains(procinfo.CmdLine, s.Value) ||
			(s.proc.checkRE != nil && s.proc.checkRE.FindString(procinfo.CmdLine) != "") {
			found = append(found, proc.Pid)
			ages = append(ages, procinfo.Created)

			log.Printf("pid: %d, age: %v, cmd: %v",
				proc.Pid, time.Since(procinfo.Created).Round(time.Second), procinfo.CmdLine)

			if s.proc.running && time.Since(procinfo.Created) > s.Interval.Duration {
				return &result{
					state: StateCritical,
					output: fmt.Sprintf("%s: process restarted since last check, age: %v, pid: %d, proc: %s",
						s.Value, time.Since(procinfo.Created), proc.Pid, procinfo.CmdLine),
				}
			}
		}
	}

	return s.checkProcessCounts(found, ages)
}

// checkProcessCounts validates process check thresholds.
func (s *Service) checkProcessCounts(pids []int32, ages []time.Time) *result {
	count := len(pids)

	agesText := ""
	if len(ages) == 1 {
		agesText = fmt.Sprintf(", age: %v", durafmt.ParseShort(time.Since(ages[0]).Round(time.Second)))
	}

	r := &result{
		state: StateOK,
		output: fmt.Sprintf("%s: found %d processes; min %d, max: %d%s, pids: %v",
			s.Value, count, s.proc.countMin, s.proc.countMax, agesText, pids),
	}

	switch {
	case s.proc.countMax != 0 && count > s.proc.countMax:
		r.state = StateCritical
	case s.proc.countMin != 0 && count < s.proc.countMin:
		r.state = StateCritical
	case s.proc.running && count > 0:
		r.state = StateCritical
		r.output = fmt.Sprintf("%s: found %d processes; max 0 (should not be running)%s, pids: %v",
			s.Value, count, agesText, pids)
	case !s.proc.running && count == 0:
		r.state = StateCritical
		r.output = s.Value + ": not running!"
	}

	return r
}

// getProcInfo returns age and cli args for a process.
func getProcInfo(ctx context.Context, p *process.Process) (*ProcInfo, error) {
	var (
		err      error
		procinfo ProcInfo
	)

	procinfo.CmdLine, err = p.CmdlineWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("CmdlineWithContext: %w", err)
	}

	created, err := p.CreateTimeWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateTimeWithContext: %w", err)
	}

	procinfo.Created = time.Unix(created/epochOffset, 0).Round(time.Millisecond)

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
