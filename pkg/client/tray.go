//go:build darwin || windows || linux

//nolint:errcheck
package client

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"github.com/energye/systray"
	"golift.io/starr"
	"golift.io/version"
)

/* This file handles the OS GUI elements. */

// This is arbitrary to avoid conflicts.
const timerPrefix = "TimErr"

// This variable holds all the menu items.
var menu = make(map[string]*systray.MenuItem) //nolint:gochecknoglobals

// startTray Run()s readyTray to bring up the web server and the GUI app.
func (c *Client) startTray(ctx context.Context, cancel context.CancelFunc, clientInfo *clientinfo.ClientInfo) {
	systray.Run(func() {
		defer os.Exit(0)
		defer c.CapturePanic()

		b, _ := bindata.Asset(ui.SystrayIcon)
		systray.SetTemplateIcon(b, b)
		systray.SetTooltip(version.Print(c.Flags.Name()))
		// systray.SetOnClick(c.showMenu) // buggy
		systray.SetOnRClick(c.showMenu)
		systray.SetOnDClick(func(menu systray.IMenu) { c.openGUI() })
		c.makeMenus(ctx)         // make the menu before starting the web server.
		c.setupMenus(clientInfo) // code that runs on reload, too.

		// This starts the web server, and waits for reload/exit signals.
		if err := c.Exit(ctx, cancel); err != nil {
			c.Errorf("Server: %v", err)
			os.Exit(1) // web server problem
		}
	}, func() {
		// This code only fires from menu->quit.
		if err := c.stop(ctx, website.EventUser); err != nil {
			c.Errorf("Server: %v", err)
			os.Exit(1) // web server problem
		}
		// because systray wants to control the exit code? no..
		os.Exit(0)
	})
}

func (c *Client) showMenu(menu systray.IMenu) {
	if err := menu.ShowMenu(); err != nil {
		c.Errorf("Menu Failed: %v", err)
	}
}

