//go:build darwin || windows
// +build darwin windows

package client

import (
	"os"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/getlantern/systray"
	"golift.io/version"
)

/* This file handles the OS GUI elements. */

// startTray Run()s readyTray to bring up the web server and the GUI app.
func (c *Client) startTray() {
	systray.Run(func() {
		defer os.Exit(0)
		defer c.CapturePanic()

		b, _ := bindata.Asset(ui.SystrayIcon)
		systray.SetTemplateIcon(b, b)
		systray.SetTooltip(c.Flags.Name() + " v" + version.Version)
		c.makeChannels() // make these before starting the web server.
		c.makeMoreChannels()
		c.setupChannels(c.watchKillerChannels, c.watchNotifiarrMenu,
			c.watchLogsChannels, c.watchConfigChannels, c.watchGuiChannels)

		// This starts the web server, and waits for reload/exit signals.
		if err := c.Exit(); err != nil {
			c.Errorf("Server: %v", err)
			os.Exit(1) // web server problem
		}
	}, func() {
		// This code only fires from menu->quit.
		if err := c.exit(); err != nil {
			c.Errorf("Server: %v", err)
			os.Exit(1) // web server problem
		}
		// because systray wants to control the exit code? no..
		os.Exit(0)
	})
}

// setupChannels runs the channel watcher loops in go routines with a panic catcher.
func (c *Client) setupChannels(funcs ...func()) {
	for _, f := range funcs {
		go func(f func()) {
			defer c.CapturePanic()
			f()
		}(f)
	}
}

func (c *Client) makeChannels() {
	c.menu["stat"] = ui.WrapMenu(systray.AddMenuItem("Running", "web server state unknown"))

	conf := systray.AddMenuItem("Config", "show configuration")
	c.menu["conf"] = ui.WrapMenu(conf)
	c.menu["view"] = ui.WrapMenu(conf.AddSubMenuItem("View", "show configuration"))
	c.menu["edit"] = ui.WrapMenu(conf.AddSubMenuItem("Edit", "edit configuration"))
	c.menu["write"] = ui.WrapMenu(conf.AddSubMenuItem("Write", "write config file"))
	c.menu["key"] = ui.WrapMenu(conf.AddSubMenuItem("API Key", "set API Key"))
	c.menu["load"] = ui.WrapMenu(conf.AddSubMenuItem("Reload", "reload configuration"))

	link := systray.AddMenuItem("Links", "external resources")
	c.menu["link"] = ui.WrapMenu(link)
	c.menu["info"] = ui.WrapMenu(link.AddSubMenuItem(c.Flags.Name(), version.Print(c.Flags.Name())))
	c.menu["info"].Disable()
	c.menu["hp"] = ui.WrapMenu(link.AddSubMenuItem("Notifiarr.com", "open Notifiarr.com"))
	c.menu["wiki"] = ui.WrapMenu(link.AddSubMenuItem("Notifiarr Wiki", "open Notifiarr wiki"))
	c.menu["disc1"] = ui.WrapMenu(link.AddSubMenuItem("Notifiarr Discord", "open Notifiarr discord server"))
	c.menu["disc2"] = ui.WrapMenu(link.AddSubMenuItem("Go Lift Discord", "open Go Lift discord server"))
	c.menu["gh"] = ui.WrapMenu(link.AddSubMenuItem("GitHub Project", c.Flags.Name()+" on GitHub"))

	logs := systray.AddMenuItem("Logs", "log file info")
	c.menu["logs"] = ui.WrapMenu(logs)
	c.menu["logs_view"] = ui.WrapMenu(logs.AddSubMenuItem("View", "view the application log"))
	c.menu["logs_http"] = ui.WrapMenu(logs.AddSubMenuItem("HTTP", "view the HTTP log"))
	c.menu["logs_svcs"] = ui.WrapMenu(logs.AddSubMenuItem("Services", "view the Services log"))
	c.menu["logs_rotate"] = ui.WrapMenu(logs.AddSubMenuItem("Rotate", "rotate both log files"))

	if c.Config.Services.LogFile == "" {
		c.menu["logs_svcs"].Hide()
	}
}

