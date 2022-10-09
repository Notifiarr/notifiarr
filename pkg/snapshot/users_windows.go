package snapshot

import (
	"context"
	"fmt"

	wapi "github.com/iamacarpet/go-win64api"
)

// ErrNilUsers is a custom error to hopefully avoid a stack trace panic. Not sure.
var ErrNilUsers = fmt.Errorf("user list was nil")

// GetUsers collects logged in users.
func (s *Snapshot) GetUsers(ctx context.Context) error {
	// This has a bug: https://github.com/iamacarpet/go-win64api/issues/36
	return nil

	users, err := wapi.ListLoggedInUsers()
	if err != nil {
		return fmt.Errorf("getting userlist: %w", err)
	}

	if users == nil {
		return fmt.Errorf("getting userlist: %w", ErrNilUsers)
	}

	count := len(users)
	s.System.Users = count

	return nil
}
