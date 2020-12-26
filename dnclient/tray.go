// +build darwin windows

package dnclient

import (
	"context"
	"fmt"
	"os"
	"path"
	"syscall"

	"github.com/Go-Lift-TV/discordnotifier-client/bindata"
	"github.com/gen2brain/dlgs"
	"github.com/getlantern/systray"
	"golift.io/version"
)

/* This file handles the OS GUI elements. */

func (c *Client) startTray() {
	if !hasGUI() {
		return
	}

	os.Stdout.Close()

	systray.Run(c.readyTray, func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout.Duration)
		defer cancel()

		if c.server != nil {
			if err := c.server.Shutdown(ctx); err != nil {
				c.Print("[ERROR]", err)
				os.Exit(1) // web server problem
			}
		}

		os.Exit(0)
	})
}

func (c *Client) readyTray() {
	b, err := bindata.Asset(systrayIcon)
	if err == nil {
		systray.SetTemplateIcon(b, b)
	} else {
		c.Print("[ERROR] reading icon:", err)
		systray.SetTitle("DNC")
	}

	systray.SetTooltip(c.Flags.Name())

	menu := make(map[string]*systray.MenuItem)
	menu["stat"] = systray.AddMenuItem("Running", "web server running, uncheck to pause")
	menu["conf"] = systray.AddMenuItem("Config", "show configuration")
	menu["link"] = systray.AddMenuItem("Links", "external resources")
	menu["info"] = menu["link"].AddSubMenuItem(c.Flags.Name(), version.Print(c.Flags.Name()))
	menu["hp"] = menu["link"].AddSubMenuItem("DiscordNotifier.com", "open DiscordNotifier.com")
	menu["wiki"] = menu["link"].AddSubMenuItem("DiscordNotifier Wiki", "open DiscordNotifier wiki")
	menu["disc1"] = menu["link"].AddSubMenuItem("DiscordNotifier Discord", "open DiscordNotifier discord server")
	menu["disc2"] = menu["link"].AddSubMenuItem("Go Lift Discord", "open Go Lift discord server")
	menu["love"] = menu["link"].AddSubMenuItem("<3 ?x?.io", "show some love")
	menu["gh"] = menu["link"].AddSubMenuItem("GitHub Project", c.Flags.Name()+" on GitHub")
	menu["view"] = menu["conf"].AddSubMenuItem("View", "show configuration")
	menu["edit"] = menu["conf"].AddSubMenuItem("Edit", "edit configuration")
	menu["key"] = menu["conf"].AddSubMenuItem("API Key", "set API Key")
	menu["load"] = menu["conf"].AddSubMenuItem("Reload", "reload configuration")
	menu["logs"] = systray.AddMenuItem("Logs", "show log file")
	menu["exit"] = systray.AddMenuItem("Quit", "Exit "+c.Flags.Name())

	menu["info"].Disable()
	menu["stat"].Check()
	c.watchGuiChannels(menu)
}

func (c *Client) watchGuiChannels(menu map[string]*systray.MenuItem) {
	// nolint:errcheck
	for {
		select {
		case sigc := <-c.signal:
			if sigc == syscall.SIGHUP {
				c.reloadConfiguration()
			} else {
				c.Printf("[%s] Need help? %s\n=====> Exiting! Caught Signal: %v", c.Flags.Name(), helpLink, sigc)
				systray.Quit()
			}
		case <-menu["exit"].ClickedCh:
			c.Printf("[%s] Need help? %s\n=====> Exiting! User Requested", c.Flags.Name(), helpLink)
			systray.Quit()
		case <-menu["stat"].ClickedCh:
			if c.server == nil {
				c.Print("Starting Web Server")
				c.StartWebServer()
				menu["stat"].Check()
				menu["stat"].SetTooltip("web server running, uncheck to pause")
			} else {
				c.Print("Pausing Web Server")
				c.StopWebServer()
				menu["stat"].Uncheck()
				menu["stat"].SetTooltip("web server paused, click to start")
			}
		case <-menu["gh"].ClickedCh:
			openURL("https://github.com/Go-Lift-TV/discordnotifier-client/")
		case <-menu["hp"].ClickedCh:
			openURL("https://discordnotifier.com/")
		case <-menu["wiki"].ClickedCh:
			openURL("https://trash-guides.info/Misc/Discord-Notifier-Basic-Setup/")
		case <-menu["logs"].ClickedCh:
			openLog(c.Config.LogFile)
		case <-menu["disc1"].ClickedCh:
			openURL("https://discord.gg/AURf8Yz")
		case <-menu["disc2"].ClickedCh:
			openURL("https://golift.io/discord")
		case <-menu["love"].ClickedCh:
			dlgs.Warning(Title, "nitusa loves you!\n<3")
		case <-menu["view"].ClickedCh:
			dlgs.Info(Title+": Configuration", c.displayConfig())
		case <-menu["edit"].ClickedCh:
			openFile(c.Flags.ConfigFile)
		case <-menu["load"].ClickedCh:
			c.reloadConfiguration()
		case <-menu["key"].ClickedCh:
			key, ok, err := dlgs.Entry(Title+": Configuration", "API Key", c.Config.APIKey)
			if err != nil {
				c.Print("[ERROR] Updating API Key:", err)
			} else if ok && key != c.Config.APIKey {
				// updateKey shuts down the web server and changes the API key.
				// The server has to shut down to avoid race conditions.
				c.Print("Updating API Key!")
				c.RestartWebServer(func() { c.Config.APIKey = key })
			}
		}
	}
}

func (c *Client) displayConfig() (s string) {
	s = "Config File: " + c.Flags.ConfigFile
	s += fmt.Sprintf("\nDebug: %v", c.Config.Debug)
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
	} else {
		s += fmt.Sprintf("\nLog File: %v (no rotation)", c.Config.LogFile)
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