// makeMoreChannels makes the Notifiarr menu and Debug menu items.
//nolint:lll
func (c *Client) makeMoreChannels() {
	data := systray.AddMenuItem("Notifiarr", "plex sessions, system snapshots, service checks")
	c.menu["data"] = ui.WrapMenu(data)
	c.menu["sync_cf"] = ui.WrapMenu(data.AddSubMenuItem("Sync Custom Formats", "[premium feature] trigger custom format sync"))
	c.menu["snap_log"] = ui.WrapMenu(data.AddSubMenuItem("Log Full Snapshot", "write snapshot data to log file"))
	c.menu["svcs_log"] = ui.WrapMenu(data.AddSubMenuItem("Log Service Checks", "check all services and log results"))
	c.menu["svcs_prod"] = ui.WrapMenu(data.AddSubMenuItem("Check Services", "check all services and send results to notifiarr"))
	c.menu["plex_prod"] = ui.WrapMenu(data.AddSubMenuItem("Plex Sessions", "send plex sessions to notifiarr"))
	c.menu["snap_prod"] = ui.WrapMenu(data.AddSubMenuItem("System Snapshot", "send system snapshot to notifiarr"))
	c.menu["svcs_test"] = ui.WrapMenu(data.AddSubMenuItem("Test Service Checks", "send all service check results to test endpoint"))
	c.menu["plex_test"] = ui.WrapMenu(data.AddSubMenuItem("Test Plex Sessions", "send plex sessions to notifiarr test endpoint"))
	c.menu["snap_test"] = ui.WrapMenu(data.AddSubMenuItem("Test System Snapshot", "send system snapshot to notifiarr test endpoint"))
	c.menu["plex_dev"] = ui.WrapMenu(data.AddSubMenuItem("Dev Plex Sessions", "send plex sessions to notifiarr dev endpoint"))
	c.menu["snap_dev"] = ui.WrapMenu(data.AddSubMenuItem("Dev System Snapshot", "send system snapshot to notifiarr dev endpoint"))
	c.menu["app_ques"] = ui.WrapMenu(data.AddSubMenuItem("Stuck Items Check", "check app queues for stuck items and send to notifiarr"))
	c.menu["app_ques_dev"] = ui.WrapMenu(data.AddSubMenuItem("Stuck Items Check (Dev)", "check app queues for stuck items and send to notifiarr dev"))
	c.menu["send_dash"] = ui.WrapMenu(data.AddSubMenuItem("Send Dashboard States", "collect and send all application states for a dashboard update"))

	debug := systray.AddMenuItem("Debug", "Debug Menu")
	c.menu["debug"] = ui.WrapMenu(debug)
	c.menu["debug_logs"] = ui.WrapMenu(debug.AddSubMenuItem("View Log", "view the Debug log"))
	// debug.AddSeparator() // not exist: https://github.com/getlantern/systray/issues/170
	ui.WrapMenu(debug.AddSubMenuItem("__________", "")).Disable() // fake separator.
	c.menu["debug_panic"] = ui.WrapMenu(debug.AddSubMenuItem("Panic", "cause an application panic (crash)"))

	if c.Config.LogConfig.DebugLog == "" {
		c.menu["debug_logs"].Hide()
	}

	if !c.Config.Debug {
		c.menu["svcs_test"].Hide()
		c.menu["plex_test"].Hide()
		c.menu["snap_test"].Hide()
		c.menu["plex_dev"].Hide()
		c.menu["snap_dev"].Hide()
		c.menu["app_ques_dev"].Hide()
		c.menu["debug"].Hide()
	}

	c.menu["update"] = ui.WrapMenu(systray.AddMenuItem("Update", "Check GitHub for Update"))
	c.menu["exit"] = ui.WrapMenu(systray.AddMenuItem("Quit", "Exit "+c.Flags.Name()))
}

func (c *Client) watchKillerChannels() {
	defer systray.Quit() // this kills the app

	for {
		select {
		case <-c.menu["exit"].Clicked():
			c.Errorf("Need help? %s\n=====> Exiting! User Requested", mnd.HelpLink)
			return
		case <-c.menu["debug"].Clicked():
			// turn on and off debug?
			// u.menu["debug"].Check()
		case <-c.menu["debug_panic"].Clicked():
			c.menuPanic()
		case <-c.menu["debug_logs"].Clicked():
			go ui.OpenLog(c.Config.LogConfig.DebugLog) // nolint:errcheck
			c.Print("User Viewing Debug File:", c.Config.LogConfig.DebugLog)
		case <-c.menu["load"].Clicked():
			if err := c.reloadConfiguration("User Requested"); err != nil {
				c.Errorf("Need help? %s\n=====> Exiting! Reloading Configuration: %v", mnd.HelpLink, err)
				os.Exit(1) //nolint:gocritic // exit now since config is bad and everything is disabled.
			}
		}
	}
}

// nolint:errcheck
func (c *Client) watchGuiChannels() {
	for {
		select {
		case <-c.menu["stat"].Clicked():
			c.toggleServer()
		case <-c.menu["gh"].Clicked():
			go ui.OpenURL("https://github.com/Notifiarr/notifiarr/")
		case <-c.menu["hp"].Clicked():
			go ui.OpenURL("https://notifiarr.com/")
		case <-c.menu["wiki"].Clicked():
			go ui.OpenURL("https://trash-guides.info/Notifiarr/Quick-Start/")
		case <-c.menu["disc1"].Clicked():
			go ui.OpenURL("https://notifiarr.com/discord")
		case <-c.menu["disc2"].Clicked():
			go ui.OpenURL("https://golift.io/discord")
		}
	}
}

