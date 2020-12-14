package dnclient

import (
	"sync"

	"golift.io/starr"
)

// LidarrConfig represents the input data for a Lidarr server.
type LidarrConfig struct {
	Name string `json:"name" toml:"name" xml:"name" yaml:"name"`
	*starr.Config
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logLidarr() {
	if count := len(c.Lidarr); count == 1 {
		c.Printf(" => Lidarr Config: 1 server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v,",
			c.Lidarr[0].Name, c.Lidarr[0].URL, c.Lidarr[0].APIKey != "", c.Lidarr[0].Timeout, c.Lidarr[0].ValidSSL)
	} else {
		c.Print(" => Lidarr Config:", count, "servers")

		for _, f := range c.Lidarr {
			c.Printf(" =>    Server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.Name, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}
