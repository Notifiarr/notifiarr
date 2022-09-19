package backups

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr"
)

// Backup initializes a backup check for all instances of the provided app.
func (a *Action) Backup(event website.EventType, app starr.App) error {
	switch app {
	default:
		return fmt.Errorf("%w: %s", common.ErrInvalidApp, app)
	case "":
		return fmt.Errorf("%w: <no app provided>", common.ErrInvalidApp)
	case "All":
		a.cmd.Exec(event, TrigLidarrBackup)
		a.cmd.Exec(event, TrigProwlarrBackup)
		a.cmd.Exec(event, TrigRadarrBackup)
		a.cmd.Exec(event, TrigReadarrBackup)
		a.cmd.Exec(event, TrigSonarrBackup)
	case starr.Lidarr:
		a.cmd.Exec(event, TrigLidarrBackup)
	case starr.Prowlarr:
		a.cmd.Exec(event, TrigProwlarrBackup)
	case starr.Radarr:
		a.cmd.Exec(event, TrigRadarrBackup)
	case starr.Readarr:
		a.cmd.Exec(event, TrigReadarrBackup)
	case starr.Sonarr:
		a.cmd.Exec(event, TrigSonarrBackup)
	}

	return nil
}

func (c *cmd) makeBackupTriggersLidarr() {
	var ticker *time.Ticker

	//nolint:gosec
	for idx, app := range c.Apps.Lidarr {
		if ci := website.GetClientInfo(); ci != nil &&
			ci.Actions.Apps.Lidarr.Backup(idx+1) != mnd.Disabled && app.Enabled() {
			randomTime := time.Duration(rand.Intn(randomMinutes))*time.Second +
				time.Duration(rand.Intn(randomMinutes))*time.Minute
			ticker = time.NewTicker(checkInterval + randomTime)

			break
		}
	}

	c.Add(&common.Action{
		Name: TrigLidarrBackup,
		Fn:   c.sendLidarrBackups,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) makeBackupTriggersRadarr() {
	var ticker *time.Ticker

	//nolint:gosec
	for idx, app := range c.Apps.Radarr {
		if ci := website.GetClientInfo(); ci != nil &&
			ci.Actions.Apps.Radarr.Backup(idx+1) != mnd.Disabled && app.Enabled() {
			randomTime := time.Duration(rand.Intn(randomMinutes))*time.Second +
				time.Duration(rand.Intn(randomMinutes))*time.Minute
			ticker = time.NewTicker(checkInterval + randomTime)

			break
		}
	}

	c.Add(&common.Action{
		Name: TrigRadarrBackup,
		Fn:   c.sendRadarrBackups,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) makeBackupTriggersReadarr() {
	var ticker *time.Ticker

	//nolint:gosec
	for idx, app := range c.Apps.Readarr {
		if ci := website.GetClientInfo(); ci != nil &&
			ci.Actions.Apps.Readarr.Backup(idx+1) != mnd.Disabled && app.Enabled() {
			randomTime := time.Duration(rand.Intn(randomMinutes))*time.Second +
				time.Duration(rand.Intn(randomMinutes))*time.Minute
			ticker = time.NewTicker(checkInterval + randomTime)

			break
		}
	}

	c.Add(&common.Action{
		Name: TrigReadarrBackup,
		Fn:   c.sendReadarrBackups,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) makeBackupTriggersSonarr() {
	var ticker *time.Ticker

	//nolint:gosec
	for idx, app := range c.Apps.Sonarr {
		if ci := website.GetClientInfo(); ci != nil &&
			ci.Actions.Apps.Sonarr.Backup(idx+1) != mnd.Disabled && app.Enabled() {
			randomTime := time.Duration(rand.Intn(randomMinutes))*time.Second +
				time.Duration(rand.Intn(randomMinutes))*time.Minute
			ticker = time.NewTicker(checkInterval + randomTime)

			break
		}
	}

	c.Add(&common.Action{
		Name: TrigSonarrBackup,
		Fn:   c.sendSonarrBackups,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) makeBackupTriggersProwlarr() {
	var ticker *time.Ticker

	//nolint:gosec
	for idx, app := range c.Apps.Prowlarr {
		if ci := website.GetClientInfo(); ci != nil &&
			ci.Actions.Apps.Prowlarr.Backup(idx+1) != mnd.Disabled && app.Enabled() {
			randomTime := time.Duration(rand.Intn(randomMinutes))*time.Second +
				time.Duration(rand.Intn(randomMinutes))*time.Minute
			ticker = time.NewTicker(checkInterval + randomTime)

			break
		}
	}

	c.Add(&common.Action{
		Name: TrigProwlarrBackup,
		Fn:   c.sendProwlarrBackups,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) sendLidarrBackups(ctx context.Context, event website.EventType) {
	for idx, app := range c.Apps.Lidarr {
		if ci := website.GetClientInfo(); event != website.EventCron ||
			(ci != nil && ci.Actions.Apps.Lidarr.Backup(idx+1) != mnd.Disabled) {
			c.sendBackups(ctx, &genericInstance{
				event: event,
				name:  starr.Lidarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
				skip:  !app.Enabled(),
			})
		}
	}
}

func (c *cmd) sendProwlarrBackups(ctx context.Context, event website.EventType) {
	for idx, app := range c.Apps.Prowlarr {
		if ci := website.GetClientInfo(); event != website.EventCron ||
			(ci != nil && ci.Actions.Apps.Prowlarr.Backup(idx+1) != mnd.Disabled) {
			c.sendBackups(ctx, &genericInstance{
				event: event,
				name:  starr.Prowlarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
				skip:  !app.Enabled(),
			})
		}
	}
}

func (c *cmd) sendRadarrBackups(ctx context.Context, event website.EventType) {
	for idx, app := range c.Apps.Radarr {
		if ci := website.GetClientInfo(); event != website.EventCron ||
			(ci != nil && ci.Actions.Apps.Radarr.Backup(idx+1) != mnd.Disabled) {
			c.sendBackups(ctx, &genericInstance{
				event: event,
				name:  starr.Radarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
				skip:  !app.Enabled(),
			})
		}
	}
}

func (c *cmd) sendReadarrBackups(ctx context.Context, event website.EventType) {
	for idx, app := range c.Apps.Readarr {
		if ci := website.GetClientInfo(); event != website.EventCron ||
			(ci != nil && ci.Actions.Apps.Readarr.Backup(idx+1) != mnd.Disabled) {
			c.sendBackups(ctx, &genericInstance{
				event: event,
				name:  starr.Readarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
				skip:  !app.Enabled(),
			})
		}
	}
}

func (c *cmd) sendSonarrBackups(ctx context.Context, event website.EventType) {
	for idx, app := range c.Apps.Sonarr {
		if ci := website.GetClientInfo(); event != website.EventCron ||
			(ci != nil && ci.Actions.Apps.Sonarr.Backup(idx+1) != mnd.Disabled) {
			c.sendBackups(ctx, &genericInstance{
				event: event,
				name:  starr.Sonarr,
				cName: app.Name,
				int:   idx + 1,
				app:   app,
				skip:  !app.Enabled(),
			})
		}
	}
}

func (c *cmd) sendBackups(ctx context.Context, input *genericInstance) {
	if input.skip {
		return
	}

	fileList, err := input.app.GetBackupFilesContext(ctx)
	if err != nil {
		c.Errorf("[%s requested] Getting %s Backup Files (%d): %v", input.event, input.name, input.int, err)
		return
	} else if len(fileList) == 0 {
		c.Printf("[%s requested] %s has no backup files (%d)", input.event, input.name, input.int)
		return
	}

	send := &Payload{
		App:   input.name,
		Int:   input.int,
		Name:  input.cName,
		Files: fileList,
	}

	c.SendData(&website.Request{
		Route:      website.BackupRoute,
		Event:      input.event,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("%s Backup File List (%d)", input.name, input.int),
		Payload:    send,
	})
}
