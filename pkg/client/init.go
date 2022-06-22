package client

/*
  This file contains the procedures that validate config data and initialize each app.
  All startup logs come from below. Every procedure in this file is run once on startup.
*/

import (
	"path"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golift.io/cnfg"
	"golift.io/version"
)

// PrintStartupInfo prints info about our startup config.
// This runs once on startup, and again during reloads.
func (c *Client) PrintStartupInfo(clientInfo *website.ClientInfo) {
	if clientInfo != nil {
		c.Printf("==> %s", clientInfo)
		c.printVersionChangeInfo()
	}

	if hi, err := c.website.GetHostInfo(); err != nil {
		c.Errorf("=> Unknown Host Info (this is bad): %v", err)
	} else {
		c.Printf("==> Unique ID: %s (%s)", hi.HostID, hi.Hostname)
	}

	title := cases.Title(language.AmericanEnglish)

	c.Printf("==> %s <==", mnd.HelpLink)
	c.Printf("==> %s Startup Settings <==", title.String(strings.ToLower(c.Config.Mode)))
	c.printLidarr()
	c.printProwlarr()
	c.printRadarr()
	c.printReadarr()
	c.printSonarr()
	c.printDeluge()
	c.printQbit()
	c.printSABnzbd()
	c.printPlex()
	c.printTautulli()
	c.printMySQL()
	c.Printf(" => Timeout: %s, Quiet: %v", c.Config.Timeout, c.Config.Quiet)
	c.Printf(" => Trusted Upstream Networks: %v", c.Config.Allow)

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		c.Print(" => Web HTTPS Listen:", "https://"+c.Config.BindAddr+path.Join("/", c.Config.URLBase))
		c.Print(" => Web Cert & Key Files:", c.Config.SSLCrtFile+", "+c.Config.SSLKeyFile)
	} else {
		c.Print(" => Web HTTP Listen:", "http://"+c.Config.BindAddr+path.Join("/", c.Config.URLBase))
	}

	c.printLogFileInfo()
}

func (c *Client) printVersionChangeInfo() {
	const clientVersion = "clientVersion"

	values, err := c.website.GetValue(clientVersion)
	if err != nil {
		c.Errorf("XX> Getting version from databse: %v", err)
	}

	previousVersion := string(values[clientVersion])
	if previousVersion == "" ||
		previousVersion == version.Version ||
		version.Version == "" {
		return
	}

	c.Printf("==> Detected application version change! %s => %s", previousVersion, version.Version)

	go func() { // in background to improve startup time.
		err := c.website.SetValue(clientVersion, []byte(version.Version))
		if err != nil {
			c.Errorf("Updating version in database: %v", err)
		}
	}()
}

