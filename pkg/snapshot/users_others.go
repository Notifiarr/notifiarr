//go:build !windows

package snapshot

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v4/host"
)

// GetUsers collects logged in users.
func (s *Snapshot) GetUsers(ctx context.Context) error {
	users, err := host.UsersWithContext(ctx)
	if err != nil {
		return fmt.Errorf("getting userlist: %w", err)
	}

	s.System.Users = len(users)

	return nil
}
