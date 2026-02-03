package client

/*
  This file contains the procedures that validate config data and initialize each app.
  All startup logs come from below. Every procedure in this file is run once on startup.
*/

import (
	"context"
	"os"
	"path"

	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
	"golift.io/version"
)

const (
	server       = "server"
	servers      = "servers"
	starrLogLine = " =>    Server %d: %s apikey:%v timeout:%s valid_ssl:%v " +
		"stuck/fin:%v/%v corrupt:%v backup:%v http/pass:%v/%v"
)

// PrintStartupInfo prints info about our startup config.
// This runs once on startup, and again during reloads.
func (c *Client) PrintStartupInfo(ctx context.Context, clientInfo *clientinfo.ClientInfo) {
	reqID := mnd.GetID(ctx)

	if clientInfo != nil {
		logs.Log.Printf(reqID, "==> %s", clientInfo)
		c.printVersionChangeInfo(reqID)
	} else {
		clientInfo = &clientinfo.ClientInfo{}
	}

	switch host, err := website.GetHostInfo(ctx); {
	case err != nil:
		logs.Log.Errorf(reqID, "=> Unknown Host Info (this is bad): %v", err)
	case c.Config.HostID == "":
		c.Config.HostID = host.HostID
		logs.Log.Printf(reqID, "==> {UNSAVED} Unique Host ID: %s (%s)", c.Config.HostID, host.Hostname)
	default:
		logs.Log.Printf(reqID, "==> Unique Host ID: %s (%s)", host.HostID, host.Hostname)
	}

	hostname, _ := os.Hostname()

	logs.Log.Printf(reqID, "==> %s <==", mnd.HelpLink)
	logs.Log.Printf(reqID, "==> %s Startup Settings <==", hostname)
	c.printLidarr(reqID, &clientInfo.Actions.Apps.Lidarr)
	c.printProwlarr(reqID, &clientInfo.Actions.Apps.Prowlarr)
	c.printRadarr(reqID, &clientInfo.Actions.Apps.Radarr)
	c.printReadarr(reqID, &clientInfo.Actions.Apps.Readarr)
	c.printSonarr(reqID, &clientInfo.Actions.Apps.Sonarr)
	c.printDeluge(reqID)
	c.printTransmission(reqID)
	c.printNZBGet(reqID)
	c.printQbit(reqID)
	c.printRtorrent(reqID)
	c.printSABnzbd(reqID)
	c.printPlex(reqID)
	c.printTautulli(reqID)
	c.printMySQL(reqID)
	logs.Log.Printf(reqID, " => Timeout: %s, Quiet: %v", c.Config.Timeout, c.Config.Quiet)

	if c.Config.UIPassword.Webauth() {
		logs.Log.Printf(reqID, " => Trusted Upstream Networks: %v, Auth Proxy Header: %s",
			c.allow, c.Config.UIPassword.Header())
	} else {
		logs.Log.Printf(reqID, " => Trusted Upstream Networks: %v", c.allow)
	}

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		logs.Log.Print(reqID, " => Web HTTPS Listen:", "https://"+c.Config.BindAddr+path.Join("/", c.Config.URLBase))
		logs.Log.Print(reqID, " => Web Cert & Key Files:", c.Config.SSLCrtFile+", "+c.Config.SSLKeyFile)
	} else {
		logs.Log.Print(reqID, " => Web HTTP Listen:", "http://"+c.Config.BindAddr+path.Join("/", c.Config.URLBase))
	}

	c.printLogFileInfo(reqID)
}

func (c *Client) printVersionChangeInfo(reqID string) {
	const clientVersion = "clientVersion"

	values, err := website.GetState(reqID, clientVersion)
	if err != nil {
		logs.Log.Errorf(reqID, "XX> Getting version from database: %v", err)
	}

	currentVersion := version.Version + "-" + version.Revision
	previousVersion := string(values[clientVersion])

	if previousVersion == currentVersion || version.Version == "" {
		return
	}

	if previousVersion == "" {
		hostname, _ := os.Hostname()
		logs.Log.Printf(reqID, "==> Detected a new client, %s. Welcome to Notifiarr!", hostname)
	} else {
		logs.Log.Printf(reqID, "==> Detected application version change! %s => %s", previousVersion, currentVersion)
	}

	err = website.SetState(clientVersion, []byte(currentVersion))
	if err != nil {
		logs.Log.Errorf(reqID, "Updating version in database: %v", err)
	}
}

