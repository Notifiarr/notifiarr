package notifiarr

import (
	"sync"
	"time"
)

// Timer is used to set a cooldown time.
type Timer struct {
	lock  sync.Mutex
	start time.Time
}

// Active returns true if a timer is active, otherwise it becomes active.
func (t *Timer) Active(d time.Duration) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	if time.Since(t.start) < d {
		return true
	}

	t.start = time.Now()

	return false
}
