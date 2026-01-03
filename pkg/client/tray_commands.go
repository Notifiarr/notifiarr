//go:build darwin || windows || linux

package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/configfile"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
)

/* This file contains methods that are triggered from the GUI menu. */

func (c *Client) toggleServer(ctx context.Context) {
	if !menu["stat"].Checked() {
		ui.Toast(ctx, "Started web server") //nolint:errcheck
		logs.Log.Printf("[user requested] Starting Web Server, baseurl: %s, bind address: %s",
			c.Config.URLBase, c.Config.BindAddr)
		go c.RunWebServer()

		return
	}

	ui.Toast(ctx, "Paused web server") //nolint:errcheck
	logs.Log.Printf("[user requested] Pausing Web Server")

	if err := c.StopWebServer(ctx); err != nil {
		logs.Log.Errorf("Unable to Pause Server: %v", err)
	}
}

func (c *Client) rotateLogs(ctx context.Context) {
	logs.Log.Printf("[user requested] Rotating Log Files!")
	ui.Toast(ctx, "Rotating log files") //nolint:errcheck

	for _, err := range logs.Log.Rotate() {
		if err != nil {
			ui.Toast(ctx, "Error rotating log files: %v", err) //nolint:errcheck
			logs.Log.Errorf("Rotating Log Files: %v", err)
		}
	}
}

