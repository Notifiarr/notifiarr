package crontimer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/hako/durafmt"
	"golift.io/cnfg"
)

// TrigPollSite is our site polling trigger identifier.
const (
	TrigPollSite common.TriggerName = "Polling Notifiarr for new settings."
	TrigUpCheck  common.TriggerName = "Telling Notifiarr website we are still up!"
)

const (
	// How often to poll the website for changes.
	// This only fires when:
	// 1. the client isn't reachable from the website.
	// 2. the client didn't get a valid response to clientInfo.
	pollDur = 4 * time.Minute
	// This just tells the website the client is up.
	upCheckDur = 14*time.Minute + 57*time.Second
	// How long to be up before sending first up check.
	checkWait          = 2 * time.Minute
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
	*clientinfo.CronConfig
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
func (t *Timer) run(_ context.Context, input *common.ActionInput) {
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

// Stop satisifies an interface.
func (a *Action) Stop() {}

// Verify the interfaces are satisfied.
var (
	_ = common.Run(&Action{nil})
	_ = common.Create(&Action{nil})
)

// Run files after create to take some immediate action.
func (a *Action) Run(ctx context.Context) {
	a.cmd.PollUpCheck(ctx, &common.ActionInput{Type: website.EventStart})
}

func (c *cmd) create() {
	ci := clientinfo.Get()
	// This poller is sorta shoehorned in here for lack of a better place to put it.
	if ci == nil {
		c.startWebsitePoller()
		return
	}

	c.Printf("==> Started Notifiarr Website Up-Checker, interval: %s", durafmt.Parse(upCheckDur))
	c.Add(&common.Action{
		Name: TrigUpCheck,
		Fn:   c.PollUpCheck,
		D:    cnfg.Duration{Duration: upCheckDur},
	})

	for _, custom := range ci.Actions.Custom {
		timer := &Timer{
			CronConfig: custom,
			ch:         make(chan *common.ActionInput, 1),
			website:    c.Config.Server,
		}
		custom.URI = "/" + strings.TrimPrefix(custom.URI, "/")

		if custom.Interval.Duration < time.Minute {
			c.Errorf("Website provided custom cron interval under 1 minute. Interval: %s Name: %s, URI: %s",
				custom.Interval, custom.Name, custom.URI)

			custom.Interval.Duration = time.Minute
		}

		c.list = append(c.list, timer)

		c.Add(&common.Action{
			Name: common.TriggerName(fmt.Sprintf("Running Custom Cron Timer '%s'", custom.Name)),
			Fn:   timer.run,
			C:    timer.ch,
			D:    cnfg.Duration{Duration: custom.Interval.Duration},
		})
	}

	c.Printf("==> Custom Timers Enabled: %d timers provided", len(ci.Actions.Custom))
}

func (c *cmd) startWebsitePoller() {
	if c.ValidAPIKey() != nil {
		return // only poll if the api key length is valid.
	}

	c.Printf("==> Started Notifiarr Website Poller, interval: %s", durafmt.Parse(pollDur))
	c.Add(&common.Action{
		Name: TrigPollSite,
		Fn:   c.PollForReload,
		D:    cnfg.Duration{Duration: pollDur + time.Duration(c.Config.Rand().Intn(randomSeconds))*time.Second},
	})
}

func (c *cmd) PollUpCheck(ctx context.Context, input *common.ActionInput) {
	_, err := c.GetData(&website.Request{
		Route:      website.ClientRoute,
		Event:      website.EventCheck,
		Payload:    c.CIC.Info(ctx, true), // true avoids polling tautulli.
		LogPayload: true,
		ErrorsOnly: true,
		LogMsg:     string(TrigUpCheck),
	})
	if err != nil {
		c.Errorf("[%s requested] Polling Notifiarr: %v", input.Type, err)
		return
	}
}

// PollForReload is only started if the initial connection to the website failed.
// This will keep checking until it works, then reload to grab settings and start properly.
func (c *cmd) PollForReload(ctx context.Context, input *common.ActionInput) {
	body, err := c.GetData(&website.Request{
		Route:      website.ClientRoute,
		Event:      website.EventPoll,
		Payload:    c.CIC.Info(ctx, true), // true avoids polling tautulli.
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

	if ci := clientinfo.Get(); ci == nil {
		c.Printf("[%s requested] API Key checked out, reloading to pick up configuration from website!", input.Type)
		defer c.ReloadApp("client info reload")
	}
}
