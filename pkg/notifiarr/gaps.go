package notifiarr

import (
	"fmt"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"golift.io/starr/radarr"
)

/* Gaps allows filling gaps in Radarr collections. */

// gapsConfig is the configuration returned from the notifiarr website.
type gapsConfig struct {
	Instances intList `json:"instances"`
	Minutes   int     `json:"timer"`
}

func (t *Triggers) SendGaps(source string) {
	if t.stop == nil {
		return
	}

	t.gaps <- source
}

func (c *Config) sendGaps(source string) {
	ci, err := c.GetClientInfo()
	if err != nil {
		c.Errorf("Cannot send Radarr Collection Gaps: %v", err)
		return
	} else if len(ci.Actions.Gaps.Instances) == 0 || len(c.Apps.Radarr) == 0 {
		c.Errorf("Cannot send Radarr Collection Gaps: instances (%d) or radarrs (%d) are zero.",
			len(ci.Actions.Gaps.Instances), len(c.Apps.Radarr))
		return
	}

	c.Printf("Sending Radarr Collections Gaps to Notifiarr: triggered by %s", source)

	for i, r := range c.Apps.Radarr {
		instance := i + 1
		if r.URL == "" || r.APIKey == "" || !ci.Actions.Gaps.Instances.Has(instance) {
			continue
		}

		if err := c.sendInstanceGaps(instance, r); err != nil {
			c.Errorf("Radarr Collection Gaps request for '%d:%s' failed: %v", instance, r.URL, err)
		} else {
			c.Printf("Sent Collection Gaps to Notifiarr for Radarr: %d:%s", instance, r.URL)
		}
	}
}

func (c *Config) sendInstanceGaps(instance int, app *apps.RadarrConfig) error {
	type radarrGapsPayload struct {
		Instance int             `json:"instance"`
		Name     string          `json:"name"`
		Movies   []*radarr.Movie `json:"movies"`
	}

	movies, err := app.GetMovie(0)
	if err != nil {
		return fmt.Errorf("getting movies: %w", err)
	}

	//nolint:bodyclose // already closed
	resp, _, err := c.SendData(c.BaseURL+GapsRoute+"?app=radarr", &radarrGapsPayload{
		Movies:   movies,
		Name:     app.Name,
		Instance: instance,
	}, false)
	if err != nil {
		return fmt.Errorf("sending collection gaps: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrNon200, resp.Status)
	}

	return nil
}
