package update

import "os"

/* A fake signal, because why not. */

type Signal struct {
	Text string
}

// Make sure Command provides a Signal interface.
var _ = os.Signal(&Signal{})

// Signal allows you to pass this into an exit signal channel.
// This does not do anything, it only satisfies an interface.
func (s *Signal) Signal() {}

// String allows you to pass this into an exit signal channel.
func (s *Signal) String() string {
	if s.Text == "" {
		return "fake signal"
	}

	return s.Text
}
