package notifiarr

import (
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr"
)

// BackupPayload is the data we send to notifiarr.
type BackupPayload struct {
	App   starr.App           `json:"app"`
	Int   int                 `json:"instance"`
	Name  string              `json:"name"`
	Files []*starr.BackupFile `json:"backups"`
}

// Used to find and identiy each trigger in the trigger list.
const (
	TrigLidarrBackup   TriggerName = "Sending Lidarr Backup File List to Notifiarr."
	TrigProwlarrBackup TriggerName = "Sending Prowlarr Backup File List to Notifiarr."
	TrigRadarrBackup   TriggerName = "Sending Radarr Backup File List to Notifiarr."
	TrigReadarrBackup  TriggerName = "Sending Readarr Backup File List to Notifiarr."
	TrigSonarrBackup   TriggerName = "Sending Sonarr Backup File List to Notifiarr."
)

func (c *Config) makeBackupTriggers() {
	// Built-in backup check timers. Non-adjustable.
	const (
		lidarrBackupCheckDur   = 6*time.Hour + 10*time.Minute
		prowlarrBackupCheckDur = 6*time.Hour + 20*time.Minute
		radarrBackupCheckDur   = 6*time.Hour + 30*time.Minute
		readarrBackupCheckDur  = 6*time.Hour + 40*time.Minute
		sonarrBackupCheckDur   = 6*time.Hour + 50*time.Minute
	)

	c.Trigger.add(&action{
		Name: TrigLidarrBackup,
		Fn:   c.sendLidarrBackups,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(lidarrBackupCheckDur),
	}, &action{
		Name: TrigProwlarrBackup,
		Fn:   c.sendProwlarrBackups,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(prowlarrBackupCheckDur),
	}, &action{
		Name: TrigRadarrBackup,
		Fn:   c.sendRadarrBackups,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(radarrBackupCheckDur),
	}, &action{
		Name: TrigReadarrBackup,
		Fn:   c.sendReadarrBackups,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(readarrBackupCheckDur),
	}, &action{
		Name: TrigSonarrBackup,
		Fn:   c.sendSonarrBackups,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(sonarrBackupCheckDur),
	})
}

func (t *Triggers) Backup(event EventType, app starr.App) error {
	switch app {
	default:
		return fmt.Errorf("%w: %s", ErrInvalidApp, app)
	case "":
		return fmt.Errorf("%w: <no app provided>", ErrInvalidApp)
	case "All":
		t.exec(event, TrigLidarrBackup)
		t.exec(event, TrigProwlarrBackup)
		t.exec(event, TrigRadarrBackup)
		t.exec(event, TrigReadarrBackup)
		t.exec(event, TrigSonarrBackup)
	case starr.Lidarr:
		t.exec(event, TrigLidarrBackup)
	case starr.Prowlarr:
		t.exec(event, TrigProwlarrBackup)
	case starr.Radarr:
		t.exec(event, TrigRadarrBackup)
	case starr.Readarr:
		t.exec(event, TrigReadarrBackup)
	case starr.Sonarr:
		t.exec(event, TrigSonarrBackup)
	}

	return nil
}

func (c *Config) sendLidarrBackups(event EventType) {
	for idx, app := range c.Apps.Lidarr {
		if app.Backup != mnd.Disabled || event != EventCron {
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

func (c *Config) sendProwlarrBackups(event EventType) {
	for idx, app := range c.Apps.Prowlarr {
		if app.Backup != mnd.Disabled || event != EventCron {
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

func (c *Config) sendRadarrBackups(event EventType) {
	for idx, app := range c.Apps.Radarr {
		if app.Backup != mnd.Disabled || event != EventCron {
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

func (c *Config) sendReadarrBackups(event EventType) {
	for idx, app := range c.Apps.Readarr {
		if app.Backup != mnd.Disabled || event != EventCron {
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

func (c *Config) sendSonarrBackups(event EventType) {
	for idx, app := range c.Apps.Sonarr {
		if app.Backup != mnd.Disabled || event != EventCron {
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

	send := &BackupPayload{
		App:   input.name,
		Int:   input.int,
		Name:  input.cName,
		Files: fileList,
	}

	c.QueueData(&SendRequest{
		Route:      CorruptRoute,
		Event:      input.event,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("%s Backup File List (%d)", input.name, input.int),
		Payload:    send,
	})
}
