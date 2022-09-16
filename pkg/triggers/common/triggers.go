package common

import (
	"context"
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

// Config is the input data shared by most triggers.
// Everything is mandatory.
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

// Action defines a trigger/timer that can be executed.
type Action struct {
	Name TriggerName
	Fn   func(context.Context, website.EventType) // most actions use this for triggers.
	C    chan website.EventType                   // if provided, T is optional.
	T    *time.Ticker                             // if provided, C is optional.
	Hide bool                                     // prevent logging.
}

// Exec runs a trigger. This is abastraction method used in a bunch of places.
func (c *Config) Exec(event website.EventType, name TriggerName) bool {
	trig := c.Get(name)
	if c.stop == nil || trig == nil || trig.C == nil {
		return false
	}

	trig.C <- event

	return true
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

// Stop shuts down the loop/goroutine that handles all triggers and timers.
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

// WithInstance returns a trigger name with an instance ID.
func (name TriggerName) WithInstance(instance int) TriggerName {
	return TriggerName(fmt.Sprintf(string(name), instance))
}
