// +build !linux,!freebsd

package snapshot

import (
	"context"
)

func (s *Snapshot) GetMemoryUsage(ctx context.Context, run bool) error {
	return s.getMemoryUsageShared(ctx, run)
}
