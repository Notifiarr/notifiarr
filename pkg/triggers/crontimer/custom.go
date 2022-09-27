package crontimer

import (
	"context"
	"encoding/json"
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
	ch      chan *common.ActionInput
}

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Run fires a custom cron timer (GET).
func (t *Timer) Run(input *common.ActionInput) {
	if t.ch == nil {
		return
	}

	t.ch <- input
}

// run responds to the channel that the timer fired into.
func (t *Timer) run(ctx context.Context, input *common.ActionInput) {
	t.website.SendData(&website.Request{
		Route:      website.Route(t.URI),
		Event:      input.Type,
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
			ch:         make(chan *common.ActionInput, 1),
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

// PollForReload checks if the website wants the client to reload (new settings).
func (c *cmd) PollForReload(ctx context.Context, input *common.ActionInput) {
	body, err := c.GetData(&website.Request{
		Route:      website.ClientRoute,
		Event:      website.EventPoll,
		Payload:    c.Server.Info(ctx),
		LogPayload: true,
	})
	if err != nil {
		c.Errorf("[%s requested] Polling Notifiarr: %v", input.Type, err)
		return
	}

	var v struct {
		Reload     bool      `json:"reload"`
		LastSync   time.Time `json:"lastSync"`
		LastChange time.Time `json:"lastChange"`
	}

	if err = json.Unmarshal(body.Details.Response, &v); err != nil {
		c.Errorf("[%s requested] Polling Notifiarr: %v", input.Type, err)
		return
	}

	if v.Reload {
		c.Printf("[%s requested] Website indicated new configurations; reloading to pick them up!"+
			" Last Sync: %v, Last Change: %v, Diff: %v", input.Type, v.LastSync, v.LastChange, v.LastSync.Sub(v.LastChange))
		defer c.ReloadApp("poll triggered reload")
	} else if ci := website.GetClientInfo(); ci == nil {
		c.Printf("[%s requested] API Key checked out, reloading to pick up configuration from website!", input.Type)
		defer c.ReloadApp("client info reload")
	}
}
