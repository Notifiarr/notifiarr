package cooldown

import (
	"time"
)

const defaultCleanTimer = time.Minute

type cooler struct {
	// Key we are cooling down.
	Key string
	// How long to cool down the key.
	Dur  time.Duration
	Last time.Time
}

// Timer is used to set a cooldown timer on a key.
type Timer struct {
	skipCleanup bool
	cleanTimer  time.Duration
	track       map[string]*cooler
	ch          chan *cooler
	rep         chan bool
}

// NewTimer returns a struct for which you can use Active().
func NewTimer(skipCleanup bool, cleanTimer time.Duration) *Timer {
	t := &Timer{
		skipCleanup: skipCleanup,
		cleanTimer:  cleanTimer,
	}
	t.start()

	return t
}

// Active returns true if a cooldown timer is active for a key. If it's not active, it's saved (for next time).
func (t *Timer) Active(key string, coolFor time.Duration) bool {
	t.ch <- &cooler{Key: key, Dur: coolFor}
	return <-t.rep
}

// StopTimer kills the active cooler. Do not call Activ() after you call this method.
func (t *Timer) StopTimer() {
	t.ch <- nil // stop signal is nil
	<-t.rep
	close(t.rep)
	t.rep = nil
}

// start sets up the Timer.
func (t *Timer) start() {
	cleanTimer := defaultCleanTimer
	if t.cleanTimer > 0 {
		cleanTimer = t.cleanTimer
	}

	t.track = make(map[string]*cooler)
	t.ch = make(chan *cooler)
	t.rep = make(chan bool)

	timer := time.NewTicker(cleanTimer)
	if t.skipCleanup {
		timer.Stop()
	}

	go t.chanWatcher(timer)
}

// stop closes everything.
func (t *Timer) stop(timer *time.Ticker) {
	for key := range t.track {
		t.track[key] = nil
		delete(t.track, key)
	}

	timer.Stop()

	t.track = nil
	close(t.ch)
	t.ch = nil
	t.rep <- true
}

// chanWatcher runs a loop, and deletes any keys older than the last cooldown we had for that key.
func (t *Timer) chanWatcher(timer *time.Ticker) {
	defer t.stop(timer)

	for {
		select {
		case cooler := <-t.ch:
			if cooler == nil {
				return // nil signals a stop.
			}

			if t.track[cooler.Key] != nil && time.Since(t.track[cooler.Key].Last) < cooler.Dur {
				t.rep <- true
				continue
			}

			cooler.Last = time.Now()
			t.track[cooler.Key] = cooler
			t.rep <- false
		case tick := <-timer.C:
			for key, val := range t.track {
				if tick.After(val.Last.Add(val.Dur)) {
					delete(t.track, key)
				}
			}
		}
	}
}

// Running returns false if StopTimer has been called; true otherwise.
func (t *Timer) Running() bool {
	return t.ch != nil
}

// Len returns the tracked item length and the channel queue length.
func (t *Timer) Len() (int, int) {
	return len(t.track), len(t.ch)
}
