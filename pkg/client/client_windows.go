package client

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/hako/durafmt"
	"github.com/kardianos/osext"
	"golift.io/version"
)

// This is the pop-up a user sees when they click update in the menu.
func (c *Client) upgradeWindows(update *update.Update) {
	yes, _ := ui.Question(mnd.Title, "An Update is available! Upgrade Now?\n\n"+
		"Your Version: "+update.Version+"\n"+
		"New Version: "+update.Current+"\n"+
		"Date: "+update.RelDate.Format("Jan 2, 2006")+" ("+
		durafmt.Parse(time.Since(update.RelDate).Round(time.Hour)).String()+" ago)", false)
	if yes {
		if err := c.updateNow(update, "user requested"); err != nil {
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

func (c *Client) AutoWatchUpdate() {
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

	c.Print("Auto-updater enabled. Check interval:", durafmt.Parse(dur).String())

	go func() {
		defer c.CapturePanic()

		time.Sleep(update.SleepTime)
		// Check for update on startup.
		if err := c.checkAndUpdate("startup check"); err != nil {
			c.Errorf("Startup-Update Failed: %v", err)
		}
	}()

	ticker := time.NewTicker(dur)
	for range ticker.C {
		if err := c.checkAndUpdate("automatic"); err != nil {
			c.Errorf("Auto-Update Failed: %v", err)
		}
	}
}

func (c *Client) checkAndUpdate(how string) error {
	c.Debugf("Checking GitHub for Update.")

	u, err := update.Check(mnd.UserRepo, version.Version)
	if err != nil {
		return fmt.Errorf("checking GitHub for update: %w", err)
	} else if !u.Outdate {
		return nil
	} else if err = c.updateNow(u, how); err != nil {
		return err
	}

	return nil
}

func (c *Client) updateNow(u *update.Update, msg string) error {
	c.Printf("[UPDATE] Downloading and installing update! %s => %s: %s", u.Version, u.Current, u.CurrURL)

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
	backupFile, err := update.Now(cmd)
	if err != nil {
		return fmt.Errorf("installing update: %w", err)
	}

	c.Printf("Update installed to %s restarting! Backup: %s", cmd.Path, backupFile)
	// And exit, so we can restart.
	c.sigkil <- &update.Signal{Text: "upgrade request: " + msg}

	return nil
}

func (c *Client) handleAptHook() error {
	return fmt.Errorf("this feature is not supported on this platform") //nolint:goerr113
}

func (c *Client) checkReloadSignal(sigc os.Signal) error {
	return c.reloadConfiguration("Caught Signal: " + sigc.String())
}

func (c *Client) setSignals() {
	signal.Notify(c.sigkil, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c.sighup, syscall.SIGHUP)
}
