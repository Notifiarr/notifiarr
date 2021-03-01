// +build darwin windows

package client

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/bindata"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/snapshot"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/ui"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/update"
	"github.com/getlantern/systray"
	"github.com/hako/durafmt"
	"golift.io/version"
)

/* This file handles the OS GUI elements. */

// startTray Run()s readyTray to bring up the web server and the GUI app.
func (c *Client) startTray() {
	systray.Run(c.readyTray, c.exitTray)
}

func (c *Client) exitTray() {
	c.sigkil = nil

	if err := c.Exit(); err != nil {
		c.Errorf("Shutting down web server: %v", err)
		os.Exit(1) // web server problem
	}
	// because systray wants to control the exit code? no..
	os.Exit(0)
}

// readyTray creates the system tray/menu bar app items, and starts the web server.
func (c *Client) readyTray() {
	b, err := bindata.Asset(ui.SystrayIcon)
	if err == nil {
		systray.SetTemplateIcon(b, b)
	} else {
		c.Errorf("Reading Icon: %v", err)
		systray.SetTitle("DNC")
	}

	systray.SetTooltip(c.Flags.Name() + " v" + version.Version)

	c.makeChannels() // make these before starting the web server.
	c.menu["info"].Disable()
	c.menu["dninfo"].Hide()
	c.menu["alert"].Hide() // currently unused.

	go c.watchKillerChannels()
	c.StartWebServer()
	c.watchGuiChannels()
}

func (c *Client) makeChannels() {
	c.menu["stat"] = ui.WrapMenu(systray.AddMenuItem("Running", "web server state unknown"))

	conf := systray.AddMenuItem("Config", "show configuration")
	c.menu["conf"] = ui.WrapMenu(conf)
	c.menu["view"] = ui.WrapMenu(conf.AddSubMenuItem("View", "show configuration"))
	c.menu["edit"] = ui.WrapMenu(conf.AddSubMenuItem("Edit", "edit configuration"))
	c.menu["key"] = ui.WrapMenu(conf.AddSubMenuItem("API Key", "set API Key"))
	c.menu["load"] = ui.WrapMenu(conf.AddSubMenuItem("Reload", "reload configuration"))

	link := systray.AddMenuItem("Links", "external resources")
	c.menu["link"] = ui.WrapMenu(link)
	c.menu["info"] = ui.WrapMenu(link.AddSubMenuItem(c.Flags.Name(), version.Print(c.Flags.Name())))
	c.menu["hp"] = ui.WrapMenu(link.AddSubMenuItem("DiscordNotifier.com", "open DiscordNotifier.com"))
	c.menu["wiki"] = ui.WrapMenu(link.AddSubMenuItem("DiscordNotifier Wiki", "open DiscordNotifier wiki"))
	c.menu["disc1"] = ui.WrapMenu(link.AddSubMenuItem("DiscordNotifier Discord", "open DiscordNotifier discord server"))
	c.menu["disc2"] = ui.WrapMenu(link.AddSubMenuItem("Go Lift Discord", "open Go Lift discord server"))
	c.menu["gh"] = ui.WrapMenu(link.AddSubMenuItem("GitHub Project", c.Flags.Name()+" on GitHub"))

	logs := systray.AddMenuItem("Logs", "log file info")
	c.menu["logs"] = ui.WrapMenu(logs)
	c.menu["logs_view"] = ui.WrapMenu(logs.AddSubMenuItem("View", "view the application log"))
	c.menu["logs_http"] = ui.WrapMenu(logs.AddSubMenuItem("HTTP", "view the HTTP log"))
	c.menu["logs_rotate"] = ui.WrapMenu(logs.AddSubMenuItem("Rotate", "rotate both log files"))

	data := systray.AddMenuItem("Data", "Data Collection and Snapshots")
	c.menu["data"] = ui.WrapMenu(data)
	c.menu["snap_log"] = ui.WrapMenu(data.AddSubMenuItem("Log Snapshot", "write snapshot data to log file"))
	c.menu["snap_send"] = ui.WrapMenu(data.AddSubMenuItem("Test Snapshot", "send snapshot to notifiarr test endpoint"))

	// These start hidden.
	c.menu["update"] = ui.WrapMenu(systray.AddMenuItem("Update", "Check GitHub for Update"))
	c.menu["dninfo"] = ui.WrapMenu(systray.AddMenuItem("Info!", "info from DiscordNotifier.com"))
	c.menu["alert"] = ui.WrapMenu(systray.AddMenuItem("Alert!", "alert from DiscordNotifier.com"))

	c.menu["exit"] = ui.WrapMenu(systray.AddMenuItem("Quit", "Exit "+c.Flags.Name()))
}

func (c *Client) watchKillerChannels() {
	defer systray.Quit() // this kills the app

	for {
		select {
		case sigc := <-c.sighup:
			c.Printf("Caught Signal: %v (reloading configuration)", sigc)
			c.reloadConfiguration("caught signal " + sigc.String())
		case sigc := <-c.sigkil:
			c.Errorf("Need help? %s\n=====> Exiting! Caught Signal: %v", helpLink, sigc)
			return
		case <-c.menu["exit"].Clicked():
			c.Errorf("Need help? %s\n=====> Exiting! User Requested", helpLink)
			return
		}
	}
}