func (c *Client) printLogFileInfo() { //nolint:cyclop
	if c.Config.LogFile != "" {
		if c.Config.LogFiles > 0 {
			c.Printf(" => Log File: %s (%d @ %dMb)", c.Config.LogFile, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			c.Printf(" => Log File: %s (no rotation)", c.Config.LogFile)
		}
	}

	if c.Config.HTTPLog != "" {
		if c.Config.LogFiles > 0 {
			c.Printf(" => HTTP Log: %s (%d @ %dMb)", c.Config.HTTPLog, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			c.Printf(" => HTTP Log: %s (no rotation)", c.Config.HTTPLog)
		}
	}

	if c.Config.Debug && c.Config.LogConfig.DebugLog != "" {
		if c.Config.LogFiles > 0 {
			c.Printf(" => Debug Log: %s (%d @ %dMb)", c.Config.LogConfig.DebugLog, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			c.Printf(" => Debug Log: %s (no rotation)", c.Config.LogConfig.DebugLog)
		}
	}

	if c.Config.Services.LogFile != "" && !c.Config.Services.Disabled && len(c.Config.Service) > 0 {
		if c.Config.LogFiles > 0 {
			c.Printf(" => Service Checks Log: %s (%d @ %dMb)", c.Config.Services.LogFile, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			c.Printf(" => Service Checks Log: %s (no rotation)", c.Config.Services.LogFile)
		}
	}
}

// printPlex is called on startup to print info about configured Plex instance(s).
func (c *Client) printPlex() {
	plex := c.Config.Plex
	if !plex.Configured() {
		return
	}

	name := plex.Name
	if name == "" {
		name = "<possible connection error>"
	}

	c.Printf(" => Plex Config: 1 server: %s @ %s (enables incoming APIs and webhook)", name, plex.URL)
}

// printLidarr is called on startup to print info about each configured server.
func (c *Client) printLidarr() {
	if len(c.Config.Lidarr) == 1 {
		f := c.Config.Lidarr[0]

		c.Printf(" => Lidarr Config: 1 server: %s apikey:%v timeout:%s verify_ssl:%v stuck_items:%v corrupt:%v backup:%v",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.StuckItem, f.Corrupt != "" &&
				f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)

		return
	}

	c.Print(" => Lidarr Config:", len(c.Config.Lidarr), "servers")

	for i, f := range c.Config.Lidarr {
		c.Printf(" =>    Server %d: %s apikey:%v timeout:%s verify_ssl:%v stuck_items:%v corrupt:%v backup:%v",
			i+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.StuckItem,
			f.Corrupt != "" && f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)
	}
}

// printProwlarr is called on startup to print info about each configured server.
func (c *Client) printProwlarr() {
	if len(c.Config.Prowlarr) == 1 {
		f := c.Config.Prowlarr[0]

		c.Printf(" => Prowlarr Config: 1 server: %s apikey:%v timeout:%s verify_ssl:%v corrupt:%v backup:%v",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.Corrupt != "" &&
				f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)

		return
	}

	c.Print(" => Prowlarr Config:", len(c.Config.Prowlarr), "servers")

	for i, f := range c.Config.Prowlarr {
		c.Printf(" =>    Server %d: %s apikey:%v timeout:%s verify_ssl:%v corrupt:%v backup:%v",
			i+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.Corrupt != "" &&
				f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)
	}
}

// printRadarr is called on startup to print info about each configured server.
func (c *Client) printRadarr() {
	if len(c.Config.Radarr) == 1 {
		f := c.Config.Radarr[0]

		c.Printf(" => Radarr Config: 1 server: %s apikey:%v timeout:%s verify_ssl:%v stuck_items:%v corrupt:%v backup:%v",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.StuckItem, f.Corrupt != "" &&
				f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)

		return
	}

	c.Print(" => Radarr Config:", len(c.Config.Radarr), "servers")

	for i, f := range c.Config.Radarr {
		c.Printf(" =>    Server %d: %s apikey:%v timeout:%s verify_ssl:%v stuck_items:%v corrupt:%v backup:%v",
			i+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.StuckItem, f.Corrupt != "" &&
				f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)
	}
}

// printReadarr is called on startup to print info about each configured server.
func (c *Client) printReadarr() {
	if len(c.Config.Readarr) == 1 {
		f := c.Config.Readarr[0]

		c.Printf(" => Readarr Config: 1 server: %s apikey:%v timeout:%s verify_ssl:%v stuck_items:%v corrupt:%v backup:%v",
			f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.StuckItem, f.Corrupt != "" &&
				f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)

		return
	}

	c.Print(" => Readarr Config:", len(c.Config.Readarr), "servers")

	for i, f := range c.Config.Readarr {
		c.Printf(" =>    Server %d: %s apikey:%v timeout:%s verify_ssl:%v stuck_items:%v corrupt:%v backup:%v",
			i+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.StuckItem, f.Corrupt != "" &&
				f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)
	}
}

// printSonarr is called on startup to print info about each configured server.
func (c *Client) printSonarr() {
	if len(c.Config.Sonarr) == 1 {
		f := c.Config.Sonarr[0]

		c.Printf(" => Sonarr Config: 1 server: %s apikey:%v timeout:%s verify_ssl:%v stuck_items:%v corrupt:%v backup:%v",
			f.URL, f.APIKey != "", f.Timeout.String(), f.ValidSSL, f.StuckItem, f.Corrupt != "" &&
				f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)

		return
	}

	c.Print(" => Sonarr Config:", len(c.Config.Sonarr), "servers")

	for i, f := range c.Config.Sonarr {
		c.Printf(" =>    Server %d: %s apikey:%v timeout:%s verify_ssl:%v stuck_items:%v corrupt:%v backup:%v",
			i+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, f.StuckItem, f.Corrupt != "" &&
				f.Corrupt != mnd.Disabled, f.Backup != mnd.Disabled)
	}
}

// printDeluge is called on startup to print info about each configured server.
func (c *Client) printDeluge() {
	if len(c.Config.Deluge) == 1 {
		f := c.Config.Deluge[0]

		c.Printf(" => Deluge Config: 1 server: %s password:%v timeout:%s verify_ssl:%v",
			f.Config.URL, f.Password != "", cnfg.Duration{Duration: f.Timeout.Duration}, f.VerifySSL)

		return
	}

	c.Print(" => Deluge Config:", len(c.Config.Deluge), "servers")

	for i, f := range c.Config.Deluge {
		c.Printf(" =>    Server %d: %s password:%v timeout:%s verify_ssl:%v",
			i+1, f.Config.URL, f.Password != "", cnfg.Duration{Duration: f.Timeout.Duration}, f.VerifySSL)
	}
}

// printQbit is called on startup to print info about each configured server.
func (c *Client) printQbit() {
	if len(c.Config.Qbit) == 1 {
		f := c.Config.Qbit[0]

		c.Printf(" => Qbit Config: 1 server: %s username:%s password:%v timeout:%s verify_ssl:%v",
			f.Config.URL, f.User, f.Pass != "", cnfg.Duration{Duration: f.Timeout.Duration}, f.VerifySSL)

		return
	}

	c.Print(" => Qbit Config:", len(c.Config.Qbit), "servers")

	for i, f := range c.Config.Qbit {
		c.Printf(" =>    Server %d: %s username:%s password:%v timeout:%s verify_ssl:%v",
			i+1, f.Config.URL, f.User, f.Pass != "", cnfg.Duration{Duration: f.Timeout.Duration}, f.VerifySSL)
	}
}

// printSABnzbd is called on startup to print info about each configured SAB downloader.
func (c *Client) printSABnzbd() {
	if len(c.Config.SabNZB) == 1 {
		f := c.Config.SabNZB[0]

		c.Printf(" => SABnzbd Config: 1 server: %s api_key:%v timeout:%s", f.URL, f.APIKey != "", f.Timeout)

		return
	}

	c.Print(" => SABnzbd Config:", len(c.Config.SabNZB), "servers")

	for i, f := range c.Config.SabNZB {
		c.Printf(" =>    Server %d: %s, api_key:%v timeout:%s", i+1, f.URL, f.APIKey != "", f.Timeout)
	}
}

// printTautulli is called on startup to print info about configured Tautulli instance(s).
func (c *Client) printTautulli() {
	switch taut := c.Config.Apps.Tautulli; {
	case taut == nil, taut.URL == "":
		c.Printf(" => Tautulli Config (enables name map): 0 servers")
	case taut.Name != "":
		c.Printf(" => Tautulli Config (enables name map): 1 server: %s timeout:%v check_interval:%s name:%s",
			taut.URL, taut.Timeout, taut.Interval, taut.Name)
	default:
		c.Printf(" => Tautulli Config (enables name map): 1 server: %s timeout:%s", taut.URL, taut.Timeout)
	}
}

// printMySQL is called on startup to print info about each configured SQL server.
func (c *Client) printMySQL() {
	if c.Config.Snapshot.Plugins == nil { // unlikely.
		return
	}

	if len(c.Config.Snapshot.MySQL) == 1 {
		if m := c.Config.Snapshot.MySQL[0]; m.Name != "" {
			c.Printf(" => MySQL Config: 1 server: %s user:%v timeout:%s check_interval:%s name:%s",
				m.Host, m.User, m.Timeout, m.Interval, m.Name)
		} else {
			c.Printf(" => MySQL Config: 1 server: %s user:%v timeout:%s", m.Host, m.User, m.Timeout)
		}

		return
	}

	c.Print(" => MySQL Config:", len(c.Config.Snapshot.MySQL), "servers")

	for i, m := range c.Config.Snapshot.MySQL {
		if m.Name != "" {
			c.Printf(" =>    Server %d: %s user:%v timeout:%s check_interval:%s name:%s",
				i+1, m.Host, m.User, m.Timeout, m.Interval, m.Name)
		} else {
			c.Printf(" =>    Server %d: %s user:%v timeout:%s", i+1, m.Host, m.User, m.Timeout)
		}
	}
}