// nolint:errcheck
func (c *Client) watchConfigChannels() {
	for {
		select {
		case <-c.menu["view"].Clicked():
			go ui.Info(mnd.Title+": Configuration", c.displayConfig())
		case <-c.menu["edit"].Clicked():
			go ui.OpenFile(c.Flags.ConfigFile)
			c.Print("User Editing Config File:", c.Flags.ConfigFile)
		case <-c.menu["write"].Clicked():
			go c.writeConfigFile()
		case <-c.menu["key"].Clicked():
			c.changeKey()
		}
	}
}

// nolint:errcheck
func (c *Client) watchLogsChannels() {
	for {
		select {
		case <-c.menu["logs_view"].Clicked():
			go ui.OpenLog(c.Config.LogFile)
			c.Print("User Viewing Log File:", c.Config.LogFile)
		case <-c.menu["logs_http"].Clicked():
			go ui.OpenLog(c.Config.HTTPLog)
			c.Print("User Viewing Log File:", c.Config.HTTPLog)
		case <-c.menu["logs_svcs"].Clicked():
			go ui.OpenLog(c.Config.Services.LogFile)
			c.Print("User Viewing Log File:", c.Config.Services.LogFile)
		case <-c.menu["logs_rotate"].Clicked():
			c.rotateLogs()
		case <-c.menu["update"].Clicked():
			c.checkForUpdate()
		}
	}
}

//nolint:errcheck,cyclop
func (c *Client) watchNotifiarrMenu() {
	for {
		select {
		case <-c.menu["sync_cf"].Clicked():
			ui.Notify("Starting custom format and quality profiles sync")
			c.Printf("[user requested] Triggering Custom Formats and Quality Profiles Sync for Radarr and Sonarr.")
			c.notifiarr.Trigger.SyncCF(false)
		case <-c.menu["snap_log"].Clicked():
			ui.Notify("Logging local system snapshot")
			c.logSnaps()
		case <-c.menu["svcs_log"].Clicked():
			c.Printf("[user requested] Checking services and logging results.")
			ui.Notify("Running and logging %d Service Checks", len(c.Config.Service))
			c.Config.Services.RunChecks(&services.Source{Name: "log", URL: ""})
		case <-c.menu["svcs_prod"].Clicked():
			c.Printf("[user requested] Checking services and sending results to Notifiarr.")
			ui.Notify("Running and sending %d Service Checks", len(c.Config.Service))
			c.Config.Services.RunChecks(&services.Source{Name: "user", URL: notifiarr.ProdURL})
		case <-c.menu["svcs_test"].Clicked():
			c.Printf("[user requested] Checking services and sending results to Notifiarr Test.")
			ui.Notify("Running and sending %d Service Checks (test)", len(c.Config.Service))
			c.Config.Services.RunChecks(&services.Source{Name: "user", URL: notifiarr.TestURL})
		case <-c.menu["plex_test"].Clicked():
			ui.Notify("Gathering and sending Plex sessions (test)")
			c.sendPlexSessions(notifiarr.TestURL)
		case <-c.menu["snap_test"].Clicked():
			ui.Notify("Gathering and sending system snapshot (test)")
			c.sendSystemSnapshot(notifiarr.TestURL)
		case <-c.menu["plex_dev"].Clicked():
			ui.Notify("Gathering and sending Plex sessions (dev)")
			c.sendPlexSessions(notifiarr.DevURL)
		case <-c.menu["snap_dev"].Clicked():
			ui.Notify("Gathering and sending system snapshot (dev)")
			c.sendSystemSnapshot(notifiarr.DevURL)
		case <-c.menu["app_ques"].Clicked():
			ui.Notify("Sending finished, possibly stuck, queue items")
			c.notifiarr.Trigger.SendFinishedQueueItems(notifiarr.BaseURL)
		case <-c.menu["app_ques_dev"].Clicked():
			ui.Notify("Sending finished, possibly stuck, queue items (dev)")
			c.notifiarr.Trigger.SendFinishedQueueItems(notifiarr.DevBaseURL)
		case <-c.menu["plex_prod"].Clicked():
			ui.Notify("Gathering and sending Plex sessions")
			c.sendPlexSessions(notifiarr.ProdURL)
		case <-c.menu["snap_prod"].Clicked():
			ui.Notify("Gathering and sending system snapshot")
			c.sendSystemSnapshot(notifiarr.ProdURL)
		case <-c.menu["send_dash"].Clicked():
			ui.Notify("Gathering and sending app states for dashboard")
			c.Print("User Requested State Collection for Dashboard")
			c.notifiarr.Trigger.GetState()
		}
	}
}
