//go:build darwin || windows
// +build darwin windows

package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/hako/durafmt"
	"golift.io/version"
)

/* This file contains methdos that are triggered from the GUI menu. */

func (c *Client) toggleServer() {
	if c.server == nil {
		ui.Notify("Started web server") //nolint:errcheck
		c.Print("[user requested] Starting Web Server")
		c.StartWebServer()

		return
	}

	ui.Notify("Paused web server") //nolint:errcheck
	c.Print("[user requested] Pausing Web Server")

	if err := c.StopWebServer(); err != nil {
		c.Errorf("Unable to Pause Server: %v", err)
	}
}

func (c *Client) rotateLogs() {
	c.Print("[user requested] Rotating Log Files!")
	ui.Notify("Rotating log files") //nolint:errcheck

	for _, err := range c.Logger.Rotate() {
		if err != nil {
			ui.Notify("Error rotateing log files: %v", err) //nolint:errcheck
			c.Errorf("Rotating Log Files: %v", err)
		}
	}
}

// changeKey shuts down the web server and changes the API key.
// The server has to shut down to avoid race conditions.
func (c *Client) changeKey() {
	key, ok, err := ui.Entry(mnd.Title+": Configuration", "API Key", c.Config.APIKey)
	if err != nil {
		c.Errorf("Updating API Key: %v", err)
	} else if !ok || key == c.Config.APIKey {
		return
	}

	c.Print("[user requested] Updating API Key!")

	if err := c.StopWebServer(); err != nil && !errors.Is(err, ErrNoServer) {
		c.Errorf("Unable to update API Key: %v", err)
		return
	} else if !errors.Is(err, ErrNoServer) {
		defer c.StartWebServer()
	}

	c.Config.APIKey = key
}

func (c *Client) checkForUpdate() {
	c.Print("[user requested] GitHub Update Check")

	switch update, err := update.Check(mnd.UserRepo, version.Version); {
	case err != nil:
		c.Errorf("Update Check: %v", err)
		_, _ = ui.Error(mnd.Title+" ERROR", "Checking version on GitHub: "+err.Error())
	case update.Outdate && runtime.GOOS == mnd.Windows:
		c.upgradeWindows(update)
	case update.Outdate:
		c.downloadOther(update)
	default:
		_, _ = ui.Info(mnd.Title, "You're up to date! Version: "+update.Version+"\n"+
			"Updated: "+update.RelDate.Format("Jan 2, 2006")+" ("+
			durafmt.Parse(time.Since(update.RelDate).Round(time.Hour)).String()+" ago)")
	}
}

func (c *Client) downloadOther(update *update.Update) {
	yes, _ := ui.Question(mnd.Title, "An Update is available! Download?\n\n"+
		"Your Version: "+update.Version+"\n"+
		"New Version: "+update.Current+"\n"+
		"Date: "+update.RelDate.Format("Jan 2, 2006")+" ("+
		durafmt.Parse(time.Since(update.RelDate).Round(time.Hour)).String()+" ago)", false)
	if yes {
		_ = ui.OpenURL(update.CurrURL)
	}
}

