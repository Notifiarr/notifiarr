package gaps

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
	"golift.io/starr/radarr"
)

/* Gaps allows filling gaps in Radarr collections. */

const TrigCollectionGaps common.TriggerName = "Sending Radarr Collection Gaps."

const randomMilliseconds = 5000

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
}

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.create()
}

func (c *cmd) create() {
	var dur time.Duration

	ci := clientinfo.Get()
	if ci != nil && ci.Actions.Gaps.Interval.Duration > 0 && len(ci.Actions.Gaps.Instances) > 0 {
		randomTime := time.Duration(c.Config.Rand().Intn(randomMilliseconds)) * time.Millisecond
		dur = ci.Actions.Gaps.Interval.Duration + randomTime
		c.Printf("==> Collection Gaps Timer Enabled, interval:%s", ci.Actions.Gaps.Interval)
	}

	c.Add(&common.Action{
		Name: TrigCollectionGaps,
		Fn:   c.sendGaps,
		C:    make(chan *common.ActionInput, 1),
		D:    cnfg.Duration{Duration: dur},
	})
}

// Send radarr collection gaps to the website.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(&common.ActionInput{Type: event}, TrigCollectionGaps)
}

func (c *cmd) sendGaps(ctx context.Context, input *common.ActionInput) {
	info := clientinfo.Get()
	if info == nil || len(info.Actions.Gaps.Instances) == 0 || len(c.Apps.Radarr) == 0 {
		c.Errorf("[%s requested] Cannot send Radarr Collection Gaps: instances or configured Radarrs (%d) are zero.",
			input.Type, len(c.Apps.Radarr))
		return
	}

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if !app.Enabled() || !info.Actions.Gaps.Instances.Has(instance) {
			continue
		}

		type radarrGapsPayload struct {
			Instance int             `json:"instance"`
			Name     string          `json:"name"`
			Movies   []*radarr.Movie `json:"movies"`
		}

		movies, err := app.GetMovieContext(ctx, &radarr.GetMovie{ExcludeLocalCovers: true})
		if err != nil {
			c.Errorf("[%s requested] Radarr Collection Gaps (%d:%s) failed: getting movies: %v",
				input.Type, instance, app.URL, err)
			continue
		}

		c.SendData(&website.Request{
			Route:      website.GapsRoute,
			Event:      input.Type,
			LogPayload: true,
			LogMsg:     fmt.Sprintf("Radarr Collection Gaps (%d:%s)", instance, app.URL),
			Payload:    &radarrGapsPayload{Movies: movies, Name: app.Name, Instance: instance},
		})
	}
}
