package dnclient

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"syscall"

	"github.com/Go-Lift-TV/discordnotifier-client/bindata"
	"github.com/gen2brain/dlgs"
	"github.com/getlantern/systray"
	"github.com/pkg/browser"
	"golift.io/version"
)

/* This file handles the OS GUI elements. */

func hasGUI() bool {
	switch runtime.GOOS {
	case "darwin":
		return os.Getenv("USEGUI") == "true"
	case "windows":
		return true
	default:
		return false
	}
}

func (c *Client) startTray() {
	if !hasGUI() {
		return
	}

	os.Stdout.Close()

	quitFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout.Duration)
		defer cancel()

		if c.server != nil {
			if err := c.server.Shutdown(ctx); err != nil {
				c.Print("[ERROR]", err)
				os.Exit(1) // web server problem
			}
		}

		os.Exit(0)
	}
	systray.Run(c.readyTray, quitFn)
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

	c.menu["link"] = systray.AddMenuItem("Links", "external resources")
	c.menu["info"] = c.menu["link"].AddSubMenuItem(c.Flags.Name(), version.Print(c.Flags.Name()))
	c.menu["hp"] = c.menu["link"].AddSubMenuItem("DiscordNotifier.com", "open DiscordNotifier.com")
	c.menu["disc1"] = c.menu["link"].AddSubMenuItem("DiscordNotifier Discord", "open DiscordNotifier discord server")
	c.menu["disc2"] = c.menu["link"].AddSubMenuItem("Go Lift Discord", "open Go Lift discord server")
	c.menu["love"] = c.menu["link"].AddSubMenuItem("<3 ?x?.io", "show some love")
	c.menu["gh"] = c.menu["link"].AddSubMenuItem("GitHub Project", c.Flags.Name()+" on GitHub")
	c.menu["conf"] = systray.AddMenuItem("Config", "show configuration")
	c.menu["key"] = systray.AddMenuItem("API Key", "set API Key")
	c.menu["logs"] = systray.AddMenuItem("Logs", "show log file")
	c.menu["load"] = systray.AddMenuItem("Reload", "reload configuration")
	c.menu["exit"] = systray.AddMenuItem("Quit", "Exit "+c.Flags.Name())

	c.menu["info"].Disable()
	c.watchGuiChannels()
}

func (c *Client) watchGuiChannels() {
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
		case <-c.menu["exit"].ClickedCh:
			c.Printf("[%s] Need help? %s\n=====> Exiting! User Requested", c.Flags.Name(), helpLink)
			systray.Quit()
		case <-c.menu["info"].ClickedCh:
		case <-c.menu["link"].ClickedCh:
		case <-c.menu["gh"].ClickedCh:
			browser.OpenURL("https://github.com/Go-Lift-TV/discordnotifier-client/")
		case <-c.menu["hp"].ClickedCh:
			browser.OpenURL("https://discordnotifier.com/")
		case <-c.menu["logs"].ClickedCh:
			browser.OpenFile(c.Config.LogFile)
		case <-c.menu["disc1"].ClickedCh:
			browser.OpenURL("https://discord.gg/AURf8Yz")
		case <-c.menu["disc2"].ClickedCh:
			browser.OpenURL("https://golift.io/discord")
		case <-c.menu["love"].ClickedCh:
			dlgs.Warning(Title, "nitusa loves you!\n<3")
		case <-c.menu["conf"].ClickedCh:
			dlgs.Info(Title+": Configuration", c.displayConfig())
		case <-c.menu["load"].ClickedCh:
			c.reloadConfiguration()
		case <-c.menu["key"].ClickedCh:
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
	s += fmt.Sprintf("\nUpstreams: %v", c.allow.combineUpstreams())

	if c.Config.SSLCrtFile != "" && c.Config.SSLKeyFile != "" {
		s += fmt.Sprintf("\nHTTPS: https://%s%s", c.Config.BindAddr, path.Join("/", c.Config.URLBase))
		s += fmt.Sprintf("\nCert File: %v", c.Config.SSLCrtFile)
		s += fmt.Sprintf("\nCert Key: %v", c.Config.SSLKeyFile)
	} else {
		s += fmt.Sprintf("\nHTTP: http://%s%s", c.Config.BindAddr, path.Join("/", c.Config.URLBase))
	}

	/*
	   c.initSonarr()
	   c.initRadarr()
	   c.initLidarr()
	   c.initReadarr()
	*/
	s += fmt.Sprintf("\nLidarr: %v", len(c.Config.Lidarr))
	s += fmt.Sprintf("\nSonarr: %v", len(c.Config.Sonarr))
	s += fmt.Sprintf("\nRadarr: %v", len(c.Config.Radarr))
	s += fmt.Sprintf("\nReadarr: %v", len(c.Config.Readarr))

	if c.Config.LogFile != "" {
		if c.Config.LogFiles > 0 {
			s += fmt.Sprintf("\nLog File: %v (%d @ %dMb)", c.Config.LogFile, c.Config.LogFiles, c.Config.LogFileMb)
		} else {
			s += fmt.Sprintf("\nLog File: %v (no rotation)", c.Config.LogFile)
		}
	}

	return s + "\n"
}
