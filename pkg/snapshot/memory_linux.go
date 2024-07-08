package snapshot

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

// GetMemoryUsage returns current host memory consumption.
func (s *Snapshot) GetMemoryUsage(ctx context.Context) error {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return s.getMemoryUsageShared(ctx)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		switch fields := strings.Fields(scanner.Text()); {
		case len(fields) < 3: //nolint:mnd
			continue
		case strings.EqualFold(fields[0], "MemTotal:"):
			s.System.MemTotal, _ = strconv.ParseUint(fields[1], mnd.Base10, mnd.Bits64)
		case strings.EqualFold(fields[0], "MemAvailable:"):
			s.System.MemFree, _ = strconv.ParseUint(fields[1], mnd.Base10, mnd.Bits64)
		}
	}

	// Not deferred because we want it closed before calling s.getMemoryUsageShared().
	_ = file.Close()

	s.System.MemTotal *= 1024
	s.System.MemFree *= 1024

	if s.System.MemTotal > 0 && s.System.MemFree > 0 {
		s.System.MemUsed = s.System.MemTotal - s.System.MemFree
	} else {
		return s.getMemoryUsageShared(ctx)
	}

	return nil
}
