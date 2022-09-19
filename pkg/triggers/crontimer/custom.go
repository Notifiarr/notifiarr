package crontimer

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/cnfg"
)

// TrigPollSite is our site polling trigger identifier.
const TrigPollSite common.TriggerName = "Polling Notifiarr for new settings."

const (
	// How often to poll the website for changes.
	// This only fires when:
	// 1. the cliet isn't reachable from the website.
	// 2. the client didn't get a valid response to clientInfo.
	pollDur            = 4 * time.Minute
	randomMilliseconds = 5000
	randomSeconds      = 30
)

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	list []*Timer
}

// Timer is used to trigger actions.
type Timer struct {
	*website.CronConfig
	website *website.Server
	ch      chan website.EventType
}

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Run fires a custom cron timer (GET).
func (t *Timer) Run(event website.EventType) {
	if t.ch == nil {
		return
	}

	t.ch <- event
}

// run responds to the channel that the timer fired into.
func (t *Timer) run(ctx context.Context, event website.EventType) {
	t.website.SendData(&website.Request{
		Route:      website.Route(t.URI),
		Event:      event,
		Payload:    &struct{ Cron string }{Cron: "thingy"},
		LogMsg:     "Custom Timer Request '" + t.Name + "'",
		LogPayload: true,
	})
}

// List returns a list of active triggers that can be executed.
func (a *Action) List() []*Timer {
	return a.cmd.list
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.create()
}

func (c *cmd) create() {
	ci := website.GetClientInfo()
	// This poller is sorta shoehorned in here for lack of a better place to put it.
	if ci == nil || ci.Actions.Poll {
		c.Printf("==> Started Notifiarr Poller, have_clientinfo:%v interval:%s",
			ci != nil, cnfg.Duration{Duration: pollDur.Round(time.Second)})
		c.Add(&common.Action{
			Name: TrigPollSite,
			Fn:   c.PollForReload,
			T:    time.NewTicker(pollDur + time.Duration(rand.Intn(randomSeconds))*time.Second), //nolint:gosec
		})
	}

	if ci == nil {
		return
	}

	for _, custom := range ci.Actions.Custom {
		timer := &Timer{
			CronConfig: custom,
			ch:         make(chan website.EventType, 1),
			website:    c.Config.Server,
		}
		custom.URI = "/" + strings.TrimPrefix(custom.URI, "/")

		var ticker *time.Ticker

		if custom.Interval.Duration < time.Minute {
			c.Errorf("Website provided custom cron interval under 1 minute. Ignored! Interval: %s Name: %s, URI: %s",
				custom.Interval, custom.Name, custom.URI)
		} else {
			randomTime := time.Duration(rand.Intn(randomMilliseconds)) * time.Millisecond //nolint:gosec
			ticker = time.NewTicker(custom.Interval.Duration + randomTime)
		}

		c.list = append(c.list, timer)

		c.Add(&common.Action{
			Name: common.TriggerName(fmt.Sprintf("Running Custom Cron Timer '%s'", custom.Name)),
			Fn:   timer.run,
			C:    timer.ch,
			T:    ticker,
		})
	}

	c.Printf("==> Custom Timers Enabled: %d timers provided", len(ci.Actions.Custom))
}
