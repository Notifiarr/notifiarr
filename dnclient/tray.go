// +build darwin windows

package dnclient

import (
	"context"
	"fmt"
	"os"
	"path"
	"syscall"

	"github.com/Go-Lift-TV/discordnotifier-client/bindata"
	"github.com/Go-Lift-TV/discordnotifier-client/ui"
	"github.com/getlantern/systray"
	"golift.io/version"
)

/* This file handles the OS GUI elements. */

func (c *Client) startTray() error {
	if !ui.HasGUI() {
		return c.Exit()
	}

	os.Stdout.Close()

	systray.Run(c.readyTray, func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout.Duration)
		defer cancel()

		if c.server != nil {
			if err := c.server.Shutdown(ctx); err != nil {
				c.Errorf("shutting down web server: %v", err)
				os.Exit(1) // web server problem
			}
		}

		os.Exit(0)
	})

	// We never get here, but just in case the library changes...
	return c.Exit()
}

func (c *Client) readyTray() {
	b, err := bindata.Asset(ui.SystrayIcon)
	if err == nil {
		systray.SetTemplateIcon(b, b)
	} else {
		c.Errorf("reading icon: %v", err)
		systray.SetTitle("DNC")
	}

	systray.SetTooltip(c.Flags.Name())
	c.menu["stat"] = ui.WrapMenu(systray.AddMenuItem("Running", "web server running, uncheck to pause"))

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

	c.menu["update"] = ui.WrapMenu(systray.AddMenuItem("Update", "there is a newer version available"))
	c.menu["dninfo"] = ui.WrapMenu(systray.AddMenuItem("Info", "info from DiscordNotifier.com"))
	c.menu["exit"] = ui.WrapMenu(systray.AddMenuItem("Quit", "Exit "+c.Flags.Name()))

	c.menu["dninfo"].Hide()
	c.menu["update"].Hide()
	c.menu["info"].Disable()
	c.menu["stat"].Check()
	c.watchGuiChannels()
}

// nolint:errcheck
func (c *Client) watchGuiChannels() {
	for {
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
			ui.OpenFile(c.Flags.ConfigFile)
		case <-c.menu["load"].Clicked():
			c.reloadConfiguration()
		case <-c.menu["key"].Clicked():
			c.changeKey()
		case <-c.menu["logs_view"].Clicked():
			ui.OpenLog(c.Config.LogFile)
		case <-c.menu["logs_http"].Clicked():
			ui.OpenLog(c.Config.HTTPLog)
		case <-c.menu["logs_rotate"].Clicked():
			c.rotateLogs()
		case <-c.menu["update"].Clicked():
			ui.OpenURL("https://github.com/Go-Lift-TV/discordnotifier-client/releases")
		case <-c.menu["dninfo"].Clicked():
			ui.Info(Title, "INFO: "+c.info)
			c.menu["dninfo"].Hide()
		case sigc := <-c.signal:
			if sigc != syscall.SIGHUP {
				c.Errorf("[%s] Need help? %s\n=====> Exiting! Caught Signal: %v", c.Flags.Name(), helpLink, sigc)
				systray.Quit() // this kills the app.
			}

			c.reloadConfiguration()
		case <-c.menu["exit"].Clicked():
			c.Errorf("[%s] Need help? %s\n=====> Exiting! User Requested", c.Flags.Name(), helpLink)
			systray.Quit() // this kills the app.
		}
	}
}

func (c *Client) toggleServer() {
	if c.server == nil {
		c.Print("Starting Web Server")
		c.StartWebServer()
		c.menu["stat"].Check()
		c.menu["stat"].SetTooltip("web server running, uncheck to pause")
	} else {
		c.Print("Pausing Web Server")
		c.StopWebServer()
		c.menu["stat"].Uncheck()
		c.menu["stat"].SetTooltip("web server paused, click to start")
	}
}

func (c *Client) changeKey() {
	key, ok, err := ui.Entry(Title+": Configuration", "API Key", c.Config.APIKey)
	if err != nil {
		c.Errorf("Updating API Key: %v", err)
	} else if ok && key != c.Config.APIKey {
		// updateKey shuts down the web server and changes the API key.
		// The server has to shut down to avoid race conditions.
		c.Print("Updating API Key!")
		c.RestartWebServer(func() { c.Config.APIKey = key })
	}
}

func (c *Client) rotateLogs() {
	c.Print("Rotating Log Files!")

	if _, err := c.Logger.webrotate.Rotate(); err != nil {
		c.Errorf("Rotating HTTP Log: %v", err)
	}

	if _, err := c.Logger.logrotate.Rotate(); err != nil {
		c.Errorf("Rotating Log: %v", err)
	}
}

func (c *Client) displayConfig() (s string) { //nolint: funlen
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
