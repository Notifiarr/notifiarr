package client

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/bindata"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/hako/durafmt"
	"github.com/kardianos/osext"
	"golift.io/version"
	"gopkg.in/toast.v1"
)

func (c *Client) checkReloadSignal(sigc os.Signal) {
	c.reloadConfiguration("caught signal: " + sigc.String())
}

func (c *Client) setReloadSignals() {
	signal.Notify(c.sighup, syscall.SIGHUP)
}

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

// getPNG purposely returns an empty string when there is no verified file.
// This is used to give the toast notification an icon.
// Do not throw errors if the icon is missing, it'd nbd, just return empty "".
func (c *Client) getPNG() string {
	folder, err := osext.ExecutableFolder()
	if err != nil {
		c.Debug("Error Finding app folder:", err) // purposely debug and not error
		return ""
	}

	const minimumFileSize = 100 // arbitrary

	pngPath := filepath.Join(folder, "notifiarr.png")
	if f, err := os.Stat(pngPath); err == nil && f.Size() > minimumFileSize {
		return pngPath // most code paths land here.
	} else if !os.IsNotExist(err) || (f != nil && f.Size() < minimumFileSize) {
		c.Debug("Error Stating file:", err) // purposely debug and not error
		return ""
	}

	data, err := bindata.Asset("files/favicon.png")
	if err != nil {
		c.Debug("Error Finding asset:", err) // purposely debug and not error
		return ""
	}

	if err := os.WriteFile(pngPath, data, mnd.Mode0600); err != nil {
		c.Debug("Error Writing file:", err) // purposely debug and not error
		return ""
	}

	return pngPath
}

func (c *Client) printUpdateMessage() {
	err := (&toast.Notification{
		AppID:   mnd.Title,
		Title:   mnd.Title + " Upgraded!",
		Message: mnd.Title + " updated to version " + version.Version,
		Icon:    c.getPNG(),
	}).Push()
	if err != nil {
		c.Error("Creating Toast Notification:", err)
	}
}

func (c *Client) AutoWatchUpdate() {
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
		time.Sleep(update.SleepTime)
		// Check for update on startup.
		if err := c.checkAndUpdate("startup check"); err != nil {
			c.Errorf("Auto-Update Failed: %v", err)
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

	uc := &update.Command{
		URL:    u.CurrURL,
		Logger: c.Logger.DebugLog,
		Args:   []string{"--restart", "--config", c.Flags.ConfigFile},
		Path:   os.Args[0],
	}

	if path, err := osext.Executable(); err == nil {
		uc.Path = path
	}

	// This downloads the new file to a temp name in the same folder as the running file.
	// Moves the running file to a backup name in the same folder.
	// Moves the new file to the same location that the running file was at.
	// Triggers another invocation of the app that sleeps 5 seconds then restarts.
	backupFile, err := update.Now(uc)
	if err != nil {
		return fmt.Errorf("installing update: %w", err)
	}

	c.Printf("Update installed to %s restarting! Backup: %s", uc.Path, backupFile)
	// And exit, so we can restart.
	c.sigkil <- &update.Signal{Text: "upgrade request: " + msg}

	return nil
}
