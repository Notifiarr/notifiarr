package logs

import "sync"

/* This is a helper method used in a couple spots. Doesn't have anything to do with logs. */

// Cooler allows you to save/get if something is already running or not (active).
type Cooler struct {
	lock   sync.Mutex
	active bool
}

// Active returns true if a cooler already active.
// Returns false if not, and activates the cooler.
func (c *Cooler) Active() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.active {
		return true
	}

	c.active = true

	return false
}

// Done resets a cooler.
func (c *Cooler) Done() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.active = false
}
