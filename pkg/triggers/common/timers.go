package common

import (
	"reflect"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// TrigStop is used to signal a stop/reload.
const TrigStop TriggerName = "Stopping all triggers and timers (reload)."

// Run converts all the tickers and triggers into []reflect.SelectCase.
// This allows us to run a loop with a dynamic number of channels and tickers to watch.
func (c *Config) Run() {
	if c.stop != nil {
		panic("notifiarr timers cannot run more than once")
	}

	c.stop = &Action{Name: TrigStop, C: make(chan website.EventType)}

	var (
		cases          = []reflect.SelectCase{}
		combine        = []*Action{}
		timer, trigger int
	)

	for _, action := range append(c.list, c.stop) {
		if action == nil {
			continue
		}

		// Since we may add up to 2 actions per list item, duplicate the pointer in a new combined list.
		if action.C != nil {
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(action.C)})
			combine = append(combine, action)
			trigger++
		}

		if action.T != nil {
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(action.T.C)})
			combine = append(combine, action)
			timer++
		}
	}

	go c.runTimerLoop(combine, cases)
	c.Printf("==> Started %d Notifiarr Timers with %d Triggers", timer, trigger)
}

// runTimerLoop does all of the timer/cron routines for starr apps and plex.
// Many of the menu items and trigger handlers feed into this routine too.
// All of the actions this library runs are contained in this one go routine.
// That means only 1 action can run at a time. If c.Serial is set to true, then
// some of those actions (especially dashboard) will spawn their own go routines.
func (c *Config) runTimerLoop(actions []*Action, cases []reflect.SelectCase) { //nolint:cyclop
	defer c.stopTimerLoop(actions)

	// This is how you watch a slice of reflect.SelectCase.
	// This allows watching a dynamic amount of channels and tickers.
	for {
		index, val, _ := reflect.Select(cases)
		action := actions[index]

		var event website.EventType
		if _, ok := val.Interface().(time.Time); ok {
			event = website.EventCron
		} else if event, ok = val.Interface().(website.EventType); !ok {
			event = "unknown"
		}

		exp.TimerEvents.Add(string(event)+"&&"+string(action.Name), 1)
		exp.TimerCounts.Add(string(action.Name), 1)

		if action.Fn == nil { // stop channel has no Function.
			return // called by c.Stop(), calls c.stopTimerLoop().
		}

		if event == website.EventUser && action.Name != "" {
			if err := ui.Notify(string(action.Name)); err != nil {
				c.Errorf("Displaying toast notification: %v", err)
			}
		}

		if action.Name != "" && !action.Hide {
			c.Printf("[%s requested] %s", event, action.Name)
		}

		if action.Fn != nil {
			action.Fn(event)
		}

	}
}

// stopTimerLoop is defered by runTimerLoop.
// This procedure closes all the timer channels and stops the tickers.
// These cannot be restarted and must be fully initialized again.
func (c *Config) stopTimerLoop(actions []*Action) {
	defer func() {
		defer c.CapturePanic()
		close(c.stop.C) // signal that we're done.
	}()

	c.Printf("!!> Stopping main Notifiarr loop. All timers and triggers are now disabled.")

	for _, action := range actions {
		if action.C != nil && action.C != c.stop.C { // do not close stop channel here.
			close(action.C)
			action.C = nil
		}

		if action.T != nil {
			action.T.Stop()
			action.T = nil
		}
	}

}
