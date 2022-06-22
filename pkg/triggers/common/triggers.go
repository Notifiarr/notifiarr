package common

import (
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// ErrInvalidApp is returned by triggers when a non-existent app is requested.
var ErrInvalidApp = fmt.Errorf("invalid application provided")

// ErrNoChannel is returned when the go routine is stopped.
var ErrNoChannel = fmt.Errorf("no channel to send request")

// Config is the input data needed to send payloads to notifiarr.
type Config struct {
	*website.Server // send trigger responses to website.
	*website.ClientInfo
	Snapshot *snapshot.Config
	Apps     *apps.Apps
	Serial   bool
	mnd.Logger
	triggers // add triggers.
}

// TriggerName makes sure triggers have a known name.
type TriggerName string

type Action struct {
	Name TriggerName
	Fn   func(website.EventType)   // most actions use this
	SFn  func(map[string]struct{}) // this is just for plex sessions.
	C    chan website.EventType    // if provided, T is optional
	T    *time.Ticker              // if provided, C is optional.
	Hide bool                      // prevent logging.
}

// triggers allow trigger actions in the timer routine.
type triggers struct {
	stop *Action   // Triggered by calling Stop()
	list []*Action // List of action triggers
}

// Exec runs a trigger. This is abastraction method used in a bunch of places.
func (t *triggers) Exec(event website.EventType, name TriggerName) {
	trig := t.Get(name)
	if t.stop == nil || trig == nil || trig.C == nil {
		return
	}

	trig.C <- event
}

// Get a trigger by unique name. May return nil, and that could cause a panic.
// We avoid panics by using a custom type with corresponding constants as input.
func (t *triggers) Get(name TriggerName) *Action {
	for _, trigger := range t.list {
		if trigger.Name == name {
			return trigger
		}
	}

	return nil
}

// Add adds a new action to ou list of "Actions to run."
// actions are timers or triggers, or both.
func (t *triggers) Add(action ...*Action) {
	t.list = append(t.list, action...)
}

func (t *triggers) Close(event website.EventType) {
	// Neither of these if statements should ever fire. That's a bug somewhere else.
	if t == nil {
		panic("Config is nil, cannot stop a nil config!!")
	}

	if t.stop == nil {
		panic("Notifiarr Timers cannot be stopped: not running!!")
	}

	t.stop.C <- event
	<-t.stop.C // wait for done signal.
	t.stop = nil
}

// TickerOrNil is used in a place or two. It checks it a timer is nil,
// and returns a nil ticker, otherwise it returns the ticker as requested.
func TickerOrNil(duration time.Duration) *time.Ticker {
	if duration == 0 {
		return nil
	}

	return time.NewTicker(duration)
}
