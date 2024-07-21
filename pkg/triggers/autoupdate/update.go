package autoupdate

/* This file handles manual update requests. */

import (
	"context"
	"runtime"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"golift.io/version"
)

// CheckForUpdate is called by a user clicking in the menu/tray icon.
func (a *Action) CheckForUpdate(ctx context.Context, unstable bool) {
	a.cmd.checkForUpdates(ctx, unstable)
}

func (c *cmd) checkForUpdates(ctx context.Context, unstable bool) {
	var (
		data  *update.Update
		err   error
		where = "GitHub"
	)

	if unstable {
		c.Print("[user requested] Unstable Update Check")

		data, err = update.CheckUnstable(ctx, mnd.Title, version.Revision)
		where = "Unstable website"
	} else {
		c.Print("[user requested] GitHub Update Check")

		data, err = update.CheckGitHub(ctx, mnd.UserRepo, version.Version)
	}

	switch {
	case err != nil:
		c.Errorf("%s Update Check: %v", where, err)
		_, _ = ui.Error("Checking version on " + where + ": " + err.Error())
	case data.Outdate && runtime.GOOS == mnd.Windows:
		c.upgradeWindows(ctx, data)
	case data.Outdate:
		c.downloadUpdate(data, unstable)
	default:
		_, _ = ui.Info("You're up to date! Version: " + data.Current + " @ " + where + "\n" +
			"Updated: " + data.RelDate.Format("Jan 2, 2006") + mnd.DurationAge(data.RelDate))
	}
}

// upgradeWindows is the pop-up a Windows user sees when they click update in the menu.
func (c *cmd) upgradeWindows(ctx context.Context, update *update.Update) {
	yes, _ := ui.Question("An Update is available! Upgrade Now?\n\n"+
		"Your Version: "+version.Version+"-"+version.Revision+"\n"+
		"New Version: "+update.Current+"\n"+
		"Date: "+update.RelDate.Format("Jan 2, 2006")+mnd.DurationAge(update.RelDate), false)
	if yes {
		if err := c.updateNow(ctx, update, "user requested"); err != nil {
			c.Errorf("Application Update Failed: %v", err)
			_, _ = ui.Error("Updating Notifiarr:\n" + err.Error() + "\n")
		}
	}
}

// downloadUpdate is the pop-up a mac user sees when they click update in the menu.
func (c *cmd) downloadUpdate(update *update.Update, unstable bool) {
	msg := "An Update from GitHub is available! Download?\n\n"

	if unstable {
		msg = "An Unstable Update is available! Download?\n\n"
	}

	yes, _ := ui.Question(msg+
		"Your Version: "+version.Version+"-"+version.Revision+"\n"+
		"New Version: "+update.Current+"\n"+
		"Date: "+update.RelDate.Format("Jan 2, 2006")+mnd.DurationAge(update.RelDate), false)
	if yes {
		_ = ui.OpenURL(update.CurrURL)
	}
}
