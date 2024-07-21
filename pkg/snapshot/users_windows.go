package snapshot

import (
	// wapi "github.com/iamacarpet/go-win64api"
	"context"
)

// ErrNilUsers is a custom error to hopefully avoid a stack trace panic. Not sure.
// var ErrNilUsers = errors.New("user list was nil")

// GetUsers collects logged in users.
func (s *Snapshot) GetUsers(_ context.Context) error {
	/*
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
		/**/
	// This has a bug: https://github.com/iamacarpet/go-win64api/issues/36
	return nil
}
