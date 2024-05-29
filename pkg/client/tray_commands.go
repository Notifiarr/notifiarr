//go:build darwin || windows || linux

package client

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"golift.io/version"
)

/* This file contains methods that are triggered from the GUI menu. */

const TitleError = mnd.Title + " Error"

func (c *Client) toggleServer(ctx context.Context) {
	if !menu["stat"].Checked() {
		ui.Notify("Started web server") //nolint:errcheck
		c.Printf("[user requested] Starting Web Server, baseurl: %s, bind address: %s",
			c.Config.URLBase, c.Config.BindAddr)
		c.StartWebServer(ctx)

		return
	}

	ui.Notify("Paused web server") //nolint:errcheck
	c.Print("[user requested] Pausing Web Server")

	if err := c.StopWebServer(ctx); err != nil {
		c.Errorf("Unable to Pause Server: %v", err)
	}
}

func (c *Client) rotateLogs() {
	c.Print("[user requested] Rotating Log Files!")
	ui.Notify("Rotating log files") //nolint:errcheck

	for _, err := range c.Logger.Rotate() {
		if err != nil {
			ui.Notify("Error rotating log files: %v", err) //nolint:errcheck
			c.Errorf("Rotating Log Files: %v", err)
		}
	}
}

func (c *Client) checkForUpdate(ctx context.Context, unstable bool) {
	var (
		data  *update.Update
		err   error
		where = "GitHub"
	)

	if unstable {
		c.Print("[user requested] Unstable Update Check")

		data, err = update.CheckUnstable(ctx, mnd.Title, version.Revision)
		where = "Unstable website"
	} else {
		c.Print("[user requested] GitHub Update Check")

		data, err = update.CheckGitHub(ctx, mnd.UserRepo, version.Version)
	}

	switch {
	case err != nil:
		c.Errorf("Update Check: %v", err)
		_, _ = ui.Error(TitleError, "Checking version on "+where+": "+err.Error())
	case data.Outdate && runtime.GOOS == mnd.Windows:
		c.upgradeWindows(ctx, data)
	case data.Outdate:
		c.downloadOther(data, unstable)
	default:
		_, _ = ui.Info(mnd.Title, "You're up to date! Version: "+data.Current+"\n"+
			"Updated: "+data.RelDate.Format("Jan 2, 2006")+mnd.DurationAgo(data.RelDate))
	}
}

func (c *Client) downloadOther(update *update.Update, unstable bool) {
	msg := "An Update is available! Download?\n\n"

	if unstable {
		msg = "An Unstable Update is available! Download?\n\n"
	}

	yes, _ := ui.Question(mnd.Title, msg+
		"Your Version: "+version.Version+"-"+version.Revision+"\n"+
		"New Version: "+update.Current+"\n"+
		"Date: "+update.RelDate.Format("Jan 2, 2006")+mnd.DurationAgo(update.RelDate), false)
	if yes {
		_ = ui.OpenURL(update.CurrURL)
	}
}

// This is always outdated. :( The format on screen sucks anyway. This should probably be removed.
func (c *Client) displayConfig() (s string) { //nolint: funlen,cyclop
	s = "Config File: " + c.Flags.ConfigFile
	s += fmt.Sprintf("\nTimeout: %v", c.Config.Timeout)
	s += fmt.Sprintf("\nUpstreams: %v", c.Config.Allow.Input)

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

func (c *Client) writeConfigFile(ctx context.Context) {
	val, _, _ := ui.Entry(mnd.Title, "Enter path to write config file:", c.Flags.ConfigFile)

	if val == "" {
		_, _ = ui.Error(TitleError, "No Config File Provided")
		return
	}

	c.Print("[user requested] Writing Config File:", val)

	if _, err := c.Config.Write(ctx, val, false); err != nil {
		c.Errorf("Writing Config File: %v", err)
		_, _ = ui.Error(TitleError, "Writing Config File: "+err.Error())

		return
	}

	_, _ = ui.Info(mnd.Title, "Wrote Config File: "+val)
}

func (c *Client) menuPanic() {
	defer c.CapturePanic()

	yes, err := ui.Question(mnd.Title, "You really want to panic?", true)
	if !yes || err != nil {
		return
	}

	defer c.Printf("User Requested Application Panic, good bye.")
	panic("user requested panic")
}

func (c *Client) openGUI() {
	uri := "http://127.0.0.1"
	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		uri = "https://127.0.0.1"
	}

	// This always has a colon, or the app will not start.
	port := strings.Split(c.Config.BindAddr, ":")[1]
	go ui.OpenURL(uri + ":" + port + c.Config.URLBase) //nolint:errcheck
}

func (c *Client) updatePassword(ctx context.Context) {
	pass, _, err := ui.Entry(mnd.Title, "Enter new Web UI admin password (must be 9+ characters):", "")
	if err != nil {
		c.Errorf("err: %v", err)
		return
	}

	if err := c.StopWebServer(ctx); err != nil {
		c.Errorf("Stopping web server: %v", err)

		if err = ui.Notify("Stopping web server failed, password not updated."); err != nil {
			c.Errorf("Creating Toast Notification: %v", err)
		}

		return
	}

	c.Print("[user requested] Updating Web UI password.")

	defer c.StartWebServer(ctx)

	if err := c.Config.UIPassword.Set(configfile.DefaultUsername + ":" + pass); err != nil {
		c.Errorf("Updating Web UI Password: %v", err)
		_, _ = ui.Error(TitleError, "Updating Web UI Password: "+err.Error())
	}

	if err = ui.Notify("Web UI password updated. Save config to persist this change."); err != nil {
		c.Errorf("Creating Toast Notification: %v", err)
	}
}
