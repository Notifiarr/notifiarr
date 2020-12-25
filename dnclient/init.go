//nolint:dupl,golint
package dnclient

/*
  This file contains the procedures that validate config data and initialize each app.
  All startup logs come from below. Every procedure in this file is run once on startup.
*/

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"golift.io/starr/lidarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

// InitStartup fixes config problems and prints info about our startup config.
// This runs once on startup.
func (c *Client) InitStartup() {
	c.Printf("==> %s <==", helpLink)
	c.Print("==> Startup Settings <==")
	c.initSonarr()
	c.initRadarr()
	c.initLidarr()
	c.initReadarr()
	c.Printf(" => Timeout: %v, Debug: %v, Quiet: %v", c.Config.Timeout, c.Config.Debug, c.Config.Quiet)
	c.Print(" => Allows Upstream Networks:", c.allow.combineUpstreams())

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		c.Print(" => Web HTTPS Listen:", "https://"+c.Config.BindAddr+path.Join("/", c.Config.URLBase))
		c.Print(" => Web Cert & Key Files:", c.Config.SSLCrtFile+", "+c.Config.SSLKeyFile)
	} else {
		c.Print(" => Web HTTP Listen:", "http://"+c.Config.BindAddr+path.Join("/", c.Config.URLBase))
	}

	if c.Config.LogFile != "" {
		msg := "no rotation"
		if c.Config.LogFiles > 0 {
			msg = fmt.Sprintf("%d @ %dMb", c.Config.LogFiles, c.Config.LogFileMb)
		}

		c.Printf(" => Log File: %s (%s)", c.Config.LogFile, msg)
	}
}

func (n allowedIPs) combineUpstreams() (s string) {
	for i := range n {
		if s != "" {
			s += ", "
		}

		s += n[i].String()
	}

	return s
}

// Exit stops the web server and logs our exit messages. Start() calls this.
func (c *Client) Exit() error {
	signal.Notify(c.signal, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	c.Printf("[%s] Need help? %s\n=====> Exiting! Caught Signal: %v", c.Flags.Name(), helpLink, <-c.signal)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if c.server != nil {
		defer func() { c.server = nil }()

		return c.server.Shutdown(ctx)
	}

	return nil
}

// initLidarr is called on startup to fix the config and print info about each configured server.
func (c *Client) initLidarr() {
	for i := range c.Config.Lidarr {
		c.Config.Lidarr[i].Lidarr = lidarr.New(c.Config.Lidarr[i].Config)
		if c.Config.Lidarr[i].Timeout.Duration == 0 {
			c.Config.Lidarr[i].Timeout.Duration = c.Config.Timeout.Duration
		}
	}

	if count := len(c.Config.Lidarr); count == 1 {
		c.Printf(" => Lidarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Lidarr[0].URL, c.Config.Lidarr[0].APIKey != "", c.Config.Lidarr[0].Timeout, c.Config.Lidarr[0].ValidSSL)
	} else {
		c.Print(" => Lidarr Config:", count, "servers")

		for _, f := range c.Config.Lidarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// initRadarr is called on startup to fix the config and print info about each configured server.
func (c *Client) initRadarr() {
	for i := range c.Config.Radarr {
		c.Config.Radarr[i].Radarr = radarr.New(c.Config.Radarr[i].Config)
		if c.Config.Radarr[i].Timeout.Duration == 0 {
			c.Config.Radarr[i].Timeout.Duration = c.Config.Timeout.Duration
		}
	}

	if count := len(c.Config.Radarr); count == 1 {
		c.Printf(" => Radarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Radarr[0].URL, c.Config.Radarr[0].APIKey != "", c.Config.Radarr[0].Timeout, c.Config.Radarr[0].ValidSSL)
	} else {
		c.Print(" => Radarr Config:", count, "servers")

		for _, f := range c.Config.Radarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// initReadarr is called on startup to fix the config and print info about each configured server.
func (c *Client) initReadarr() {
	for i := range c.Config.Readarr {
		c.Config.Readarr[i].Readarr = readarr.New(c.Config.Readarr[i].Config)
		if c.Config.Readarr[i].Timeout.Duration == 0 {
			c.Config.Readarr[i].Timeout.Duration = c.Config.Timeout.Duration
		}
	}

	if count := len(c.Config.Readarr); count == 1 {
		c.Printf(" => Readarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Readarr[0].URL, c.Config.Readarr[0].APIKey != "", c.Config.Readarr[0].Timeout, c.Config.Readarr[0].ValidSSL)
	} else {
		c.Print(" => Readarr Config:", count, "servers")

		for _, f := range c.Config.Readarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}

// initSonarr is called on startup to fix the config and print info about each configured server.
func (c *Client) initSonarr() {
	for i := range c.Config.Sonarr {
		c.Config.Sonarr[i].Sonarr = sonarr.New(c.Config.Sonarr[i].Config)
		if c.Config.Sonarr[i].Timeout.Duration == 0 {
			c.Config.Sonarr[i].Timeout.Duration = c.Config.Timeout.Duration
		}
	}

	if count := len(c.Config.Sonarr); count == 1 {
		c.Printf(" => Sonarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Sonarr[0].URL, c.Config.Sonarr[0].APIKey != "", c.Config.Sonarr[0].Timeout, c.Config.Sonarr[0].ValidSSL)
	} else {
		c.Print(" => Sonarr Config:", count, "servers")

		for _, f := range c.Config.Sonarr {
			c.Printf(" =>    Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}
}
