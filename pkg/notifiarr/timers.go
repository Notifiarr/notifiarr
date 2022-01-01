package notifiarr

import (
	"reflect"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/ui"
	"golift.io/cnfg"
)

const (
	stuckDur = 5*time.Minute + 1327*time.Millisecond
	pollDur  = 4*time.Minute + 977*time.Millisecond
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
	TrigStop           TriggerName = "Stop Channel is used for reloads and must not have a function."
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
func (c *Config) runTimerLoop(actions []*action, cases []reflect.SelectCase) {
	defer c.stopTimerLoop(actions)

	for sent := make(map[string]struct{}); ; {
		index, val, _ := reflect.Select(cases)

		action := actions[index]
		if action.Fn == nil && action.SFn == nil { // stop channel has no Functions
			return
		}

		var event EventType
		if _, ok := val.Interface().(time.Time); ok {
			event = EventCron
		} else if event, ok = val.Interface().(EventType); !ok {
			event = "unknown"
		}

		if event == EventUser && action.Name != "" {
			if err := ui.Notify(string(action.Name)); err != nil {
				c.Errorf("Displaying toast notification: %v", err)
			}
		}

		if action.Name != "" {
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
func (c *Config) stopTimerLoop(actions []*action) {
	defer c.CapturePanic()
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

// get a trigger by unique name. May return nil, and that could cause a panic.
func (t *Triggers) get(name TriggerName) *action {
	for _, trigger := range t.List {
		if trigger.Name == name {
			return trigger
		}
	}

	return nil
}

func (t *Triggers) add(action ...*action) {
	t.List = append(t.List, action...)
}

func tickerOrNil(duration time.Duration) *time.Ticker {
	if duration == 0 {
		return nil
	}

	return time.NewTicker(duration)
}
