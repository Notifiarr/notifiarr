package notifiarr

import (
	"fmt"

	"golift.io/cnfg"
	"golift.io/starr/radarr"
)

/* Gaps allows filling gaps in Radarr collections. */

// gapsConfig is the configuration returned from the notifiarr website.
type gapsConfig struct {
	Instances IntList       `json:"instances"`
	Interval  cnfg.Duration `json:"interval"`
}

func (t *Triggers) SendGaps(event EventType) {
	t.exec(event, TrigCollectionGaps)
}

func (c *Config) sendGaps(event EventType) {
	if c.clientInfo == nil || len(c.clientInfo.Actions.Gaps.Instances) == 0 || len(c.Apps.Radarr) == 0 {
		c.Errorf("[%s requested] Cannot send Radarr Collection Gaps: instances or configured Radarrs (%d) are zero.",
			event, len(c.Apps.Radarr))
		return
	}

	for idx, app := range c.Apps.Radarr {
		instance := idx + 1
		if app.URL == "" || app.APIKey == "" || !c.clientInfo.Actions.Gaps.Instances.Has(instance) {
			continue
		}

		type radarrGapsPayload struct {
			Instance int             `json:"instance"`
			Name     string          `json:"name"`
			Movies   []*radarr.Movie `json:"movies"`
		}

		movies, err := app.GetMovie(0)
		if err != nil {
			c.Errorf("[%s requested] Radarr Collection Gaps (%d:%s) failed: getting movies: %v", event, instance, app.URL, err)
			continue
		}

		c.QueueData(&SendRequest{
			Route:      DashRoute,
			Event:      event,
			LogPayload: true,
			LogMsg:     fmt.Sprintf("Radarr Collection Gaps (%d:%s)", instance, app.URL),
			Payload:    &radarrGapsPayload{Movies: movies, Name: app.Name, Instance: instance},
		})
	}
}
