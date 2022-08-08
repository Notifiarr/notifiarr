package backups

import (
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
	for _, app := range c.Apps.Lidarr {
		if app.Backup != mnd.Disabled && app.Timeout.Duration >= 0 && app.URL != "" {
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
	for _, app := range c.Apps.Radarr {
		if app.Backup != mnd.Disabled && app.Timeout.Duration >= 0 && app.URL != "" {
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
	for _, app := range c.Apps.Readarr {
		if app.Backup != mnd.Disabled && app.Timeout.Duration >= 0 && app.URL != "" {
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
	for _, app := range c.Apps.Sonarr {
		if app.Backup != mnd.Disabled && app.Timeout.Duration >= 0 && app.URL != "" {
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
	for _, app := range c.Apps.Prowlarr {
		if app.Backup != mnd.Disabled && app.Timeout.Duration >= 0 && app.URL != "" {
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

func (c *cmd) sendLidarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Lidarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Lidarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
				skip:  app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0,
			})
		}
	}
}

func (c *cmd) sendProwlarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Prowlarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Prowlarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
				skip:  app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0,
			})
		}
	}
}

func (c *cmd) sendRadarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Radarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Radarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
				skip:  app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0,
			})
		}
	}
}

func (c *cmd) sendReadarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Readarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Readarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
				skip:  app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0,
			})
		}
	}
}

func (c *cmd) sendSonarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Sonarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Sonarr,
				cName: app.Name,
				int:   idx + 1,
				app:   app,
				skip:  app.URL == "" || app.APIKey == "" || app.Timeout.Duration < 0,
			})
		}
	}
}

func (c *cmd) sendBackups(input *genericInstance) {
	if input.skip {
		return
	}

	fileList, err := input.app.GetBackupFiles()
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
