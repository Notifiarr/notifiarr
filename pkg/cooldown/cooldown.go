package cooldown

import (
	"time"
)

const defaultTickerInterval = time.Minute

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
	ticker      time.Duration
	track       map[string]*cooler
	ch          chan *cooler
	rep         chan bool
}

// NewTimer returns a struct for which you can use Active().
func NewTimer(skipCleanup bool, cleanTimer time.Duration) *Timer {
	timer := &Timer{
		skipCleanup: skipCleanup,
		ticker:      cleanTimer,
	}
	timer.start()

	return timer
}

// Active returns true if a cooldown timer is active for a key. If it's not active, it's saved (for next time).
func (t *Timer) Active(key string, coolFor time.Duration) bool {
	t.ch <- &cooler{Key: key, Dur: coolFor}
	return <-t.rep
}

// StopTimer kills the active cooler. Do not call Active() after you call this method.
func (t *Timer) StopTimer() {
	t.ch <- nil // stop signal is nil.
	<-t.rep     // wait until finished.
	t.rep = nil // last but not least.
}

// start sets up the Timer.
func (t *Timer) start() {
	tickerInterval := defaultTickerInterval
	if t.ticker > 0 {
		tickerInterval = t.ticker
	}

	t.track = make(map[string]*cooler)
	t.ch = make(chan *cooler)
	t.rep = make(chan bool)

	ticker := time.NewTicker(tickerInterval)
	if t.skipCleanup {
		ticker.Stop()
	}

	go t.chanWatcher(ticker)
}

// stop closes everything.
func (t *Timer) stop(ticker *time.Ticker) {
	ticker.Stop()

	for key := range t.track {
		t.track[key] = nil
		delete(t.track, key)
	}

	t.track = nil
	close(t.ch)
	t.ch = nil
	close(t.rep)
}

// chanWatcher runs a loop, and deletes any keys older than the last cooldown we had for that key.
func (t *Timer) chanWatcher(ticker *time.Ticker) {
	defer t.stop(ticker)

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
		case now := <-ticker.C:
			for key, val := range t.track {
				if now.After(val.Last.Add(val.Dur)) {
					t.track[key] = nil
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

// Sizes returns the tracked item length and the channel queue length.
func (t *Timer) Sizes() (int, int) {
	return len(t.track), len(t.ch)
}
