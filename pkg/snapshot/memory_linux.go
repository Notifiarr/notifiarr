package snapshot

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
)

// GetMemoryUsage returns current host memory consumption.
func (s *Snapshot) GetMemoryUsage(ctx context.Context, run bool) error {
	if !run {
		return nil
	}

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return s.getMemoryUsageShared(ctx, run)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		switch fields := strings.Fields(scanner.Text()); {
		case len(fields) < 3: //nolint:gomnd
			continue
		case strings.EqualFold(fields[0], "MemTotal:"):
			s.System.MemTotal, _ = strconv.ParseUint(fields[1], 10, 64)
		case strings.EqualFold(fields[0], "MemAvailable:"):
			s.System.MemFree, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}

	s.System.MemTotal *= 1024
	s.System.MemFree *= 1024

	if s.System.MemTotal > 0 && s.System.MemFree > 0 {
		s.System.MemUsed = s.System.MemTotal - s.System.MemFree
	} else {
		return s.getMemoryUsageShared(ctx, run)
	}

	return nil
}
