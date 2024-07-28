package autoupdate

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/hako/durafmt"
	"golift.io/cnfg"
	"golift.io/version"
)

// TrigAutoUpdate is our auto update trigger identifier.
const TrigAutoUpdate common.TriggerName = "Automatically check for application update."

const (
	minimumUpdateDur = 1 * time.Hour
	// waitTime is how long we wait, after startup, before doing an update check.
	waitTime = 6 * time.Minute
)

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	AutoUpdate string
	UnstableCh bool
	ConfigFile string
	sync.Mutex
	stop bool
}

// New configures the library.
func New(config *common.Config, autoUpdate, configFile string, unstable bool) *Action {
	return &Action{cmd: &cmd{
		Config:     config,
		AutoUpdate: autoUpdate,
		UnstableCh: unstable,
		ConfigFile: configFile,
	}}
}

// Run checks for an update a few minutes after the app starts up.
func (a *Action) Run(ctx context.Context) {
	if !mnd.IsWindows {
		return // Auto update only works on Windows.
	}

	time.Sleep(waitTime)
	a.cmd.Lock()
	defer a.cmd.Unlock()

	if !a.cmd.stop { // Check for update on startup, after a short wait.
		a.cmd.checkAndUpdate(ctx, &common.ActionInput{Type: website.EventStart})
	}
}

// Create initializes the library.
func (a *Action) Create() {
	if !mnd.IsWindows {
		return // Auto update only works on Windows.
	}

	a.cmd.create()
}

// Stop satisfies an interface.
func (a *Action) Stop() {
	a.cmd.Lock()
	defer a.cmd.Unlock()
	a.cmd.stop = true
}

// Verify the interfaces are satisfied.
var (
	_ = common.Run(&Action{nil})
	_ = common.Create(&Action{nil})
)

// Run fires in a go routine. Wait a minute or two then tell the website we're up.
// If app reloads in first checkWait duration, this throws an error. That's ok.
func (c *cmd) create() {
	defer c.CapturePanic()

	var dur time.Duration

	switch c.AutoUpdate {
	case "off", "no", "disabled", "disable", "false", "", "-", "0", "0s":
		return
	case "hourly":
		dur = time.Hour
	case "daily":
		dur = mnd.OneDay
	default:
		var err error
		if dur, err = time.ParseDuration(c.AutoUpdate); err != nil {
			dur = mnd.OneDay
			break
		}

		if dur < minimumUpdateDur {
			dur = minimumUpdateDur
		}
	}

	pfx := "GitHub"
	if c.UnstableCh {
		pfx = "Unstable"
	}

	c.Printf("==> Client auto-updater started. %s channel check interval: %s",
		pfx, durafmt.Parse(dur).LimitFirstN(3)) //nolint:mnd

	c.Add(&common.Action{
		Name: TrigAutoUpdate,
		Fn:   c.checkAndUpdate,
		D:    cnfg.Duration{Duration: dur},
	})
}

func (c *cmd) checkAndUpdate(ctx context.Context, action *common.ActionInput) {
	var (
		data *update.Update
		err  error
	)

	if !mnd.IsWindows {
		return // Only Windows can auto update.
	}

	//nolint:wsl
	if c.UnstableCh {
		c.Debugf("[cron requested] Checking Unstable website for Update.")
		data, err = update.CheckUnstable(ctx, mnd.DefaultName, version.Revision)
	} else {
		c.Debugf("[cron requested] Checking GitHub for Update.")
		data, err = update.CheckGitHub(ctx, mnd.UserRepo, version.Version)
	}

	if err != nil {
		c.Errorf("Auto-Update Failed checking for update: %v", err)
	} else if !data.Outdate {
		c.Debugf("Auto-Update Success, up to date.")
	} else if err = c.updateNow(ctx, data, action.Type); err != nil {
		c.Errorf("Auto-Update Failed applying update: %v", err)
	}
}

func (c *cmd) updateNow(ctx context.Context, u *update.Update, msg website.EventType) error {
	c.Printf("[UPDATE] Downloading and installing update! %s-%s => %s: %s",
		version.Version, version.Revision, u.Current, u.CurrURL)

	cmd := &update.Command{
		URL:    u.CurrURL,
		Logger: c.Logger,
		Args:   []string{"--restart", "--config", c.ConfigFile},
		Path:   os.Args[0],
	}

	if path, err := os.Executable(); err == nil {
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
	c.StopApp("upgrade request: " + string(msg))

	return nil
}
