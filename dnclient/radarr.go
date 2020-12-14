package dnclient

import (
	"sync"

	"golift.io/starr"
)

// RadarrConfig represents the input data for a Radarr server.
type RadarrConfig struct {
	Name string `json:"name" toml:"name" xml:"name" yaml:"name"`
	*starr.Config
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logRadarr() {
	if count := len(c.Radarr); count == 1 {
		c.Printf(" => Radarr Config: 1 server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Radarr[0].Name, c.Radarr[0].URL, c.Radarr[0].APIKey != "", c.Radarr[0].Timeout, c.Radarr[0].ValidSSL)
	} else {
		c.Print(" => Radarr Config:", count, "servers")

		for _, f := range c.Radarr {
			c.Printf(" =>    Server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.Name, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}
