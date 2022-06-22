package crontimer

import (
	"fmt"
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
	pollDur = 4*time.Minute + 977*time.Millisecond
)

type Timer struct {
	*website.CronTimer
	website *common.Config
	ch      chan website.EventType
}

type Action struct {
	*common.Config
	list []*Timer
}

// Run fires a custom cron timer (GET).
func (t *Timer) Run(event website.EventType) {
	if t.ch == nil {
		return
	}

	t.ch <- event
}

// run responds to the channel that the timer fired into.
func (t *Timer) run(event website.EventType) {
	payload := struct{ Cron string }{Cron: "thingy"}
	if resp, err := t.website.SendData(t.CronTimer.URI, payload, true); err != nil {
		t.website.Errorf("[%s requested] Custom Timer Request for %s failed: %v%v", event, t.CronTimer.URI, err, resp)
	}
}

func (c *Action) List() []*Timer {
	return c.list
}

func (c *Action) Create() {
	// This poller is sorta shoehorned in here for lack of a better place to put it.
	if c.ClientInfo == nil || c.ClientInfo.Actions.Poll {
		c.Printf("==> Started Notifiarr Poller, have_clientinfo:%v interval:%s",
			c.ClientInfo != nil, cnfg.Duration{Duration: pollDur.Round(time.Second)})
		c.Add(&common.Action{Name: TrigPollSite, Fn: c.PollForReload, T: time.NewTicker(pollDur)})
	}

	if c.ClientInfo == nil {
		return
	}

	for _, custom := range c.ClientInfo.Actions.Custom {
		timer := &Timer{
			CronTimer: custom,
			ch:        make(chan website.EventType, 1),
			website:   c.Config,
		}
		custom.URI = "/" + strings.TrimPrefix(custom.URI, "/")

		var ticker *time.Ticker

		if custom.Interval.Duration < time.Minute {
			c.Errorf("Website provided custom cron interval under 1 minute. Ignored! Interval: %s Name: %s, URI: %s",
				custom.Interval, custom.Name, custom.URI)
		} else {
			ticker = time.NewTicker(custom.Interval.Duration)
		}

		c.list = append(c.list, timer)

		c.Add(&common.Action{
			Name: common.TriggerName(fmt.Sprintf("Running Custom Cron Timer '%s' POST %s", custom.Name, custom.URI)),
			Fn:   timer.run,
			C:    timer.ch,
			T:    ticker,
		})
	}

	c.Printf("==> Custom Timers Enabled: %d timers provided", len(c.ClientInfo.Actions.Custom))
}
