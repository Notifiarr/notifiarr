package cooldown

import (
	"sync"
	"time"
)

const cleanTimer = time.Minute

type cooler struct {
	// Key we are cooling down.
	Key string
	// How long to cool down the key.
	Dur  time.Duration
	when time.Time
}

// Timer is used to set a cooldown timer on a key.
type Timer struct {
	track map[string]*cooler
	ch    chan *cooler
	rep   chan bool
	once  sync.Once
}

// Active returns true if a timer is active, otherwise it becomes active.
func (t *Timer) Active(key string, coolFor time.Duration) bool {
	t.once.Do(t.coolerCleaner)
	t.ch <- &cooler{Key: key, Dur: coolFor}

	return <-t.rep
}

// Cooler cleaner runs a loop every minute, and deletes any keys older than the last cooldown we had for that key.
func (t *Timer) coolerCleaner() {
	t.track = make(map[string]*cooler)
	t.ch = make(chan *cooler)
	t.rep = make(chan bool)
	timer := time.NewTicker(cleanTimer)

	go func() {
		for {
			select {
			case cooler := <-t.ch:
				if t.track[cooler.Key] == nil || time.Since(t.track[cooler.Key].when) > cooler.Dur {
					t.rep <- false

					cooler.when = time.Now()
					t.track[cooler.Key] = cooler
				} else {
					t.rep <- true
				}
			case tick := <-timer.C:
				for key, val := range t.track {
					if tick.After(val.when.Add(val.Dur)) {
						delete(t.track, key)
					}
				}
			}
		}
	}()
}
