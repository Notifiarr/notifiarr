package client

import (
	"fmt"
	"os"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/hako/durafmt"
	"github.com/kardianos/osext"
	"golift.io/version"
)

const (
	oneDay   = 24 * time.Hour
	userRepo = "Notifiarr/notifiarr"
)

func (c *Client) AutoWatchUpdate() {
	var dur time.Duration

	switch c.Config.AutoUpdate {
	case "hourly":
		dur = time.Hour
	case "daily":
		dur = oneDay
	default:
		var err error
		if dur, err = time.ParseDuration(c.Config.AutoUpdate); err != nil {
			dur = oneDay
			break
		}

		if dur < time.Minute {
			dur = time.Minute
		}
	}

	c.Print("Starting Auto-Update Thread. Update check interval:",
		durafmt.Parse(dur).String(), false)

	ticker := time.NewTimer(dur)
	for range ticker.C {
		c.Debugf("Checking GitHub for Update.")

		u, err := update.Check(userRepo, version.Version)
		if err != nil {
			c.Errorf("Checking GitHub for Update: %v", err)
		} else if !u.Outdate {
			continue
		} else if err = c.updateNow(u, "automatic"); err != nil {
			c.Errorf("Update Failed: %v", err)
			continue
		}

		return
	}
}

func (c *Client) updateNow(u *update.Update, msg string) error {
	c.Printf("[%s] Downloading and installing update! %s => %s: %s", msg, u.Version, u.Current, u.CurrURL)

	uc := &update.Command{
		URL:    u.CurrURL,
		Logger: c.Logger.DebugLog,
		Args:   []string{"-rc", c.Flags.ConfigFile},
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
	c.sigkil <- &update.Signal{Text: "upgrade request"}

	return nil
}
