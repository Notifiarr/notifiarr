// Package share is here so we can keep website cruft out of the logs package.
package share

import (
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

type Website interface {
	SendData(req *website.Request)
}

const shareCooldown = time.Hour

var (
	// Config is setup by the configfile package.
	enabled bool
	locker  sync.RWMutex
	dedupe  = newDeduper(shareCooldown)
)

func Enable() {
	locker.Lock()
	defer locker.Unlock()

	enabled = true
}

func Disable() {
	locker.Lock()
	defer locker.Unlock()

	enabled = false
}

// Match is what we send to the website.
type Match struct {
	File    string   `json:"file"`
	Matches []string `json:"matches"`
	Line    string   `json:"line"`
}

// Log sends an error message to the website.
// Identical messages are suppressed for shareCooldown; the next send notes how many were skipped.
func Log(reqID string, msg string) {
	locker.RLock()
	defer locker.RUnlock()

	if ci := data.Get("clientInfo"); ci == nil || !enabled {
		return
	}

	line, ok := dedupe.take(msg)
	if !ok {
		return
	}

	website.SendData(&website.Request{
		ReqID:      reqID,
		Payload:    &Match{File: "client_error_log", Line: line, Matches: []string{"[ERROR]"}},
		Route:      website.LogLineRoute,
		Event:      website.EventFile,
		LogPayload: true,
	})
}

type deduper struct {
	mu      sync.Mutex
	cool    time.Duration
	now     func() time.Time
	last    map[string]time.Time
	skipped map[string]int
}

func newDeduper(cool time.Duration) *deduper {
	return &deduper{
		cool:    cool,
		now:     time.Now,
		last:    make(map[string]time.Time),
		skipped: make(map[string]int),
	}
}

// take returns the (possibly annotated) message to send.
// The second return is false when this message is still on cooldown.
func (d *deduper) take(msg string) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	d.prune(now)

	if last, ok := d.last[msg]; ok && now.Sub(last) < d.cool {
		d.skipped[msg]++
		return "", false
	}

	n := d.skipped[msg]
	d.skipped[msg] = 0
	d.last[msg] = now

	if n > 0 {
		return fmt.Sprintf("%s (repeated %d times)", msg, n), true
	}

	return msg, true
}

func (d *deduper) prune(now time.Time) {
	for key, last := range d.last {
		if now.Sub(last) < d.cool || d.skipped[key] > 0 {
			continue
		}

		delete(d.last, key)
		delete(d.skipped, key)
	}
}
