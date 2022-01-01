//go:build !linux && !freebsd && !windows

package snapshot

import (
	"context"
)

// GetMemoryUsage returns current host memory consumption.
func (s *Snapshot) GetMemoryUsage(ctx context.Context) error {
	return s.getMemoryUsageShared(ctx)
}
