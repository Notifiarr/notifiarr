// +build !linux,!freebsd,!windows

package snapshot

import (
	"context"
)

// GetMemoryUsage returns current host memory consumption.
func (s *Snapshot) GetMemoryUsage(ctx context.Context, run bool) error {
	return s.getMemoryUsageShared(ctx, run)
}
