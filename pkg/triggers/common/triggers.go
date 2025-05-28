package common

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/go-co-op/gocron/v2"
	"golift.io/cnfg"
)

// ErrInvalidApp is returned by triggers when a non-existent app is requested.
var ErrInvalidApp = errors.New("invalid application provided")

// ErrNoChannel is returned when the go routine is stopped.
var ErrNoChannel = errors.New("no channel to send request")

// Config is the input data shared by most triggers.
// Everything is mandatory.
type Config struct {
	CI              *clientinfo.Config
	*website.Server // send trigger responses to website.
	Snapshot        *snapshot.Config
	Apps            *apps.Apps
	Scheduler       gocron.Scheduler
	*logs.Logger
	stop     *Action        // Triggered by calling Stop()
	list     []*Action      // List of action triggers
	Services                // for running service checks.
	reloadCh chan os.Signal // so triggers can reload the app.
	stopCh   chan os.Signal // so triggers can stop the app.
	rand     *rand.Rand
}

type Create interface {
	Create()
}

type Run interface {
	Run(ctx context.Context)
	Stop()
}

// SetReloadCh is used to set the reload channel for triggers.
// This is an exported method because the channel is not always
// available when triggers are initialized.
func (c *Config) SetReloadCh(sighup chan os.Signal) {
	c.reloadCh = sighup

	if c.rand == nil {
		c.rand = rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec
	}
}

// ReloadApp reloads the application configuration.
func (c *Config) ReloadApp(reason string) {
	if c.reloadCh == nil {
		panic("attempt to reload with no reload channel")
	}

	c.reloadCh <- &update.Signal{Text: reason}
}

// SetReloadCh is used to set the stop channel for triggers.
// This is an exported method because the channel is not always
// available when triggers are initialized.
func (c *Config) SetStopCh(stopCh chan os.Signal) {
	c.stopCh = stopCh
}

// StopApp shuts down the application.
func (c *Config) StopApp(reason string) {
	if c.stopCh == nil {
		panic("attempt to stop with no stop channel")
	}

	c.stopCh <- &update.Signal{Text: reason}
}

// ActionInput is used to send data to a trigger action.
type ActionInput struct {
	Type website.EventType
	Args []string
}

// TriggerName makes sure triggers have a known name.
type TriggerName string

// Action defines a trigger/timer that can be executed.
type Action struct {
	Name TriggerName
	D    cnfg.Duration                       // how often the timer fires, sets ticker.
	J    *scheduler.CronJob                  // If provided, D is ignored.
	Fn   func(context.Context, *ActionInput) // most actions use this for triggers.
	C    chan *ActionInput                   // if provided, D is optional.
	t    *time.Ticker                        // if provided, C is optional.
	job  gocron.Job                          // created if J is non-nil .
	Hide bool                                // prevent logging.
}

// Services is the input interface to do things with services via triggers.
type Services interface {
	RunChecks(et website.EventType)
}

// Exec runs a trigger. This is abstraction method is used in a bunch of places.
func (c *Config) Exec(input *ActionInput, name TriggerName) bool {
	trig := c.Get(name)
	if c.stop == nil || trig == nil || trig.C == nil {
		return false
	}

	trig.C <- input

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
	for _, a := range action {
		if a.J != nil {
			a.job = a.J.New(c.Scheduler, func() { a.C <- &ActionInput{Type: website.EventCron} })
		} else if a.D.Duration != 0 {
			a.t = time.NewTicker(a.D.Duration)
		}
	}

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

	c.stop.C <- &ActionInput{Type: event}
	<-c.stop.C // wait for done signal.
	c.stop = nil
}

// Rand returns a cryptographically-insecure random number generator.
func (c *Config) Rand() *rand.Rand {
	return c.rand
}

// WithInstance returns a trigger name with an instance ID.
func (name TriggerName) WithInstance(instance int) TriggerName {
	return TriggerName(fmt.Sprint(name, instance))
}
