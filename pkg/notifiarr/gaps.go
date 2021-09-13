package notifiarr

import (
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/cnfg"
	"golift.io/starr/radarr"
)

/* Gaps allows filling gaps in Radarr collections. */

// gapsConfig is the configuration returned from the notifiarr website.
type gapsConfig struct {
	Instances intList       `json:"instances"`
	Interval  cnfg.Duration `json:"interval"`
}

func (t *Triggers) SendGaps(event EventType) {
	if t.stop == nil {
		return
	}

	t.gaps <- event
}

func (c *Config) sendGaps(event EventType) {
	if c.ClientInfo == nil || len(c.Actions.Gaps.Instances) == 0 || len(c.Apps.Radarr) == 0 {
		c.Errorf("Cannot send Radarr Collection Gaps: instances or configured Radarrs (%d) are zero.", len(c.Apps.Radarr))
		return
	}

	c.Printf("Sending Radarr Collections Gaps to Notifiarr: triggered by %s", event)

	for i, r := range c.Apps.Radarr {
		instance := i + 1
		if r.URL == "" || r.APIKey == "" || !c.Actions.Gaps.Instances.Has(instance) {
			continue
		}

		if err := c.sendInstanceGaps(event, instance, r); err != nil {
			c.Errorf("Radarr Collection Gaps request for '%d:%s' failed: %v", instance, r.URL, err)
		} else {
			c.Printf("Sent Collection Gaps to Notifiarr for Radarr: %d:%s", instance, r.URL)
		}
	}
}

func (c *Config) sendInstanceGaps(event EventType, instance int, app *apps.RadarrConfig) error {
	type radarrGapsPayload struct {
		Instance int             `json:"instance"`
		Name     string          `json:"name"`
		Movies   []*radarr.Movie `json:"movies"`
	}

	movies, err := app.GetMovie(0)
	if err != nil {
		return fmt.Errorf("getting movies: %w", err)
	}

	_, err = c.SendData(GapsRoute.Path(event, "app=radarr"), &radarrGapsPayload{
		Movies:   movies,
		Name:     app.Name,
		Instance: instance,
	}, false)
	if err != nil {
		return fmt.Errorf("sending collection gaps: %w", err)
	}

	return nil
}
