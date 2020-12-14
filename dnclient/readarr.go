package dnclient

import (
	"sync"

	"golift.io/starr"
)

func (c *Config) fixReadarrConfig() {
	for i := range c.Readarr {
		if c.Readarr[i].Timeout.Duration == 0 {
			c.Readarr[i].Timeout.Duration = c.Timeout.Duration
		}
	}
}

// ReadarrConfig represents the input data for a Readarr server.
type ReadarrConfig struct {
	Name string `json:"name" toml:"name" xml:"name" yaml:"name"`
	*starr.Config
	sync.RWMutex `json:"-" toml:"-" xml:"-" yaml:"-"`
}

func (c *Client) logReadarr() {
	if count := len(c.Readarr); count == 1 {
		c.Printf(" => Readarr Config: 1 server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Readarr[0].Name, c.Readarr[0].URL, c.Readarr[0].APIKey != "", c.Readarr[0].Timeout, c.Readarr[0].ValidSSL)
	} else {
		c.Print(" => Readarr Config:", count, "servers")

		for _, f := range c.Readarr {
			c.Printf(" =>    Server: %s / %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.Name, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

func (c *Client) handleReadarr(p *IncomingPayload) (string, error) {
	return "", nil
}