func (c *Client) printLogFileInfo(reqID string) { //nolint:cyclop
	if c.Config.LogFile != "" {
		if c.Config.LogFiles > 0 {
			logs.Log.Printf(reqID, " => Log File: %s (%d @ %dMb)", c.Config.LogFile, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			logs.Log.Printf(reqID, " => Log File: %s (no rotation)", c.Config.LogFile)
		}
	}

	if c.Config.HTTPLog != "" {
		if c.Config.LogFiles > 0 {
			logs.Log.Printf(reqID, " => HTTP Log: %s (%d @ %dMb)", c.Config.HTTPLog, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			logs.Log.Printf(reqID, " => HTTP Log: %s (no rotation)", c.Config.HTTPLog)
		}
	}

	if c.Config.Debug && c.Config.LogConfig.DebugLog != "" {
		if c.Config.LogFiles > 0 {
			logs.Log.Printf(reqID, " => Debug Log: %s (%d @ %dMb)",
				c.Config.LogConfig.DebugLog, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			logs.Log.Printf(reqID, " => Debug Log: %s (no rotation)", c.Config.LogConfig.DebugLog)
		}
	}

	if c.Config.Services.LogFile != "" && !c.Config.Services.Disabled && len(c.Config.Service) > 0 {
		if c.Config.LogFiles > 0 {
			logs.Log.Printf(reqID, " => Service Checks Log: %s (%d @ %dMb)",
				c.Config.Services.LogFile, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			logs.Log.Printf(reqID, " => Service Checks Log: %s (no rotation)", c.Config.Services.LogFile)
		}
	}
}

// printPlex is called on startup to print info about configured Plex instance(s).
func (c *Client) printPlex(reqID string) {
	if !c.apps.Plex.Enabled() {
		return
	}

	name := c.apps.Plex.Server.Name()
	if name == "" {
		name = "<connection error?>"
	}

	logs.Log.Printf(reqID,
		" => Plex Config: 1 server: %s @ %s (enables incoming APIs and webhook) timeout:%v check_interval:%s ",
		name, c.apps.Plex.Server.URL, c.apps.Plex.Timeout, c.apps.Plex.Interval)
}

// printLidarr is called on startup to print info about each configured server.
func (c *Client) printLidarr(reqID string, app *clientinfo.InstanceConfig) {
	s := servers
	if len(c.Config.Lidarr) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => Lidarr Config:", len(c.Config.Lidarr), s)

	for idx, f := range c.Config.Lidarr {
		logs.Log.Printf(reqID, starrLogLine,
			idx+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, app.Stuck(idx+1), app.Finished(idx+1),
			app.Corrupt(idx+1) != "" && app.Corrupt(idx+1) != mnd.Disabled, app.Backup(idx+1) != mnd.Disabled,
			f.HTTPPass != "" && f.HTTPUser != "", f.Password != "" && f.Username != "")
	}
}

// printProwlarr is called on startup to print info about each configured server.
func (c *Client) printProwlarr(reqID string, app *clientinfo.InstanceConfig) {
	s := servers
	if len(c.Config.Prowlarr) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => Prowlarr Config:", len(c.Config.Prowlarr), s)

	for idx, f := range c.Config.Prowlarr {
		logs.Log.Printf(reqID, starrLogLine,
			idx+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, "na", "na",
			app.Corrupt(idx+1) != "" && app.Corrupt(idx+1) != mnd.Disabled, app.Backup(idx+1) != mnd.Disabled,
			f.HTTPPass != "" && f.HTTPUser != "", f.Password != "" && f.Username != "")
	}
}

// printRadarr is called on startup to print info about each configured server.
func (c *Client) printRadarr(reqID string, app *clientinfo.InstanceConfig) {
	s := servers
	if len(c.Config.Radarr) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => Radarr Config:", len(c.Config.Radarr), s)

	for idx, f := range c.Config.Radarr {
		logs.Log.Printf(reqID, starrLogLine,
			idx+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, app.Stuck(idx+1), app.Finished(idx+1),
			app.Corrupt(idx+1) != "" && app.Corrupt(idx+1) != mnd.Disabled, app.Backup(idx+1) != mnd.Disabled,
			f.HTTPPass != "" && f.HTTPUser != "", f.Password != "" && f.Username != "")
	}
}

// printReadarr is called on startup to print info about each configured server.
func (c *Client) printReadarr(reqID string, app *clientinfo.InstanceConfig) {
	s := servers
	if len(c.Config.Readarr) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => Readarr Config:", len(c.Config.Readarr), s)

	for idx, f := range c.Config.Readarr {
		logs.Log.Printf(reqID, starrLogLine,
			idx+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, app.Stuck(idx+1), app.Finished(idx+1),
			app.Corrupt(idx+1) != "" && app.Corrupt(idx+1) != mnd.Disabled, app.Backup(idx+1) != mnd.Disabled,
			f.HTTPPass != "" && f.HTTPUser != "", f.Password != "" && f.Username != "")
	}
}

// printSonarr is called on startup to print info about each configured server.
func (c *Client) printSonarr(reqID string, app *clientinfo.InstanceConfig) {
	s := servers
	if len(c.Config.Sonarr) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => Sonarr Config:", len(c.Config.Sonarr), s)

	for idx, f := range c.Config.Sonarr {
		logs.Log.Printf(reqID, starrLogLine,
			idx+1, f.URL, f.APIKey != "", f.Timeout, f.ValidSSL, app.Stuck(idx+1), app.Finished(idx+1),
			app.Corrupt(idx+1) != "" && app.Corrupt(idx+1) != mnd.Disabled, app.Backup(idx+1) != mnd.Disabled,
			f.HTTPPass != "" && f.HTTPUser != "", f.Password != "" && f.Username != "")
	}
}

// printDeluge is called on startup to print info about each configured server.
func (c *Client) printDeluge(reqID string) {
	s := servers
	if len(c.Config.Deluge) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => Deluge Config:", len(c.Config.Deluge), s)

	for i, f := range c.Config.Deluge {
		logs.Log.Printf(reqID, " =>    Server %d: %s password:%v timeout:%s valid_ssl:%v",
			i+1, f.Config.URL, f.Password != "", cnfg.Duration{Duration: f.Timeout.Duration}, f.ValidSSL)
	}
}

// printTransmission is called on startup to print info about each configured server.
func (c *Client) printTransmission(reqID string) {
	s := servers
	if len(c.Config.Transmission) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => Transmission Config:", len(c.Config.Transmission), s)

	for i, f := range c.Config.Transmission {
		logs.Log.Printf(reqID, " =>    Server %d: %s username:%s password:%v timeout:%s valid_ssl:%v",
			i+1, f.URL, f.User, f.Pass != "", cnfg.Duration{Duration: f.Timeout.Duration}, f.ValidSSL)
	}
}

// printNZBGet is called on startup to print info about each configured server.
func (c *Client) printNZBGet(reqID string) {
	s := servers
	if len(c.Config.NZBGet) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => NZBGet Config:", len(c.Config.NZBGet), s)

	for i, f := range c.Config.NZBGet {
		logs.Log.Printf(reqID, " =>    Server %d: %s username:%s password:%v timeout:%s valid_ssl:%v",
			i+1, f.Config.URL, f.User, f.Pass != "", cnfg.Duration{Duration: f.Timeout.Duration}, f.ValidSSL)
	}
}

// printQbit is called on startup to print info about each configured server.
func (c *Client) printQbit(reqID string) {
	s := servers
	if len(c.Config.Qbit) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => Qbit Config:", len(c.Config.Qbit), s)

	for i, f := range c.Config.Qbit {
		logs.Log.Printf(reqID, " =>    Server %d: %s username:%s password:%v timeout:%s valid_ssl:%v",
			i+1, f.Config.URL, f.User, f.Pass != "", cnfg.Duration{Duration: f.Timeout.Duration}, f.ValidSSL)
	}
}

// printRtorrent is called on startup to print info about each configured server.
func (c *Client) printRtorrent(reqID string) {
	s := servers
	if len(c.Config.Rtorrent) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => rTorrent Config:", len(c.Config.Rtorrent), s)

	for i, f := range c.Config.Rtorrent {
		logs.Log.Printf(reqID, " =>    Server %d: %s username:%s password:%v timeout:%s valid_ssl:%v",
			i+1, f.URL, f.User, f.Pass != "", cnfg.Duration{Duration: f.Timeout.Duration}, f.ValidSSL)
	}
}

// printSABnzbd is called on startup to print info about each configured SAB downloader.
func (c *Client) printSABnzbd(reqID string) {
	s := servers
	if len(c.Config.SabNZB) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => SABnzbd Config:", len(c.Config.SabNZB), s)

	for i, f := range c.Config.SabNZB {
		logs.Log.Printf(reqID, " =>    Server %d: %s, api_key:%v timeout:%s", i+1, f.URL, f.APIKey != "", f.Timeout)
	}
}

// printTautulli is called on startup to print info about configured Tautulli instance(s).
func (c *Client) printTautulli(reqID string) {
	switch taut := c.Config.Tautulli; {
	case !taut.Enabled():
		logs.Log.Printf(reqID, " => Tautulli Config (enables name map): 0 servers")
	case taut.Name != "":
		logs.Log.Printf(reqID, " => Tautulli Config (enables name map): 1 server: %s timeout:%v check_interval:%s name:%s",
			taut.URL, taut.Timeout, taut.Interval, taut.Name)
	default:
		logs.Log.Printf(reqID, " => Tautulli Config (enables name map): 1 server: %s timeout:%s", taut.URL, taut.Timeout)
	}
}

// printMySQL is called on startup to print info about each configured SQL server.
func (c *Client) printMySQL(reqID string) {
	s := servers
	if len(c.Config.Snapshot.MySQL) == 1 {
		s = server
	}

	logs.Log.Print(reqID, " => MySQL Config:", len(c.Config.Snapshot.MySQL), s)

	for i, m := range c.Config.Snapshot.MySQL {
		if m.Name != "" {
			logs.Log.Printf(reqID, " =>    Server %d: %s user:%v timeout:%s check_interval:%s name:%s",
				i+1, m.Host, m.User, m.Timeout, m.Interval, m.Name)
		} else {
			logs.Log.Printf(reqID, " =>    Server %d: %s user:%v timeout:%s", i+1, m.Host, m.User, m.Timeout)
		}
	}
}