func (c *Client) watchGuiChannels() {
	for {
		// nolint:errcheck
		select {
		case <-c.menu["stat"].Clicked():
			c.toggleServer()
		case <-c.menu["gh"].Clicked():
			ui.OpenURL("https://github.com/Go-Lift-TV/discordnotifier-client/")
		case <-c.menu["hp"].Clicked():
			ui.OpenURL("https://discordnotifier.com/")
		case <-c.menu["wiki"].Clicked():
			ui.OpenURL("https://trash-guides.info/Misc/Discord-Notifier-Basic-Setup/")
		case <-c.menu["disc1"].Clicked():
			ui.OpenURL("https://discord.gg/AURf8Yz")
		case <-c.menu["disc2"].Clicked():
			ui.OpenURL("https://golift.io/discord")
		case <-c.menu["view"].Clicked():
			ui.Info(Title+": Configuration", c.displayConfig())
		case <-c.menu["edit"].Clicked():
			c.Print("User Editing Config File:", c.Flags.ConfigFile)
			ui.OpenFile(c.Flags.ConfigFile)
		case <-c.menu["load"].Clicked():
			c.reloadConfiguration("UI requested")
		case <-c.menu["key"].Clicked():
			c.changeKey()
		case <-c.menu["logs_view"].Clicked():
			c.Print("User Viewing Log File:", c.Config.LogFile)
			ui.OpenLog(c.Config.LogFile)
		case <-c.menu["logs_http"].Clicked():
			c.Print("User Viewing Log File:", c.Config.HTTPLog)
			ui.OpenLog(c.Config.HTTPLog)
		case <-c.menu["logs_rotate"].Clicked():
			c.rotateLogs()
		case <-c.menu["snap_log"].Clicked():
			c.testSnaps("")
		case <-c.menu["snap_send"].Clicked():
			c.testSnaps(snapshot.NotifiarrTestURL)
		case <-c.menu["update"].Clicked():
			c.checkForUpdate()
		case <-c.menu["dninfo"].Clicked():
			c.menu["dninfo"].Hide()
			ui.Info(Title, "INFO: "+c.info)
		}
	}
}

func (c *Client) toggleServer() {
	if c.server == nil {
		c.Print("Starting Web Server")
		c.StartWebServer()

		return
	}

	c.Print("Pausing Web Server")

	if err := c.StopWebServer(); err != nil {
		c.Errorf("Unable to Pause Server: %v", err)
	}
}

func (c *Client) rotateLogs() {
	c.Print("Rotating Log Files!")

	for _, err := range c.Logger.Rotate() {
		if err != nil {
			c.Errorf("Rotating Log Files: %v", err)
		}
	}
}

// changeKey shuts down the web server and changes the API key.
// The server has to shut down to avoid race conditions.
func (c *Client) changeKey() {
	key, ok, err := ui.Entry(Title+": Configuration", "API Key", c.Config.APIKey)
	if err != nil {
		c.Errorf("Updating API Key: %v", err)
	} else if !ok || key == c.Config.APIKey {
		return
	}

	c.Print("Updating API Key!")

	if err := c.StopWebServer(); err != nil && !errors.Is(err, ErrNoServer) {
		c.Errorf("Unable to update API Key: %v", err)
		return
	} else if !errors.Is(err, ErrNoServer) {
		defer c.StartWebServer()
	}

	c.Config.APIKey = key
}

func (c *Client) checkForUpdate() {
	c.Print("User Requested Update Check")

	switch update, err := update.Check("Go-Lift-TV/discordnotifier-client", version.Version); {
	case err != nil:
		c.Errorf("Update Check: %v", err)
		_, _ = ui.Error(Title, "Failure checking version on GitHub: "+err.Error())
	case update.Outdate:
		yes, _ := ui.Question(Title, "An Update is available! Download?\n\n"+
			"Your Version: "+update.Version+"\n"+
			"New Version: "+update.Current+"\n"+
			"Date: "+update.RelDate.Format("Jan 2, 2006")+" ("+
			durafmt.Parse(time.Since(update.RelDate).Round(time.Hour)).String()+" ago)", false)
		if yes {
			_ = ui.OpenURL(update.CurrURL)
		}
	default:
		_, _ = ui.Info(Title, "You're up to date! Version: "+update.Version+"\n"+
			"Updated: "+update.RelDate.Format("Jan 2, 2006")+" ("+
			durafmt.Parse(time.Since(update.RelDate).Round(time.Hour)).String()+" ago)")
	}
}

func (c *Client) displayConfig() (s string) { //nolint: funlen,cyclop
	s = "Config File: " + c.Flags.ConfigFile
	s += fmt.Sprintf("\nTimeout: %v", c.Config.Timeout)
	s += fmt.Sprintf("\nUpstreams: %v", c.allow)

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		s += fmt.Sprintf("\nHTTPS: https://%s%s", c.Config.BindAddr, path.Join("/", c.Config.URLBase))
		s += fmt.Sprintf("\nCert File: %v", c.Config.SSLCrtFile)
		s += fmt.Sprintf("\nCert Key: %v", c.Config.SSLKeyFile)
	} else {
		s += fmt.Sprintf("\nHTTP: http://%s%s", c.Config.BindAddr, path.Join("/", c.Config.URLBase))
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