// This is always outdated. :( The format on screen sucks anyway. This should probably be removed.
func (c *Client) displayConfig() string { //nolint: funlen,cyclop
	var out strings.Builder

	out.WriteString("Config File: " + c.Flags.ConfigFile)
	out.WriteString(fmt.Sprintf("\nTimeout: %v", c.Config.Timeout))
	out.WriteString(fmt.Sprintf("\nUpstreams: %v", c.allow.Input))

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		out.WriteString(fmt.Sprintf("\nHTTPS: https://%s%s", c.Config.BindAddr, c.Config.URLBase))
		out.WriteString(fmt.Sprintf("\nCert File: %v", c.Config.SSLCrtFile))
		out.WriteString(fmt.Sprintf("\nCert Key: %v", c.Config.SSLKeyFile))
	} else {
		out.WriteString(fmt.Sprintf("\nHTTP: http://%s%s", c.Config.BindAddr, c.Config.URLBase))
	}

	if c.Config.LogFiles > 0 {
		out.WriteString(fmt.Sprintf("\nLog File: %v (%d @ %dMb)", c.Config.LogFile, c.Config.LogFiles, c.Config.LogFileMb))
		out.WriteString(fmt.Sprintf("\nHTTP Log: %v (%d @ %dMb)", c.Config.HTTPLog, c.Config.LogFiles, c.Config.LogFileMb))
	} else {
		out.WriteString(fmt.Sprintf("\nLog File: %v (no rotation)", c.Config.LogFile))
		out.WriteString(fmt.Sprintf("\nHTTP Log: %v (no rotation)", c.Config.HTTPLog))
	}

	if count := len(c.Config.Lidarr); count == 1 {
		out.WriteString(fmt.Sprintf("\n- Lidarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Lidarr[0].URL, c.Config.Lidarr[0].APIKey != "", c.Config.Lidarr[0].Timeout, c.Config.Lidarr[0].ValidSSL))
	} else {
		for _, f := range c.Config.Lidarr {
			out.WriteString(fmt.Sprintf("\n- Lidarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL))
		}
	}

	if count := len(c.Config.Radarr); count == 1 {
		out.WriteString(fmt.Sprintf("\n- Radarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Radarr[0].URL, c.Config.Radarr[0].APIKey != "", c.Config.Radarr[0].Timeout, c.Config.Radarr[0].ValidSSL))
	} else {
		for _, f := range c.Config.Radarr {
			out.WriteString(fmt.Sprintf("\n- Radarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL))
		}
	}

	if count := len(c.Config.Readarr); count == 1 {
		out.WriteString(fmt.Sprintf("\n- Readarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Readarr[0].URL, c.Config.Readarr[0].APIKey != "", c.Config.Readarr[0].Timeout, c.Config.Readarr[0].ValidSSL))
	} else {
		for _, f := range c.Config.Readarr {
			out.WriteString(fmt.Sprintf("\n- Readarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL))
		}
	}

	if count := len(c.Config.Sonarr); count == 1 {
		out.WriteString(fmt.Sprintf("\n- Sonarr Config: 1 server: %s, apikey:%v, timeout:%v, verify ssl:%v",
			c.Config.Sonarr[0].URL, c.Config.Sonarr[0].APIKey != "", c.Config.Sonarr[0].Timeout, c.Config.Sonarr[0].ValidSSL))
	} else {
		for _, f := range c.Config.Sonarr {
			out.WriteString(fmt.Sprintf("\n- Sonarr Server: %s, apikey:%v, timeout:%v, verify ssl:%v",
				f.URL, f.APIKey != "", f.Timeout, f.ValidSSL))
		}
	}

	return out.String() + "\n"
}

func (c *Client) writeConfigFile(ctx context.Context) {
	val, _, _ := ui.Entry("Enter path to write config file:", c.Flags.ConfigFile)

	if val == "" {
		ui.Error("No Config File Provided")
		return
	}

	logs.Log.Print("[user requested] Writing Config File:", val)

	if _, err := c.Config.Write(ctx, val, false); err != nil {
		logs.Log.Errorf("Writing Config File: %v", err)
		ui.Error("Writing Config File: " + err.Error())

		return
	}

	ui.Info("Wrote Config File: " + val)
}

func (c *Client) menuPanic() {
	defer logs.Log.CapturePanic()

	yes, err := ui.Question("You really want to panic?", true)
	if !yes || err != nil {
		return
	}

	defer logs.Log.Printf("User Requested Application Panic, good bye.")
	panic("user requested panic")
}

func (c *Client) openGUI(ctx context.Context) {
	uri := "http://127.0.0.1"
	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		uri = "https://127.0.0.1"
	}

	// This always has a colon, or the app will not start.
	port := strings.Split(c.Config.BindAddr, ":")[1]
	if err := ui.OpenURL(ctx, uri+":"+port+c.Config.URLBase); err != nil {
		logs.Log.Errorf("Opening URL: %v", err)
	}
}

//nolint:errcheck
func (c *Client) autoStart(ctx context.Context) {
	if menu["auto"].Checked() {
		menu["auto"].Uncheck()

		if file, err := ui.DeleteStartupLink(); err != nil {
			ui.Toast(ctx, "Failed disabling autostart: %s", err.Error())
			logs.Log.Errorf("[user requested] Disabling auto start: %v", err)
		} else {
			ui.Toast(ctx, "Removed auto start file: %s", file)
			logs.Log.Printf("[user requested] Removed auto start file: %s", file)
		}

		return
	}

	menu["auto"].Check()

	if loaded, file, err := ui.CreateStartupLink(ctx); err != nil {
		ui.Toast(ctx, "Failed enabling autostart: %s", err.Error())
		logs.Log.Errorf("[user requested] Enabling auto start: %v", err)
	} else if mnd.IsDarwin && !loaded {
		ui.Toast(ctx, "Created auto start file: %s - Exiting so launchctl can restart the app.", file)
		logs.Log.Printf("[user requested] Created auto start file: %s (exiting)", file)
		c.sigkil <- &update.Signal{Text: "launchctl restart"}
	} else {
		ui.Toast(ctx, "Created auto start file: %s", file)
		logs.Log.Printf("[user requested] Created auto start file: %s", file)
	}
}

func (c *Client) updatePassword(ctx context.Context) {
	pass, _, err := ui.Entry("Enter new Web UI admin password (must be 9+ characters):", "")
	if err != nil {
		logs.Log.Errorf("err: %v", err)
		return
	}

	if err := c.StopWebServer(ctx); err != nil {
		logs.Log.Errorf("Stopping web server: %v", err)

		if err = ui.Toast(ctx, "Stopping web server failed, password not updated."); err != nil {
			logs.Log.Errorf("Creating Toast Notification: %v", err)
		}

		return
	}

	logs.Log.Printf("[user requested] Updating Web UI password.")

	if err := c.Config.UIPassword.Set(configfile.DefaultUsername + ":" + pass); err != nil {
		logs.Log.Errorf("Updating Web UI Password: %v", err)
		ui.Error("Updating Web UI Password: " + err.Error())
	}

	if err = ui.Toast(ctx, "Web UI password updated. Save config to persist this change."); err != nil {
		logs.Log.Errorf("Creating Toast Notification: %v", err)
	}

	go c.RunWebServer()
}
