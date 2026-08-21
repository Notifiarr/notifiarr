// Package share is here so we can keep website cruft out of the logs package.
package share

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

type Website interface {
	SendData(req *website.Request)
}

const (
	shareCooldown = time.Hour
	// Keep skip counts this many cooldowns so the next send can annotate, then drop them.
	skipRetain = 2
	traceOpen  = "{trace:"
	traceClose = "} "
)

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

// shareKey strips a {trace:reqID} prefix so the same error dedupes across requests.
func shareKey(msg string) string {
	if !strings.HasPrefix(msg, traceOpen) {
		return msg
	}

	idx := strings.Index(msg, traceClose)
	if idx < 0 {
		return msg
	}

	return msg[idx+len(traceClose):]
}

func repeatedSuffix(n int) string {
	if n == 1 {
		return fmt.Sprintf(" (repeated %d time)", n)
	}

	return fmt.Sprintf(" (repeated %d times)", n)
}

// take returns the (possibly annotated) message to send.
// The second return is false when this message is still on cooldown.
func (d *deduper) take(msg string) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	d.prune(now)

	key := shareKey(msg)

	if last, ok := d.last[key]; ok && now.Sub(last) < d.cool {
		d.skipped[key]++
		return "", false
	}

	n := d.skipped[key]
	d.skipped[key] = 0
	d.last[key] = now

	if n > 0 {
		return msg + repeatedSuffix(n), true
	}

	return msg, true
}

func (d *deduper) prune(now time.Time) {
	for key, last := range d.last {
		limit := d.cool
		if d.skipped[key] > 0 {
			limit *= skipRetain
		}

		if now.Sub(last) < limit {
			continue
		}

		delete(d.last, key)
		delete(d.skipped, key)
	}
}
