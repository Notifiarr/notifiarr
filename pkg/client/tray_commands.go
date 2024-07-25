//go:build darwin || windows || linux

package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
)

/* This file contains methods that are triggered from the GUI menu. */

func (c *Client) toggleServer(ctx context.Context) {
	if !menu["stat"].Checked() {
		ui.Toast("Started web server") //nolint:errcheck
		c.Printf("[user requested] Starting Web Server, baseurl: %s, bind address: %s",
			c.Config.URLBase, c.Config.BindAddr)
		c.StartWebServer(ctx)

		return
	}

	ui.Toast("Paused web server") //nolint:errcheck
	c.Print("[user requested] Pausing Web Server")

	if err := c.StopWebServer(ctx); err != nil {
		c.Errorf("Unable to Pause Server: %v", err)
	}
}

func (c *Client) rotateLogs() {
	c.Print("[user requested] Rotating Log Files!")
	ui.Toast("Rotating log files") //nolint:errcheck

	for _, err := range c.Logger.Rotate() {
		if err != nil {
			ui.Toast("Error rotating log files: %v", err) //nolint:errcheck
			c.Errorf("Rotating Log Files: %v", err)
		}
	}
}

// This is always outdated. :( The format on screen sucks anyway. This should probably be removed.
func (c *Client) displayConfig() string { //nolint: funlen,cyclop
	out := "Config File: " + c.Flags.ConfigFile
	out += fmt.Sprintf("\nTimeout: %v", c.Config.Timeout)
	out += fmt.Sprintf("\nUpstreams: %v", c.Config.Allow.Input)

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		out += fmt.Sprintf("\nHTTPS: https://%s%s", c.Config.BindAddr, c.Config.URLBase)
		out += fmt.Sprintf("\nCert File: %v", c.Config.SSLCrtFile)
		out += fmt.Sprintf("\nCert Key: %v", c.Config.SSLKeyFile)
	} else {
		out += fmt.Sprintf("\nHTTP: http://%s%s", c.Config.BindAddr, c.Config.URLBase)
	}

	if c.Config.LogFiles > 0 {
		out += fmt.Sprintf("\nLog File: %v (%d @ %dMb)", c.Config.LogFile, c.Config.LogFiles, c.Config.LogFileMb)
		out += fmt.Sprintf("\nHTTP Log: %v (%d @ %dMb)", c.Config.HTTPLog, c.Config.LogFiles, c.Config.LogFileMb)
	} else {
		out += fmt.Sprintf("\nLog File: %v (no rotation)", c.Config.LogFile)
		out += fmt.Sprintf("\nHTTP Log: %v (no rotation)", c.Config.HTTPLog)
	}

	if count := len(c.Config.Lidarr); count == 1 {
		out += fmt.Sprintf("\n- Lidarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Lidarr[0].URL, c.Config.Lidarr[0].APIKey != "", c.Config.Lidarr[0].Timeout, c.Config.Lidarr[0].ValidSSL)
	} else {
		for _, f := range c.Config.Lidarr {
			out += fmt.Sprintf("\n- Lidarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}

	if count := len(c.Config.Radarr); count == 1 {
		out += fmt.Sprintf("\n- Radarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Radarr[0].URL, c.Config.Radarr[0].APIKey != "", c.Config.Radarr[0].Timeout, c.Config.Radarr[0].ValidSSL)
	} else {
		for _, f := range c.Config.Radarr {
			out += fmt.Sprintf("\n- Radarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}

	if count := len(c.Config.Readarr); count == 1 {
		out += fmt.Sprintf("\n- Readarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Readarr[0].URL, c.Config.Readarr[0].APIKey != "", c.Config.Readarr[0].Timeout, c.Config.Readarr[0].ValidSSL)
	} else {
		for _, f := range c.Config.Readarr {
			out += fmt.Sprintf("\n- Readarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}

	if count := len(c.Config.Sonarr); count == 1 {
		out += fmt.Sprintf("\n- Sonarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Sonarr[0].URL, c.Config.Sonarr[0].APIKey != "", c.Config.Sonarr[0].Timeout, c.Config.Sonarr[0].ValidSSL)
	} else {
		for _, f := range c.Config.Sonarr {
			out += fmt.Sprintf("\n- Sonarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL)
		}
	}

	return out + "\n"
}

func (c *Client) writeConfigFile(ctx context.Context) {
	val, _, _ := ui.Entry("Enter path to write config file:", c.Flags.ConfigFile)

	if val == "" {
		_, _ = ui.Error("No Config File Provided")
		return
	}

	c.Print("[user requested] Writing Config File:", val)

	if _, err := c.Config.Write(ctx, val, false); err != nil {
		c.Errorf("Writing Config File: %v", err)
		_, _ = ui.Error("Writing Config File: " + err.Error())

		return
	}

	_, _ = ui.Info("Wrote Config File: " + val)
}

func (c *Client) menuPanic() {
	defer c.CapturePanic()

	yes, err := ui.Question("You really want to panic?", true)
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
	if err := ui.OpenURL(uri + ":" + port + c.Config.URLBase); err != nil {
		c.Errorf("Opening URL: %v", err)
	}
}

//nolint:errcheck
func (c *Client) autoStart() {
	if menu["auto"].Checked() {
		menu["auto"].Uncheck()

		if file, err := ui.DeleteStartupLink(); err != nil {
			ui.Toast("Failed disabling autostart: %s", err.Error())
			c.Errorf("[user requested] Disabling auto start: %v", err)
		} else {
			ui.Toast("Removed auto start file: %s", file)
			c.Printf("[user requested] Removed auto start file: %s", file)
		}

		return
	}

	menu["auto"].Check()

	if loaded, file, err := ui.CreateStartupLink(); err != nil {
		ui.Toast("Failed enabling autostart: %s", err.Error())
		c.Errorf("[user requested] Enabling auto start: %v", err)
	} else if mnd.IsDarwin && !loaded {
		ui.Toast("Created auto start file: %s - Exiting so launchctl can restart the app.", file)
		c.Printf("[user requested] Created auto start file: %s (exiting)", file)
		c.sigkil <- &update.Signal{Text: "launchctl restart"}
	} else {
		ui.Toast("Created auto start file: %s", file)
		c.Printf("[user requested] Created auto start file: %s", file)
	}
}

func (c *Client) updatePassword(ctx context.Context) {
	pass, _, err := ui.Entry("Enter new Web UI admin password (must be 9+ characters):", "")
	if err != nil {
		c.Errorf("err: %v", err)
		return
	}

	if err := c.StopWebServer(ctx); err != nil {
		c.Errorf("Stopping web server: %v", err)

		if err = ui.Toast("Stopping web server failed, password not updated."); err != nil {
			c.Errorf("Creating Toast Notification: %v", err)
		}

		return
	}

	c.Print("[user requested] Updating Web UI password.")

	defer c.StartWebServer(ctx)

	if err := c.Config.UIPassword.Set(configfile.DefaultUsername + ":" + pass); err != nil {
		c.Errorf("Updating Web UI Password: %v", err)
		_, _ = ui.Error("Updating Web UI Password: " + err.Error())
	}

	if err = ui.Toast("Web UI password updated. Save config to persist this change."); err != nil {
		c.Errorf("Creating Toast Notification: %v", err)
	}
}
