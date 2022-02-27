package notifiarr

import (
	"reflect"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"golift.io/cnfg"
)

// hard coded timers.
const (
	// How often to check starr apps for stuck items.
	stuckDur = 5*time.Minute + 1327*time.Millisecond
	// How often to poll the website for changes.
	// This only fires when:
	// 1. the cliet isn't reachable from the website.
	// 2. the client didn't get a valid response to clientInfo.
	pollDur = 4*time.Minute + 977*time.Millisecond
)

// TriggerName makes sure triggers have a known name.
type TriggerName string

// Identified Action Triggers. Name and explanation.
const (
	TrigSnapshot       TriggerName = "Gathering and sending System Snapshot."
	TrigDashboard      TriggerName = "Initiating State Collection for Dashboard."
	TrigCFSync         TriggerName = "Starting Custom Formats and Quality Profiles Sync for Radarr and Sonarr."
	TrigCollectionGaps TriggerName = "Sending Radarr Collection Gaps."
	TrigPlexSessions   TriggerName = "Gathering and sending Plex Sessions."
	TrigStuckItems     TriggerName = "Checking app queues and sending stuck items."
	TrigPollSite       TriggerName = "Polling Notifiarr for new settings."
	TrigStop           TriggerName = "Stopping all triggers and timers (reload)."
)

// timerConfig defines a custom GET timer from the website.
// Used to offload crons to clients.
type timerConfig struct {
	Name     string        `json:"name"`     // name of action.
	Interval cnfg.Duration `json:"interval"` // how often to GET this URI.
	URI      string        `json:"endpoint"` // endpoint for the URI.
	Desc     string        `json:"description"`
	ch       chan EventType
	getdata  func(string) (*Response, error)
	errorf   func(string, ...interface{})
}

type action struct {
	Name TriggerName
	Fn   func(EventType)           // most actions use this
	SFn  func(map[string]struct{}) // this is just for plex sessions.
	C    chan EventType            // if provided, T is optional
	T    *time.Ticker              // if provided, C is optional.
	Hide bool                      // prevent logging.
}

// Run fires a custom cron timer (GET).
func (t *timerConfig) Run(event EventType) {
	if t.ch == nil {
		return
	}

	t.ch <- event
}

// run responds to the channel that the timer fired into.
func (t *timerConfig) run(event EventType) {
	if _, err := t.getdata(t.URI); err != nil {
		t.errorf("[%s requested] Custom Timer Request for %s failed: %v", event, t.URI, err)
	}
}

// setup makes sure a timer has info to do it's job and log results.
func (t *timerConfig) setup(c *Config) {
	t.URI, t.errorf, t.getdata = c.BaseURL+"/"+t.URI, c.Errorf, c.GetData
	t.ch = make(chan EventType, 1)
}

// runTimers converts all the tickers and triggers into []reflect.SelectCase.
// This allows us to run a loop with a dynamic number of channels and tickers to watch.
func (c *Config) runTimers() {
	var (
		cases          = []reflect.SelectCase{}
		combine        = []*action{}
		timer, trigger int
	)

	for _, action := range append(c.Trigger.List, c.Trigger.stop) {
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
func (c *Config) runTimerLoop(actions []*action, cases []reflect.SelectCase) { //nolint:cyclop
	defer c.stopTimerLoop(actions)

	// Sent is used a memory buffer for plex session tracking.
	// This specifically makes sure we don't send a "finished item" more than once.
	sent := make(map[string]struct{})

	// This is how you watch a slice of reflect.SelectCase.
	// This allows watching a dynamic amount of channels and tickers.
	for {
		index, val, _ := reflect.Select(cases)
		action := actions[index]

		var event EventType
		if _, ok := val.Interface().(time.Time); ok {
			event = EventCron
		} else if event, ok = val.Interface().(EventType); !ok {
			event = "unknown"
		}

		exp.TimerEvents.Add(string(event)+"&&"+string(action.Name), 1)

		if action.Fn == nil && action.SFn == nil { // stop channel has no Functions
			return // called by c.Stop(), calls c.stopTimerLoop().
		}

		if event == EventUser && action.Name != "" {
			if err := ui.Notify(string(action.Name)); err != nil {
				c.Errorf("Displaying toast notification: %v", err)
			}
		}

		if action.Name != "" && !action.Hide {
			c.Printf("[%s requested] %s", event, action.Name)
		}

		if action.Fn != nil {
			action.Fn(event)
		} else {
			action.SFn(sent)
		}
	}
}

// stopTimerLoop is defered by runTimerLoop.
// This procedure closes all the timer channels and stops the tickers.
// These cannot be restarted and must be fully initialized again.
func (c *Config) stopTimerLoop(actions []*action) {
	defer c.CapturePanic()
	c.Printf("!!> Stopping main Notifiarr loop. All timers and triggers are now disabled.")
	c.Trigger.stop = nil

	for _, action := range actions {
		if action.C != nil {
			close(action.C)
			action.C = nil
		}

		if action.T != nil {
			action.T.Stop()
			action.T = nil
		}
	}
}

/* Helpers. */

// exec runs a trigger. This is abastraction method used in a bunch of places.
func (t *Triggers) exec(event EventType, name TriggerName) {
	trig := t.get(name)
	if t.stop == nil || trig == nil || trig.C == nil {
		return
	}

	trig.C <- event
}

// get a trigger by unique name. May return nil, and that could cause a panic.
// We avoid panics by using a custom type with corresponding constants as input.
func (t *Triggers) get(name TriggerName) *action {
	for _, trigger := range t.List {
		if trigger.Name == name {
			return trigger
		}
	}

	return nil
}

// add  adds a new action to ou list of "Actions to run."
// actions are timers or triggers, or both.
func (t *Triggers) add(action ...*action) {
	t.List = append(t.List, action...)
}

// tickerOrNil is used in a place or two. It checks it a timer is nil,
// and returns a nil ticker, otherwise it returns the ticker as requested.
func tickerOrNil(duration time.Duration) *time.Ticker {
	if duration == 0 {
		return nil
	}

	return time.NewTicker(duration)
}
