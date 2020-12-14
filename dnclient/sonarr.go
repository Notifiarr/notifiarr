package dnclient

import (
	"sync"

	"golift.io/starr"
)

func (c *Config) fixSonarrConfig() {
	for i := range c.Sonarr {
		if c.Sonarr[i].Timeout.Duration == 0 {
			c.Sonarr[i].Timeout.Duration = c.Timeout.Duration
		}
	}
}

// SonarrConfig represents the input data for a Sonarr server.
type SonarrConfig struct {
	Name string `json:"name" toml:"name" xml:"name" yaml:"name"`
	*starr.Config
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logSonarr() {
	if count := len(c.Sonarr); count == 1 {
		c.Printf(" => Sonarr Config: 1 server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Sonarr[0].Name, c.Sonarr[0].URL, c.Sonarr[0].APIKey != "", c.Sonarr[0].Timeout, c.Sonarr[0].ValidSSL)
	} else {
		c.Print(" => Sonarr Config:", count, "servers")

		for _, f := range c.Sonarr {
			c.Printf(" =>    Server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.Name, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

func (c *Client) handleSonarr(p *IncomingPayload) (string, error) {
	return "", nil
}
