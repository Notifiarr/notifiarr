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
	stop *Action   // Triggered by calling Stop()
	list []*Action // List of action triggers
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

// Exec runs a trigger. This is abastraction method used in a bunch of places.
func (c *Config) Exec(event website.EventType, name TriggerName) {
	trig := c.Get(name)
	if c.stop == nil || trig == nil || trig.C == nil {
		return
	}

	trig.C <- event
}

// Get a trigger by unique name. May return nil, and that could cause a panic.
// We avoid panics by using a custom type with corresponding constants as input.
func (c *Config) Get(name TriggerName) *Action {
	for _, trigger := range c.list {
		if trigger.Name == name {
			return trigger
		}
	}

	return nil
}

// Add adds a new action to our list of "Actions to run."
// actions are timers or triggers, or both.
func (c *Config) Add(action ...*Action) {
	c.list = append(c.list, action...)
}

func (c *Config) Stop(event website.EventType) {
	// Neither of these if statements should ever fire. That's a bug somewhere else.
	if c == nil {
		panic("Config is nil, cannot stop a nil config!!")
	}

	if c.stop == nil {
		panic("Notifiarr Timers cannot be stopped: not running!!")
	}

	c.stop.C <- event
	<-c.stop.C // wait for done signal.
	c.stop = nil
}