// setupMenus is the only code that re-runs on reload.
func (c *Client) setupMenus(clientInfo *clientinfo.ClientInfo) {
	if !ui.HasGUI() {
		return
	}

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

	go c.buildDynamicTimerMenus()

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

func (c *Client) makeMenus(ctx context.Context) {
	menu["stat"] = systray.AddMenuItem("Running", "web server state unknown")
	menu["stat"].Click(func() { c.toggleServer(ctx) })

	c.configMenu(ctx)
	c.linksMenu()
	c.logsMenu()
	c.notifiarrMenu()
	c.debugMenu()

	menu["update"] = systray.AddMenuItem("Update", "check GitHub for updated version")
	menu["update"].Click(func() { go c.checkForUpdate(ctx) })
	menu["gui"] = systray.AddMenuItem("Open WebUI", "open the web page for this Notifiarr client")
	menu["gui"].Click(c.openGUI)
	menu["sub"] = systray.AddMenuItem("Subscribe", "subscribe for premium features")
	menu["sub"].Click(func() { go ui.OpenURL("https://github.com/sponsors/Notifiarr") })
	menu["exit"] = systray.AddMenuItem("Quit", "exit "+c.Flags.Name())
	menu["exit"].Click(func() {
		c.Printf("Need help? %s\n=====> Exiting! User Requested", mnd.HelpLink)
		systray.Quit() // this kills the app
	})
}

func (c *Client) configMenu(ctx context.Context) {
	conf := systray.AddMenuItem("Config", "show configuration")
	menu["conf"] = conf

	menu["view"] = conf.AddSubMenuItem("View", "show configuration")
	menu["view"].Click(func() {
		go ui.Info(mnd.Title+": Configuration", c.displayConfig())
	})

	menu["edit"] = conf.AddSubMenuItem("Edit", "edit configuration")
	menu["edit"].Click(func() {
		go ui.OpenFile(c.Flags.ConfigFile)
		c.Print("user requested] Editing Config File:", c.Flags.ConfigFile)
	})

	menu["pass"] = conf.AddSubMenuItem("Password", "create or update the Web UI admin password")
	menu["pass"].Click(func() { c.updatePassword(ctx) })

	menu["write"] = conf.AddSubMenuItem("Write", "write config file")
	menu["write"].Click(func() {
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		c.writeConfigFile(ctx)
	})

	menu["svcs"] = conf.AddSubMenuItem("Services", "toggle service checks routine")
	menu["svcs"].Click(func() {
		if menu["svcs"].Checked() {
			menu["svcs"].Uncheck()
			c.Config.Services.Stop()
			ui.Notify("Stopped checking services!")
		} else {
			menu["svcs"].Check()
			c.Config.Services.Start(ctx)
			ui.Notify("Service checks started!")
		}
	})

	menu["load"] = conf.AddSubMenuItem("Reload", "reload configuration")
	menu["load"].Click(func() {
		c.triggerConfigReload(website.EventUser, "User Requested")
	})
}

func (c *Client) linksMenu() {
	link := systray.AddMenuItem("Links", "external resources")
	menu["link"] = link

	menu["hp"] = link.AddSubMenuItem("Notifiarr.com", "open Notifiarr.com")
	menu["hp"].Click(func() { go ui.OpenURL("https://notifiarr.com/") })

	menu["wiki"] = link.AddSubMenuItem("Notifiarr.Wiki", "open Notifiarr wiki")
	menu["wiki"].Click(func() { go ui.OpenURL("https://notifiarr.wiki/") })

	menu["trash"] = link.AddSubMenuItem("TRaSH Guide", "open TRaSH wiki for Notifiarr")
	menu["trash"].Click(func() { go ui.OpenURL("https://trash-guides.info/Notifiarr/Quick-Start/") })

	menu["disc1"] = link.AddSubMenuItem("Notifiarr Discord", "open Notifiarr discord server")
	menu["disc1"].Click(func() { go ui.OpenURL("https://notifiarr.com/discord") })

	menu["disc2"] = link.AddSubMenuItem("Go Lift Discord", "open Go Lift discord server")
	menu["disc2"].Click(func() { go ui.OpenURL("https://golift.io/discord") })

	menu["gh"] = link.AddSubMenuItem("GitHub Project", c.Flags.Name()+" on GitHub")
	menu["gh"].Click(func() { go ui.OpenURL("https://github.com/Notifiarr/notifiarr/") })
}

func (c *Client) logsMenu() {
	logs := systray.AddMenuItem("Logs", "log file info")
	menu["logs"] = logs

	menu["logs_view"] = logs.AddSubMenuItem("View", "view the application log")
	menu["logs_view"].Click(func() {
		go ui.OpenLog(c.Config.LogFile)
		c.Print("[user requested] Viewing App Log File:", c.Config.LogFile)
	})

	menu["logs_http"] = logs.AddSubMenuItem("HTTP", "view the HTTP log")
	menu["logs_http"].Click(func() {
		go ui.OpenLog(c.Config.HTTPLog)
		c.Print("[user requested] Viewing HTTP Log File:", c.Config.HTTPLog)
	})

	menu["debug_logs2"] = logs.AddSubMenuItem("Debug", "view the Debug log")
	menu["debug_logs2"].Click(func() {
		go ui.OpenLog(c.Config.LogConfig.DebugLog)
		c.Print("[user requested] Viewing Debug File:", c.Config.LogConfig.DebugLog)
	})

	menu["logs_svcs"] = logs.AddSubMenuItem("Services", "view the Services log")
	menu["logs_svcs"].Click(func() {
		go ui.OpenLog(c.Config.Services.LogFile)
		c.Print("[user requested] Viewing Services Log File:", c.Config.Services.LogFile)
	})

	menu["logs_rotate"] = logs.AddSubMenuItem("Rotate", "rotate both log files")
	menu["logs_rotate"].Click(c.rotateLogs)
}

// notifiarrMenu makes the Notifiarr menu.
//
//nolint:lll
func (c *Client) notifiarrMenu() {
	data := systray.AddMenuItem("Notifiarr", "plex sessions, system snapshots, service checks")
	menu["data"] = data
	menu["gaps"] = data.AddSubMenuItem("Send Radarr Gaps", "[premium feature] trigger radarr collections gaps")
	menu["synccf"] = data.AddSubMenuItem("TRaSH: Sync Radarr", "[premium feature] trigger TRaSH radarr sync")
	menu["syncqp"] = data.AddSubMenuItem("TRaSH: Sync Sonarr", "[premium feature] trigger TRaSH sonarr sync")
	menu["svcs_prod"] = data.AddSubMenuItem("Check and Send Services", "check all services and send results to notifiarr")
	menu["plex_prod"] = data.AddSubMenuItem("Send Plex Sessions", "send plex sessions to notifiarr")
	menu["snap_prod"] = data.AddSubMenuItem("Send System Snapshot", "send system snapshot to notifiarr")
	menu["send_dash"] = data.AddSubMenuItem("Send Dashboard States", "collect and send all application states for a dashboard update")
	menu["corrLidarr"] = data.AddSubMenuItem("Check Lidarr Corruption", "check latest backup database in each instance for corruption")
	menu["corrProwlarr"] = data.AddSubMenuItem("Check Prowlarr Corruption", "check latest backup database in each instance for corruption")
	menu["corrRadarr"] = data.AddSubMenuItem("Check Radarr Corruption", "check latest backup database in each instance for corruption")
	menu["corrReadarr"] = data.AddSubMenuItem("Check Readarr Corruption", "check latest backup database in each instance for corruption")
	menu["corrSonarr"] = data.AddSubMenuItem("Check Sonarr Corruption", "check latest backup database in each instance for corruption")
	menu["backLidarr"] = data.AddSubMenuItem("Send Lidarr Backups", "send backup file list for each instance to Notifiarr")
	menu["backProwlarr"] = data.AddSubMenuItem("Send Prowlarr Backups", "send backup file list for each instance to Notifiarr")
	menu["backRadarr"] = data.AddSubMenuItem("Send Radarr Backups", "send backup file list for each instance to Notifiarr")
	menu["backReadarr"] = data.AddSubMenuItem("Send Readarr Backups", "send backup file list for each instance to Notifiarr")
	menu["backSonarr"] = data.AddSubMenuItem("Send Sonarr Backups", "send backup file list for each instance to Notifiarr")

	c.notifiarrMenuActions()
}

func (c *Client) notifiarrMenuActions() {
	menu["gaps"].Click(func() { c.triggers.Gaps.Send(website.EventUser) })
	menu["synccf"].Click(func() { c.triggers.CFSync.SyncRadarrCF(website.EventUser) })
	menu["syncqp"].Click(func() { c.triggers.CFSync.SyncSonarrRP(website.EventUser) })
	menu["svcs_prod"].Click(func() {
		c.Print("[user requested] Checking services and sending results to Notifiarr.")
		ui.Notify("Running and sending %d Service Checks.", len(c.Config.Service))
		c.Config.Services.RunChecks(website.EventUser)
	})
	menu["plex_prod"].Click(func() { c.triggers.PlexCron.Send(website.EventUser) })
	menu["snap_prod"].Click(func() { c.triggers.SnapCron.Send(website.EventUser) })
	menu["send_dash"].Click(func() { c.triggers.Dashboard.Send(website.EventUser) })
	menu["corrLidarr"].Click(func() {
		_ = c.triggers.Backups.Corruption(&common.ActionInput{Type: website.EventUser}, starr.Lidarr)
	})
	menu["corrProwlarr"].Click(func() {
		_ = c.triggers.Backups.Corruption(&common.ActionInput{Type: website.EventUser}, starr.Prowlarr)
	})
	menu["corrRadarr"].Click(func() {
		_ = c.triggers.Backups.Corruption(&common.ActionInput{Type: website.EventUser}, starr.Radarr)
	})
	menu["corrReadarr"].Click(func() {
		_ = c.triggers.Backups.Corruption(&common.ActionInput{Type: website.EventUser}, starr.Readarr)
	})
	menu["corrSonarr"].Click(func() {
		_ = c.triggers.Backups.Corruption(&common.ActionInput{Type: website.EventUser}, starr.Sonarr)
	})
	menu["backLidarr"].Click(func() {
		_ = c.triggers.Backups.Backup(&common.ActionInput{Type: website.EventUser}, starr.Lidarr)
	})
	menu["backProwlarr"].Click(func() {
		_ = c.triggers.Backups.Backup(&common.ActionInput{Type: website.EventUser}, starr.Prowlarr)
	})
	menu["backRadarr"].Click(func() {
		_ = c.triggers.Backups.Backup(&common.ActionInput{Type: website.EventUser}, starr.Radarr)
	})
	menu["backReadarr"].Click(func() {
		_ = c.triggers.Backups.Backup(&common.ActionInput{Type: website.EventUser}, starr.Readarr)
	})
	menu["backSonarr"].Click(func() {
		_ = c.triggers.Backups.Backup(&common.ActionInput{Type: website.EventUser}, starr.Sonarr)
	})
}

func (c *Client) debugMenu() {
	debug := systray.AddMenuItem("Debug", "Debug Menu")
	menu["debug"] = debug

	menu["debug_logs"] = debug.AddSubMenuItem("View Debug Log", "view the Debug log")
	menu["debug_logs"].Click(func() {
		go ui.OpenLog(c.Config.LogConfig.DebugLog)
		c.Print("[user requested] Viewing Debug File:", c.Config.LogConfig.DebugLog)
	})

	menu["svcs_log"] = debug.AddSubMenuItem("Log Service Checks", "check all services and log results")
	menu["svcs_log"].Click(func() {
		c.Print("[user requested] Checking services and logging results.")
		ui.Notify("Running and logging %d Service Checks.", len(c.Config.Service))
		c.Config.Services.RunChecks("log")
	})

	menu["console"] = debug.AddSubMenuItem("Console", "toggle the console window")
	menu["console"].Click(func() {
		if menu["console"].Checked() {
			menu["console"].Uncheck()
			ui.HideConsoleWindow()
		} else {
			menu["console"].Check()
			ui.ShowConsoleWindow()
		}
	})

	if runtime.GOOS != mnd.Windows {
		menu["console"].Hide()
	}

	debug.AddSubMenuItem("- Danger Zone -", "").Disable()
	menu["debug_panic"] = debug.AddSubMenuItem("Application Panic", "cause an application panic (crash)")
	menu["debug_panic"].Click(c.menuPanic)
}

func (c *Client) buildDynamicTimerMenus() {
	defer c.CapturePanic()

	c.closeDynamicTimerMenus()

	timers := c.triggers.CronTimer.List()
	if len(timers) == 0 {
		return
	}

	if menu["timerinfo"] == nil {
		menu["timerinfo"] = menu["data"].AddSubMenuItem("- Custom Timers -", "")
	}

	menu["timerinfo"].Show()
	menu["timerinfo"].Disable()

	for idx, timer := range timers {
		idx := idx

		desc := fmt.Sprintf("%s; config: interval: %s, path: %s", timer.Desc, timer.Interval, timer.URI)
		if timer.Desc == "" {
			desc = fmt.Sprintf("dynamic custom timer; config: interval: %s, path: %s", timer.Interval, timer.URI)
		}

		name := timerPrefix + timer.Name
		if menu[name] == nil {
			menu[name] = menu["data"].AddSubMenuItem(timer.Name, desc)
		}

		menu[name].SetTooltip(desc)
		menu[name].Show()
		menu[name].Click(func() { timers[idx].Run(&common.ActionInput{Type: website.EventUser}) })
	}
}

func (c *Client) closeDynamicTimerMenus() {
	for name := range menu {
		if menu[name] != nil && strings.HasPrefix(name, timerPrefix) {
			menu[name].Hide()
		}
	}

	if menu["timerinfo"] != nil {
		// We get here on reload, and all previous timers are gone now.
		menu["timerinfo"].Hide()
	}
}
