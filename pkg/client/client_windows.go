package client

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/hako/durafmt"
	"github.com/kardianos/osext"
	"golift.io/version"
)

// WaitTime is how long we wait, after startup, before doing an update check.
const WaitTime = 10 * time.Minute

// This is the pop-up a user sees when they click update in the menu.
func (c *Client) upgradeWindows(ctx context.Context, update *update.Update) {
	yes, _ := ui.Question(mnd.Title, "An Update is available! Upgrade Now?\n\n"+
		"Your Version: "+version.Version+"-"+version.Revision+"\n"+
		"New Version: "+update.Current+"\n"+
		"Date: "+update.RelDate.Format("Jan 2, 2006")+mnd.DurationAgo(update.RelDate), false)
	if yes {
		if err := c.updateNow(ctx, update, "user requested"); err != nil {
			c.Errorf("Update Failed: %v", err)
			_, _ = ui.Error(mnd.Title+" ERROR", "Updating Notifiarr:\n"+err.Error()+"\n")
		}
	}
}

func (c *Client) printUpdateMessage() {
	if !c.Flags.Updated {
		return
	}

	err := ui.Notify(mnd.Title + " updated to version " + version.Version)
	if err != nil {
		c.Error("Creating Toast Notification:", err)
	}
}

func (c *Client) AutoWatchUpdate(ctx context.Context) {
	defer c.CapturePanic()

	var dur time.Duration

	switch c.Config.AutoUpdate {
	case "off", "no", "disabled", "disable", "false", "", "-", "0", "0s":
		return
	case "hourly":
		dur = time.Hour
	case "daily":
		dur = mnd.OneDay
	default:
		var err error
		if dur, err = time.ParseDuration(c.Config.AutoUpdate); err != nil {
			dur = mnd.OneDay
			break
		}

		if dur < time.Hour {
			dur = time.Hour
		}
	}

	c.startAutoUpdater(ctx, dur)
}

func (c *Client) startAutoUpdater(ctx context.Context, dur time.Duration) {
	pfx := ""
	if c.Config.UnstableCh {
		pfx = "Unstable Channel "
	}

	time.Sleep(WaitTime)
	c.Print(pfx+"Auto-updater started. Check interval:",
		durafmt.Parse(dur).LimitFirstN(3).Format(mnd.DurafmtUnits)) //nolint:mnd

	// Check for update on startup.
	if err := c.checkAndUpdate(ctx, "startup check"); err != nil {
		c.Errorf("Startup-Update Failed: %v", err)
	}

	ticker := time.NewTicker(dur)
	for range ticker.C { // the ticker never exits.
		if err := c.checkAndUpdate(ctx, "automatic"); err != nil {
			c.Errorf("Auto-Update Failed: %v", err)
		}
	}
}

func (c *Client) checkAndUpdate(ctx context.Context, how string) error {
	var (
		data *update.Update
		err  error
	)

	//nolint:wsl
	if c.Config.UnstableCh {
		c.Debugf("[cron requested] Checking Unstable website for Update.")
		data, err = update.CheckUnstable(ctx, mnd.DefaultName, version.Revision)
	} else {
		c.Debugf("[cron requested] Checking GitHub for Update.")
		data, err = update.CheckGitHub(ctx, mnd.UserRepo, version.Version)
	}

	if err != nil {
		return fmt.Errorf("checking for update: %w", err)
	} else if !data.Outdate {
		return nil
	} else if err = c.updateNow(ctx, data, how); err != nil {
		return err
	}

	return nil
}

func (c *Client) updateNow(ctx context.Context, u *update.Update, msg string) error {
	c.Printf("[UPDATE] Downloading and installing update! %s-%s => %s: %s",
		version.Version, version.Revision, u.Current, u.CurrURL)

	cmd := &update.Command{
		URL:    u.CurrURL,
		Logger: c.Logger.DebugLog,
		Args:   []string{"--restart", "--config", c.Flags.ConfigFile},
		Path:   os.Args[0],
	}

	if path, err := osext.Executable(); err == nil {
		cmd.Path = path
	}

	// This downloads the new file to a temp name in the same folder as the running file.
	// Moves the running file to a backup name in the same folder.
	// Moves the new file to the same location that the running file was at.
	// Triggers another invocation of the app that sleeps 5 seconds then restarts.
	backupFile, err := update.Now(ctx, cmd)
	if err != nil {
		return fmt.Errorf("installing update: %w", err)
	}

	c.Printf("Update installed to %s restarting! Backup: %s", cmd.Path, backupFile)
	// And exit, so we can restart.
	c.sigkil <- &update.Signal{Text: "upgrade request: " + msg}

	return nil
}

func (c *Client) handleAptHook(_ interface{}) error {
	return fmt.Errorf("this feature is not supported on this platform") //nolint:goerr113
}

func (c *Client) checkReloadSignal(ctx context.Context, sigc os.Signal) error {
	return c.reloadConfiguration(ctx, website.EventSignal, "Caught Signal: "+sigc.String())
}

func (c *Client) setSignals() {
	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c.sighup, syscall.SIGHUP)
}
