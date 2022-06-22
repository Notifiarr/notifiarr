package backups

import (
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr"
)

func (c *Config) Backup(event website.EventType, app starr.App) error {
	switch app {
	default:
		return fmt.Errorf("%w: %s", common.ErrInvalidApp, app)
	case "":
		return fmt.Errorf("%w: <no app provided>", common.ErrInvalidApp)
	case "All":
		c.Exec(event, TrigLidarrBackup)
		c.Exec(event, TrigProwlarrBackup)
		c.Exec(event, TrigRadarrBackup)
		c.Exec(event, TrigReadarrBackup)
		c.Exec(event, TrigSonarrBackup)
	case starr.Lidarr:
		c.Exec(event, TrigLidarrBackup)
	case starr.Prowlarr:
		c.Exec(event, TrigProwlarrBackup)
	case starr.Radarr:
		c.Exec(event, TrigRadarrBackup)
	case starr.Readarr:
		c.Exec(event, TrigReadarrBackup)
	case starr.Sonarr:
		c.Exec(event, TrigSonarrBackup)
	}

	return nil
}

func (c *Config) makeBackupTriggers() {
	c.Add(&common.Action{
		Name: TrigLidarrBackup,
		Fn:   c.sendLidarrBackups,
		C:    make(chan website.EventType, 1),
		T:    time.NewTicker(lidarrBackupCheckDur),
	}, &common.Action{
		Name: TrigProwlarrBackup,
		Fn:   c.sendProwlarrBackups,
		C:    make(chan website.EventType, 1),
		T:    time.NewTicker(prowlarrBackupCheckDur),
	}, &common.Action{
		Name: TrigRadarrBackup,
		Fn:   c.sendRadarrBackups,
		C:    make(chan website.EventType, 1),
		T:    time.NewTicker(radarrBackupCheckDur),
	}, &common.Action{
		Name: TrigReadarrBackup,
		Fn:   c.sendReadarrBackups,
		C:    make(chan website.EventType, 1),
		T:    time.NewTicker(readarrBackupCheckDur),
	}, &common.Action{
		Name: TrigSonarrBackup,
		Fn:   c.sendSonarrBackups,
		C:    make(chan website.EventType, 1),
		T:    time.NewTicker(sonarrBackupCheckDur),
	})
}

func (c *Config) sendLidarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Lidarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Lidarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
			})
		}
	}
}

func (c *Config) sendProwlarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Prowlarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Prowlarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
			})
		}
	}
}

func (c *Config) sendRadarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Radarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Radarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
			})
		}
	}
}

func (c *Config) sendReadarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Readarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Readarr,
				int:   idx + 1,
				app:   app,
				cName: app.Name,
			})
		}
	}
}

func (c *Config) sendSonarrBackups(event website.EventType) {
	for idx, app := range c.Apps.Sonarr {
		if app.Backup != mnd.Disabled || event != website.EventCron {
			c.sendBackups(&genericInstance{
				event: event,
				name:  starr.Sonarr,
				cName: app.Name,
				int:   idx + 1,
				app:   app,
			})
		}
	}
}

func (c *Config) sendBackups(input *genericInstance) {
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

	c.QueueData(&website.SendRequest{
		Route:      website.BackupRoute,
		Event:      input.event,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("%s Backup File List (%d)", input.name, input.int),
		Payload:    send,
	})
}
