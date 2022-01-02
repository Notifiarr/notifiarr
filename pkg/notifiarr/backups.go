package notifiarr

import (
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr"
)

// BackupPayload is the data we send to notifiarr.
type BackupPayload struct {
	App   string              `json:"app"`
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

// Built-in backup check timers. Non-adjustable.
const (
	lidarrBackupCheckDur   = 6*time.Hour + 10*time.Minute
	prowlarrBackupCheckDur = 6*time.Hour + 20*time.Minute
	radarrBackupCheckDur   = 6*time.Hour + 30*time.Minute
	readarrBackupCheckDur  = 6*time.Hour + 40*time.Minute
	sonarrBackupCheckDur   = 6*time.Hour + 50*time.Minute
)

func (c *Config) makeBackupTriggers() {
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

// SendAllStarrBackups sends backups from all apps.
func (t *Triggers) SendAllStarrBackups(event EventType) {
	t.SendSonarrBackups(event)
	t.SendProwlarrBackups(event)
	t.SendRadarrBackups(event)
	t.SendReadarrBackups(event)
	t.SendSonarrBackups(event)
}

func (t *Triggers) SendLidarrBackups(event EventType) {
	if trig := t.get(TrigLidarrBackup); trig != nil && t.stop != nil {
		trig.C <- event
	}
}

func (t *Triggers) SendProwlarrBackups(event EventType) {
	if trig := t.get(TrigProwlarrBackup); trig != nil && t.stop != nil {
		trig.C <- event
	}
}

func (t *Triggers) SendRadarrBackups(event EventType) {
	if trig := t.get(TrigRadarrBackup); trig != nil && t.stop != nil {
		trig.C <- event
	}
}

func (t *Triggers) SendReadarrBackups(event EventType) {
	if trig := t.get(TrigReadarrBackup); trig != nil && t.stop != nil {
		trig.C <- event
	}
}

func (t *Triggers) SendSonarrBackups(event EventType) {
	if trig := t.get(TrigSonarrBackup); trig != nil && t.stop != nil {
		trig.C <- event
	}
}

func (c *Config) sendLidarrBackups(event EventType) {
	for i, app := range c.Apps.Lidarr {
		if app.Backup != mnd.Disabled || event != EventCron {
			c.sendBackups(&checkInstanceCorruption{
				event: event,
				name:  string(starr.Lidarr),
				int:   i + 1,
				app:   app,
			})
		}
	}
}

func (c *Config) sendProwlarrBackups(event EventType) {
	for i, app := range c.Apps.Prowlarr {
		if app.Backup != mnd.Disabled || event != EventCron {
			c.sendBackups(&checkInstanceCorruption{
				event: event,
				name:  string(starr.Prowlarr),
				int:   i + 1,
				app:   app,
			})
		}
	}
}

func (c *Config) sendRadarrBackups(event EventType) {
	for i, app := range c.Apps.Radarr {
		if app.Backup != mnd.Disabled || event != EventCron {
			c.sendBackups(&checkInstanceCorruption{
				event: event,
				name:  string(starr.Radarr),
				int:   i + 1,
				app:   app,
			})
		}
	}
}

func (c *Config) sendReadarrBackups(event EventType) {
	for i, app := range c.Apps.Readarr {
		if app.Backup != mnd.Disabled || event != EventCron {
			c.sendBackups(&checkInstanceCorruption{
				event: event,
				name:  string(starr.Readarr),
				int:   i + 1,
				app:   app,
			})
		}
	}
}

func (c *Config) sendSonarrBackups(event EventType) {
	for i, app := range c.Apps.Sonarr {
		if app.Backup != mnd.Disabled || event != EventCron {
			c.sendBackups(&checkInstanceCorruption{
				event: event,
				name:  string(starr.Sonarr),
				cName: app.Name,
				int:   i + 1,
				app:   app,
			})
		}
	}
}

func (c *Config) sendBackups(input *checkInstanceCorruption) {
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

	resp, err := c.SendData(BackupRoute.Path(input.event), send, true)
	if err != nil {
		c.Errorf("[%s requested] Sending %s Backup File List to Notifiarr (%d): %v: %s",
			input.event, input.name, input.int, err, resp)
	} else {
		c.Printf("[%s requested] Sent %s Backup File List to Notifiarr (%d): %s",
			input.event, input.name, input.int, resp)
	}
}
