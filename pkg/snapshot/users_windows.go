package snapshot

import (
	"context"
	"fmt"

	wapi "github.com/iamacarpet/go-win64api"
)

// GetUsers collects logged in users.
func (s *Snapshot) GetUsers(ctx context.Context) error {
	users, err := wapi.ListLoggedInUsers()
	if err != nil {
		return fmt.Errorf("getting userlist: %w", err)
	}

	s.System.Users = len(users)

	return nil
}