// This is always outdated. :( The format on screen sucs anyway. This should probably be removed.
func (c *Client) displayConfig() (s string) { //nolint: funlen,cyclop
	s = "Config File: " + c.Flags.ConfigFile
	s += fmt.Sprintf("\nTimeout: %v", c.Config.Timeout)
	s += fmt.Sprintf("\nUpstreams: %v", c.Config.Allow)

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		s += fmt.Sprintf("\nHTTPS: https://%s%s", c.Config.BindAddr, c.Config.URLBase)
		s += fmt.Sprintf("\nCert File: %v", c.Config.SSLCrtFile)
		s += fmt.Sprintf("\nCert Key: %v", c.Config.SSLKeyFile)
	} else {
		s += fmt.Sprintf("\nHTTP: http://%s%s", c.Config.BindAddr, c.Config.URLBase)
	}

	if c.Config.LogFiles > 0 {
		s += fmt.Sprintf("\nLog File: %v (%d @ %dMb)", c.Config.LogFile, c.Config.LogFiles, c.Config.LogFileMb)
		s += fmt.Sprintf("\nHTTP Log: %v (%d @ %dMb)", c.Config.HTTPLog, c.Config.LogFiles, c.Config.LogFileMb)
	} else {
		s += fmt.Sprintf("\nLog File: %v (no rotation)", c.Config.LogFile)
		s += fmt.Sprintf("\nHTTP Log: %v (no rotation)", c.Config.HTTPLog)
	}

	if count := len(c.Config.Lidarr); count == 1 {
		s += fmt.Sprintf("\n- Lidarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Lidarr[0].URL, c.Config.Lidarr[0].APIKey != "", c.Config.Lidarr[0].Timeout, c.Config.Lidarr[0].ValidSSL)
	} else {
		for _, f := range c.Config.Lidarr {
			s += fmt.Sprintf("\n- Lidarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}

	if count := len(c.Config.Radarr); count == 1 {
		s += fmt.Sprintf("\n- Radarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Radarr[0].URL, c.Config.Radarr[0].APIKey != "", c.Config.Radarr[0].Timeout, c.Config.Radarr[0].ValidSSL)
	} else {
		for _, f := range c.Config.Radarr {
			s += fmt.Sprintf("\n- Radarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}

	if count := len(c.Config.Readarr); count == 1 {
		s += fmt.Sprintf("\n- Readarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Readarr[0].URL, c.Config.Readarr[0].APIKey != "", c.Config.Readarr[0].Timeout, c.Config.Readarr[0].ValidSSL)
	} else {
		for _, f := range c.Config.Readarr {
			s += fmt.Sprintf("\n- Readarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}

	if count := len(c.Config.Sonarr); count == 1 {
		s += fmt.Sprintf("\n- Sonarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Sonarr[0].URL, c.Config.Sonarr[0].APIKey != "", c.Config.Sonarr[0].Timeout, c.Config.Sonarr[0].ValidSSL)
	} else {
		for _, f := range c.Config.Sonarr {
			s += fmt.Sprintf("\n- Sonarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}

	return s + "\n"
}

// sendPlexSessions is triggered from a menu-bar item.
func (c *Client) sendPlexSessions(url string) {
	c.Printf("[user requested] Sending Plex Sessions to %s", url)

	if body, err := c.notifiarr.SendMeta(notifiarr.PlexCron, url, nil, false); err != nil {
		c.Errorf("[user requested] Sending Plex Sessions to %s: %v", url, err)
	} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
		c.Printf("[user requested] Sent Plex Sessions to %s, reply: %s", url, fields[3])
	} else {
		c.Printf("[user requested] Sent Plex Sessions to %s", url)
	}
}

func (c *Client) writeConfigFile() {
	val, _, _ := ui.Entry(mnd.Title, "Enter path to write config file:", c.Flags.ConfigFile)

	if val == "" {
		_, _ = ui.Error(mnd.Title+" Error", "No Config File Provided")
		return
	}

	c.Print("[user requested] Writing Config File:", val)

	if _, err := c.Config.Write(val); err != nil {
		c.Errorf("Writing Config File: %v", err)
		_, _ = ui.Error(mnd.Title+" Error", "Writing Config File: "+err.Error())

		return
	}

	_, _ = ui.Info(mnd.Title, "Wrote Config File: "+val)
}

func (c *Client) menuPanic() {
	yes, err := ui.Question(mnd.Title, "You really want to panic?", true)
	if !yes || err != nil {
		return
	}

	defer c.Printf("User Requested Application Panic, good bye.")
	panic("user requested panic")
}

// sendSystemSnapshot is triggered from a menu-bar item, and from --send cli arg.
func (c *Client) sendSystemSnapshot(url string) {
	c.Printf("[user requested] Sending System Snapshot to %s", url)

	snaps, errs, debug := c.Config.Snapshot.GetSnapshot()
	for _, err := range errs {
		if err != nil {
			c.Errorf("[user requested] %v", err)
		}
	}

	for _, err := range debug {
		if err != nil {
			c.Errorf("[user requested] %v", err)
		}
	}

	payload := &notifiarr.Payload{Type: notifiarr.SnapCron, Snap: snaps}
	if _, body, err := c.notifiarr.SendData(url, payload, true); err != nil {
		c.Errorf("[user requested] Sending System Snapshot to %s: %v", url, err)
	} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
		c.Printf("[user requested] Sent System Snapshot to %s, reply: %s", url, fields[3])
	} else {
		c.Printf("[user requested] Sent System Snapshot to %s", url)
	}
}

// logSnaps writes a full snapshot payload to the log file.
func (c *Client) logSnaps() {
	c.Printf("[user requested] Collecting Snapshot from Plex and the System (for log file).")

	snaps, errs, debug := c.Config.Snapshot.GetSnapshot()
	for _, err := range errs {
		if err != nil {
			c.Errorf("[user requested] %v", err)
		}
	}

	for _, err := range debug {
		if err != nil {
			c.Errorf("[user requested] %v", err)
		}
	}

	var (
		plex *plex.Sessions
		err  error
	)

	if c.Config.Plex.Configured() {
		if plex, err = c.Config.Plex.GetXMLSessions(); err != nil {
			c.Errorf("[user requested] %v", err)
		}
	}

	b, _ := json.MarshalIndent(&notifiarr.Payload{
		Type: notifiarr.LogLocal,
		Snap: snaps,
		Plex: plex,
	}, "", "  ")
	c.Printf("[user requested] Snapshot Data:\n%s", string(b))
}
