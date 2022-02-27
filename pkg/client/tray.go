//go:build darwin || windows

package client

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/notifiarr"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/getlantern/systray"
	"golift.io/starr"
	"golift.io/version"
)

/* This file handles the OS GUI elements. */

// This is arbitrary to avoid conflicts.
const timerPrefix = "TimErr"

// This variable holds all the menu items.
var menu = make(map[string]*systray.MenuItem) //nolint:gochecknoglobals

// startTray Run()s readyTray to bring up the web server and the GUI app.
func (c *Client) startTray(clientInfo *notifiarr.ClientInfo) {
	systray.Run(func() {
		defer os.Exit(0)
		defer c.CapturePanic()

		b, _ := bindata.Asset(ui.SystrayIcon)
		systray.SetTemplateIcon(b, b)
		systray.SetTooltip(c.Flags.Name() + " v" + version.Version)
		c.makeChannels() // make these before starting the web server.
		c.makeMoreChannels()
		c.setupChannels(c.watchKillerChannels, c.watchNotifiarrMenu, c.watchLogsChannels,
			c.watchConfigChannels, c.watchGuiChannels, c.watchTopChannels)
		c.setupMenus(clientInfo)

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

func (c *Client) setupMenus(clientInfo *notifiarr.ClientInfo) {
	if !ui.HasGUI() {
		return
	}

	menu["mode"].SetTitle("Mode: " + strings.Title(c.Config.Mode))

	if !c.Config.Debug {
		menu["debug"].Hide()
	} else {
		menu["debug"].Show()

		if c.Config.LogConfig.DebugLog == "" {
			menu["debug_logs"].Hide()
			menu["debug_logs2"].Hide()
		} else {
			menu["debug_logs"].Show()
			menu["debug_logs2"].Show()
		}
	}

	if c.Config.Services.LogFile == "" {
		menu["logs_svcs"].Hide()
	} else {
		menu["logs_svcs"].Show()
	}

	if !c.Config.Services.Disabled {
		menu["svcs"].Check()
	} else {
		menu["svcs"].Uncheck()
	}

	if clientInfo == nil {
		return
	}

	go c.buildDynamicTimerMenus(clientInfo)

	if clientInfo.IsSub() {
		menu["sub"].SetTitle("Subscriber \u2764\ufe0f")
		menu["sub"].Check()
		menu["sub"].Disable()
		menu["sub"].SetTooltip("THANK YOU for supporting the project!")
	} else if clientInfo.IsPatron() {
		menu["sub"].SetTitle("Patron \U0001f9e1")
		menu["sub"].SetTooltip("THANK YOU for supporting the project!")
		menu["sub"].Check()
	}
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
	menu["stat"] = systray.AddMenuItem("Running", "web server state unknown")

	conf := systray.AddMenuItem("Config", "show configuration")
	menu["conf"] = conf
	menu["view"] = conf.AddSubMenuItem("View", "show configuration")
	menu["edit"] = conf.AddSubMenuItem("Edit", "edit configuration")
	menu["write"] = conf.AddSubMenuItem("Write", "write config file")
	menu["svcs"] = conf.AddSubMenuItem("Services", "toggle service checks routine")
	menu["load"] = conf.AddSubMenuItem("Reload", "reload configuration")

	link := systray.AddMenuItem("Links", "external resources")
	menu["link"] = link
	menu["info"] = link.AddSubMenuItem(c.Flags.Name(), version.Print(c.Flags.Name()))
	menu["info"].Disable()
	menu["hp"] = link.AddSubMenuItem("Notifiarr.com", "open Notifiarr.com")
	menu["wiki"] = link.AddSubMenuItem("Notifiarr.Wiki", "open Notifiarr wiki")
	menu["trash"] = link.AddSubMenuItem("TRaSH Guide", "open TRaSH wiki for Notifiarr")
	menu["disc1"] = link.AddSubMenuItem("Notifiarr Discord", "open Notifiarr discord server")
	menu["disc2"] = link.AddSubMenuItem("Go Lift Discord", "open Go Lift discord server")
	menu["gh"] = link.AddSubMenuItem("GitHub Project", c.Flags.Name()+" on GitHub")

	logs := systray.AddMenuItem("Logs", "log file info")
	menu["logs"] = logs
	menu["logs_view"] = logs.AddSubMenuItem("View", "view the application log")
	menu["logs_http"] = logs.AddSubMenuItem("HTTP", "view the HTTP log")
	menu["debug_logs2"] = logs.AddSubMenuItem("Debug", "view the Debug log")
	menu["logs_svcs"] = logs.AddSubMenuItem("Services", "view the Services log")
	menu["logs_rotate"] = logs.AddSubMenuItem("Rotate", "rotate both log files")
}

// makeMoreChannels makes the Notifiarr menu and Debug menu items.
//nolint:lll
func (c *Client) makeMoreChannels() {
	data := systray.AddMenuItem("Notifiarr", "plex sessions, system snapshots, service checks")
	menu["data"] = data
	menu["gaps"] = data.AddSubMenuItem("Send Radarr Gaps", "[premium feature] trigger radarr collections gaps")
	menu["sync_cf"] = data.AddSubMenuItem("Sync Custom Formats", "[premium feature] trigger custom format sync")
	menu["svcs_prod"] = data.AddSubMenuItem("Check and Send Services", "check all services and send results to notifiarr")
	menu["plex_prod"] = data.AddSubMenuItem("Send Plex Sessions", "send plex sessions to notifiarr")
	menu["snap_prod"] = data.AddSubMenuItem("Send System Snapshot", "send system snapshot to notifiarr")
	menu["app_ques"] = data.AddSubMenuItem("Stuck Queue Items Check", "check app queues for stuck items and send to notifiarr")
	menu["send_dash"] = data.AddSubMenuItem("Send Dashboard States", "collect and send all application states for a dashboard update")
	menu["corrLidarr"] = data.AddSubMenuItem("Check Lidarr Backups", "check latest backup database in each instance for corruption")
	menu["corrProwlarr"] = data.AddSubMenuItem("Check Prowlarr Backups", "check latest backup database in each instance for corruption")
	menu["corrRadarr"] = data.AddSubMenuItem("Check Radarr Backups", "check latest backup database in each instance for corruption")
	menu["corrReadarr"] = data.AddSubMenuItem("Check Readarr Backups", "check latest backup database in each instance for corruption")
	menu["corrSonarr"] = data.AddSubMenuItem("Check Sonarr Backups", "check latest backup database in each instance for corruption")
	menu["backLidarr"] = data.AddSubMenuItem("Send Lidarr Backups", "send backup file list for each instance to Notifiarr")
	menu["backProwlarr"] = data.AddSubMenuItem("Send Prowlarr Backups", "send backup file list for each instance to Notifiarr")
	menu["backRadarr"] = data.AddSubMenuItem("Send Radarr Backups", "send backup file list for each instance to Notifiarr")
	menu["backReadarr"] = data.AddSubMenuItem("Send Readarr Backups", "send backup file list for each instance to Notifiarr")
	menu["backSonarr"] = data.AddSubMenuItem("Send Sonarr Backups", "send backup file list for each instance to Notifiarr")
	// custom timers get added onto data after this.

	debug := systray.AddMenuItem("Debug", "Debug Menu")
	menu["debug"] = debug
	menu["mode"] = debug.AddSubMenuItem("Mode: "+strings.Title(c.Config.Mode), "toggle application mode")
	menu["debug_logs"] = debug.AddSubMenuItem("View Debug Log", "view the Debug log")
	menu["svcs_log"] = debug.AddSubMenuItem("Log Service Checks", "check all services and log results")

	debug.AddSubMenuItem("- Danger Zone -", "").Disable()
	menu["debug_panic"] = debug.AddSubMenuItem("Application Panic", "cause an application panic (crash)")
	menu["update"] = systray.AddMenuItem("Update", "check GitHub for updated version")
	menu["sub"] = systray.AddMenuItem("Subscribe", "subscribe for premium features")
	menu["exit"] = systray.AddMenuItem("Quit", "exit "+c.Flags.Name())
}

// Listen to the top-menu-item channels so they don't back up with junk.
func (c *Client) watchTopChannels() {
	for {
		select {
		case <-menu["conf"].ClickedCh: // unused, top menu.
		case <-menu["link"].ClickedCh: // unused, top menu.
		case <-menu["logs"].ClickedCh: // unused, top menu.
		case <-menu["data"].ClickedCh: // unused, top menu.
		case <-menu["debug"].ClickedCh: // unused, top menu.
		}
	}
}

func (c *Client) closeDynamicTimerMenus() {
	for name := range menu {
		if !strings.HasPrefix(name, timerPrefix) || menu[name].ClickedCh == nil {
			continue
		}

		close(menu[name].ClickedCh)
		menu[name].ClickedCh = nil
	}
}

// dynamic & reusable menu items with reflection, anyone?
func (c *Client) buildDynamicTimerMenus(clientInfo *notifiarr.ClientInfo) {
	defer c.CapturePanic()

	if clientInfo == nil || len(clientInfo.Actions.Custom) == 0 {
		return
	}

	if menu["timerinfo"] == nil {
		menu["timerinfo"] = menu["data"].AddSubMenuItem("- Custom Timers -", "")
	} else {
		// Re-use the already-created menu. This happens after reload.
		menu["timerinfo"].Show()
	}

	menu["timerinfo"].Disable()
	defer menu["timerinfo"].Hide()

	timers := clientInfo.Actions.Custom
	cases := make([]reflect.SelectCase, len(timers))

	for idx, timer := range timers {
		desc := fmt.Sprintf("%s; config: interval: %s, path: %s", timer.Desc, timer.Interval, timer.URI)
		if timer.Desc == "" {
			desc = fmt.Sprintf("dynamic custom timer; config: interval: %s, path: %s", timer.Interval, timer.URI)
		}

		name := timerPrefix + timer.Name
		if menu[name] == nil {
			menu[name] = menu["data"].AddSubMenuItem(timer.Name, desc)
		} else {
			// Re-use the already-created menu. This happens after reload.
			menu[name].ClickedCh = make(chan struct{})
			menu[name].SetTooltip(desc)
		}

		menu[name].Show()
		defer menu[name].Hide()

		cases[idx] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(menu[name].ClickedCh)}
	}

	c.Debugf("Created %d Notifiarr custom timer menu channels.", len(cases))
	defer c.Debugf("All %d Notifiarr custom timer menu channels stopped.", len(cases))

	for {
		if idx, _, ok := reflect.Select(cases); ok {
			timers[idx].Run(notifiarr.EventUser)
		} else if cases = append(cases[:idx], cases[idx+1:]...); len(cases) < 1 {
			// Channel cases[idx] has been closed, remove it.
			return // no menus left to watch, exit.
		}
	}
}

func (c *Client) watchKillerChannels() {
	defer systray.Quit() // this kills the app

	for {
		select {
		case <-menu["exit"].ClickedCh:
			c.Printf("Need help? %s\n=====> Exiting! User Requested", mnd.HelpLink)
			return
		case <-menu["debug_panic"].ClickedCh:
			c.menuPanic()
		case <-menu["load"].ClickedCh:
			c.triggerConfigReload(notifiarr.EventUser, "User Requested")
		}
	}
}

// nolint:errcheck
func (c *Client) watchGuiChannels() {
	for {
		select {
		case <-menu["stat"].ClickedCh:
			c.toggleServer()
		case <-menu["gh"].ClickedCh:
			go ui.OpenURL("https://github.com/Notifiarr/notifiarr/")
		case <-menu["hp"].ClickedCh:
			go ui.OpenURL("https://notifiarr.com/")
		case <-menu["wiki"].ClickedCh:
			go ui.OpenURL("https://notifiarr.wiki/")
		case <-menu["trash"].ClickedCh:
			go ui.OpenURL("https://trash-guides.info/Notifiarr/Quick-Start/")
		case <-menu["disc1"].ClickedCh:
			go ui.OpenURL("https://notifiarr.com/discord")
		case <-menu["disc2"].ClickedCh:
			go ui.OpenURL("https://golift.io/discord")
		case <-menu["sub"].ClickedCh:
			go ui.OpenURL("https://github.com/sponsors/Notifiarr")
		}
	}
}

// nolint:errcheck
func (c *Client) watchConfigChannels() {
	for {
		select {
		case <-menu["view"].ClickedCh:
			go ui.Info(mnd.Title+": Configuration", c.displayConfig())
		case <-menu["edit"].ClickedCh:
			go ui.OpenFile(c.Flags.ConfigFile)
			c.Print("user requested] Editing Config File:", c.Flags.ConfigFile)
		case <-menu["write"].ClickedCh:
			go c.writeConfigFile()
		case <-menu["mode"].ClickedCh:
			if c.Config.Mode == notifiarr.ModeDev {
				c.Config.Mode = c.website.Setup(notifiarr.ModeProd)
			} else {
				c.Config.Mode = c.website.Setup(notifiarr.ModeDev)
			}

			menu["mode"].SetTitle("Mode: " + strings.Title(c.Config.Mode))
			c.Printf("[user requested] Application mode changed to %s!", strings.Title(c.Config.Mode))
			ui.Notify("Application mode changed to %s!", strings.Title(c.Config.Mode))
		case <-menu["svcs"].ClickedCh:
			if menu["svcs"].Checked() {
				menu["svcs"].Uncheck()
				c.Config.Services.Stop()
				ui.Notify("Stopped checking services!")
			} else {
				menu["svcs"].Check()
				c.Config.Services.Start()
				ui.Notify("Service checks started!")
			}
		}
	}
}

// nolint:errcheck
func (c *Client) watchLogsChannels() {
	for {
		select {
		case <-menu["logs_view"].ClickedCh:
			go ui.OpenLog(c.Config.LogFile)
			c.Print("[user requested] Viewing App Log File:", c.Config.LogFile)
		case <-menu["logs_http"].ClickedCh:
			go ui.OpenLog(c.Config.HTTPLog)
			c.Print("[user requested] Viewing HTTP Log File:", c.Config.HTTPLog)
		case <-menu["logs_svcs"].ClickedCh:
			go ui.OpenLog(c.Config.Services.LogFile)
			c.Print("[user requested] Viewing Services Log File:", c.Config.Services.LogFile)
		case <-menu["debug_logs"].ClickedCh:
			go ui.OpenLog(c.Config.LogConfig.DebugLog)
			c.Print("[user requested] Viewing Debug File:", c.Config.LogConfig.DebugLog)
		case <-menu["debug_logs2"].ClickedCh:
			go ui.OpenLog(c.Config.LogConfig.DebugLog)
			c.Print("[user requested] Viewing Debug File:", c.Config.LogConfig.DebugLog)
		case <-menu["logs_rotate"].ClickedCh:
			c.rotateLogs()
		case <-menu["update"].ClickedCh:
			go c.checkForUpdate()
		}
	}
}

//nolint:errcheck,cyclop
func (c *Client) watchNotifiarrMenu() {
	for {
		select {
		case <-menu["gaps"].ClickedCh:
			c.website.Trigger.SendGaps(notifiarr.EventUser)
		case <-menu["sync_cf"].ClickedCh:
			c.website.Trigger.SyncCF(notifiarr.EventUser)
		case <-menu["svcs_log"].ClickedCh:
			c.Print("[user requested] Checking services and logging results.")
			ui.Notify("Running and logging %d Service Checks.", len(c.Config.Service))
			c.Config.Services.RunChecks("log")
		case <-menu["svcs_prod"].ClickedCh:
			c.Print("[user requested] Checking services and sending results to Notifiarr.")
			ui.Notify("Running and sending %d Service Checks.", len(c.Config.Service))
			c.Config.Services.RunChecks(notifiarr.EventUser)
		case <-menu["app_ques"].ClickedCh:
			c.website.Trigger.SendStuckQueueItems(notifiarr.EventUser)
		case <-menu["plex_prod"].ClickedCh:
			c.website.Trigger.SendPlexSessions(notifiarr.EventUser)
		case <-menu["snap_prod"].ClickedCh:
			c.website.Trigger.SendSnapshot(notifiarr.EventUser)
		case <-menu["send_dash"].ClickedCh:
			c.website.Trigger.SendDashboardState(notifiarr.EventUser)
		case <-menu["corrLidarr"].ClickedCh:
			_ = c.website.Trigger.Corruption(notifiarr.EventUser, starr.Lidarr)
		case <-menu["corrProwlarr"].ClickedCh:
			_ = c.website.Trigger.Corruption(notifiarr.EventUser, starr.Prowlarr)
		case <-menu["corrRadarr"].ClickedCh:
			_ = c.website.Trigger.Corruption(notifiarr.EventUser, starr.Radarr)
		case <-menu["corrReadarr"].ClickedCh:
			_ = c.website.Trigger.Corruption(notifiarr.EventUser, starr.Readarr)
		case <-menu["corrSonarr"].ClickedCh:
			_ = c.website.Trigger.Corruption(notifiarr.EventUser, starr.Sonarr)
		case <-menu["backLidarr"].ClickedCh:
			_ = c.website.Trigger.Backup(notifiarr.EventUser, starr.Lidarr)
		case <-menu["backProwlarr"].ClickedCh:
			_ = c.website.Trigger.Backup(notifiarr.EventUser, starr.Prowlarr)
		case <-menu["backRadarr"].ClickedCh:
			_ = c.website.Trigger.Backup(notifiarr.EventUser, starr.Radarr)
		case <-menu["backReadarr"].ClickedCh:
			_ = c.website.Trigger.Backup(notifiarr.EventUser, starr.Readarr)
		case <-menu["backSonarr"].ClickedCh:
			_ = c.website.Trigger.Backup(notifiarr.EventUser, starr.Sonarr)
		}
	}
}
