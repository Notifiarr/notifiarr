package common

import (
	"context"
	"reflect"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/cnfg"
)

// TrigStop is used to signal a stop/reload.
const TrigStop TriggerName = "Stopping all triggers and timers (reload)."

// Run converts all the tickers and triggers into []reflect.SelectCase.
// This allows us to run a loop with a dynamic number of channels and tickers to watch.
func (c *Config) Run(ctx context.Context) {
	if c.stop != nil {
		panic("notifiarr timers cannot run more than once")
	}

	c.stop = &Action{Name: TrigStop, C: make(chan *ActionInput)}

	var (
		cases   = []reflect.SelectCase{}
		combine = []*Action{}
	)

	for _, action := range append(c.list, c.stop) {
		if action == nil {
			continue
		}

		// Since we may add up to 2 actions per list item, duplicate the pointer in a new combined list.
		if action.C != nil {
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(action.C)})
			combine = append(combine, action)
		}

		if action.t != nil {
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(action.t.C)})
			combine = append(combine, action)
		}
	}

	go c.runTimerLoop(ctx, combine, cases)
	c.printStartupLog()
}

func (c *Config) GatherTriggerInfo() (map[string]cnfg.Duration, map[string]cnfg.Duration) {
	var (
		triggers = make(map[string]cnfg.Duration)
		timers   = make(map[string]cnfg.Duration)
	)

	for _, action := range append(c.list, c.stop) {
		if action == nil {
			continue
		}

		if action.C != nil {
			triggers[string(action.Name)] = action.D
		}

		if action.t != nil {
			timers[string(action.Name)] = action.D
		}
	}

	return triggers, timers
}

func (c *Config) printStartupLog() {
	triggers, timers := c.GatherTriggerInfo()
	c.Printf("==> Actions Started: %d Timers and %d Triggers", len(timers), len(triggers))

	for name, dur := range triggers {
		if _, ok := timers[name]; ok {
			c.Debugf("==> Enabled Action: %s Trigger and Timer, interval: %s", name, dur)
		} else {
			c.Debugf("==> Enabled Action: %s Trigger only.", name)
		}
	}

	for name, dur := range timers {
		if _, ok := triggers[name]; !ok {
			c.Debugf("==> Enabled Action: %s Timer only, interval: %s", name, dur)
		}
	}
}

// runTimerLoop does all of the timer/cron routines for starr apps and plex.
// Many of the menu items and trigger handlers feed into this routine too.
// All of the actions this library runs are contained in this one go routine.
// That means only 1 action can run at a time. If c.Serial is set to true, then
// some of those actions (especially dashboard) will spawn their own go routines.
func (c *Config) runTimerLoop(ctx context.Context, actions []*Action, cases []reflect.SelectCase) {
	defer func() {
		c.CapturePanic()
		c.stopTimerLoop(actions)
	}()

	// This is how you watch a slice of reflect.SelectCase.
	// This allows watching a dynamic amount of channels and tickers.
	for {
		index, val, _ := reflect.Select(cases)
		action := actions[index]

		input := &ActionInput{}

		if _, ok := val.Interface().(time.Time); ok {
			input.Type = website.EventCron
		} else if input, ok = val.Interface().(*ActionInput); !ok {
			input.Type = "unknown"
		}

		mnd.TimerEvents.Add(string(input.Type)+"&&"+string(action.Name), 1)
		mnd.TimerCounts.Add(string(action.Name), 1)

		if action.Fn == nil { // stop channel has no Function.
			return // called by c.Stop(), calls c.stopTimerLoop().
		}

		c.runEventAction(ctx, input, action)
	}
}

func (c *Config) runEventAction(ctx context.Context, input *ActionInput, action *Action) {
	if input.Type == website.EventUser && action.Name != "" {
		if err := ui.Notify(string(action.Name)); err != nil {
			c.Errorf("Displaying toast notification: %v", err)
		}
	}

	if action.Name != "" && !action.Hide {
		c.Printf("[%s requested] Event Triggered: %s", input.Type, action.Name)
	}

	if action.Fn != nil {
		action.Fn(ctx, input)
	}
}

// stopTimerLoop is deferred by runTimerLoop.
// This procedure closes all the timer channels and stops the tickers.
// These cannot be restarted and must be fully initialized again.
func (c *Config) stopTimerLoop(actions []*Action) {
	defer close(c.stop.C) // signal that we're done.

	c.Printf("!!> Stopping main Notifiarr loop. All timers and triggers are now disabled.")

	for _, action := range actions {
		if action.t != nil {
			action.t.Stop()
			action.t = nil
		}

		if action.C != nil && action.C != c.stop.C { // do not close stop channel here.
			close(action.C)
			action.C = nil
		}
	}
}
