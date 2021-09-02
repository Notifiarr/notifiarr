package notifiarr

import (
	"fmt"
	"net/http"

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
	c.Printf("Sending Radarr Collections Gaps to Notifiarr: %s", source)

	ci, err := c.GetClientInfo()
	if err != nil {
		c.Errorf("Cannot send Radarr Collection Gaps: %v", err)
		return
	} else if len(ci.Actions.Gaps.Instances) == 0 {
		return
	}

	for i, r := range c.Apps.Radarr {
		if r.DisableCF || r.URL == "" || r.APIKey == "" || !ci.Actions.Gaps.Instances.Has(i) {
			continue
		}

		if err := c.sendInstanceGaps(i); err != nil {
			c.Errorf("Radarr Collection Gaps request for '%d:%s' failed: %v", i+1, r.URL, err)
		} else {
			c.Printf("Sent Collection Gaps to Notifiarr for Radarr: %d:%s", i+1, r.URL)
		}
	}
}

func (c *Config) sendInstanceGaps(i int) error {
	type radarrGapsPayload struct {
		Instance int             `json:"instance"`
		Name     string          `json:"name"`
		Movies   []*radarr.Movie `json:"movies"`
	}

	movies, err := c.Apps.Radarr[i].GetMovie(0)
	if err != nil {
		return fmt.Errorf("getting movies: %w", err)
	}

	//nolint:bodyclose // already closed
	resp, _, err := c.SendData(c.BaseURL+GapsRoute+"?app=radarr", &radarrGapsPayload{
		Movies:   movies,
		Name:     c.Apps.Radarr[i].Name,
		Instance: i,
	}, false)
	if err != nil {
		return fmt.Errorf("sending collection gaps: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrNon200, resp.Status)
	}

	return nil
}
