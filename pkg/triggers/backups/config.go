package backups

import (
	"context"
	"errors"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/starr"
)

// Action is the data needed by the backups package to perform backup ad corruption checks.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	lidarr   map[int]string
	prowlarr map[int]string
	radarr   map[int]string
	readarr  map[int]string
	sonarr   map[int]string
}

// Errors returned by this package.
var (
	ErrNoDBInBackup = errors.New("no database file found in backup")
)

// Intervals at which these apps database backups are checked for corruption.
const (
	checkInterval = 5 * time.Hour
	randomMinutes = 60
)

// Trigger Types.
const (
	TrigLidarrCorrupt   common.TriggerName = "Checking Lidarr for database backup corruption."
	TrigProwlarrCorrupt common.TriggerName = "Checking Prowlarr for database backup corruption."
	TrigRadarrCorrupt   common.TriggerName = "Checking Radarr for database backup corruption."
	TrigReadarrCorrupt  common.TriggerName = "Checking Readarr for database backup corruption."
	TrigSonarrCorrupt   common.TriggerName = "Checking Sonarr for database backup corruption."
	TrigLidarrBackup    common.TriggerName = "Sending Lidarr Backup File List to Notifiarr."
	TrigProwlarrBackup  common.TriggerName = "Sending Prowlarr Backup File List to Notifiarr."
	TrigRadarrBackup    common.TriggerName = "Sending Radarr Backup File List to Notifiarr."
	TrigReadarrBackup   common.TriggerName = "Sending Readarr Backup File List to Notifiarr."
	TrigSonarrBackup    common.TriggerName = "Sending Sonarr Backup File List to Notifiarr."
)

// Info contains a pile of information about a Starr database (backup).
// This is the data sent to notifiarr.com.
type Info struct {
	App    starr.App `json:"app"`
	Int    int       `json:"instance"`
	Name   string    `json:"name"`
	File   string    `json:"file,omitempty"`
	Ver    string    `json:"version,omitempty"`
	Integ  string    `json:"integrity,omitempty"`
	Quick  string    `json:"quick,omitempty"`
	Rows   int       `json:"rows,omitempty"`
	Size   int64     `json:"bytes,omitempty"`
	Tables int64     `json:"tables,omitempty"`
	Date   time.Time `json:"date,omitempty"`
}

// genericInstance is used to abstract all starr apps to reusable methods.
// It's also used in the go file.
type genericInstance struct {
	skip  bool
	event website.EventType
	last  string      // app.Corrupt
	name  starr.App   // Lidarr, Radarr, ..
	cName string      // configured app name
	int   int         // instance ID: 1, 2, 3...
	app   interface { // all starr apps satisfy this interface. yay!
		GetBackupFiles() ([]*starr.BackupFile, error)
		GetBackupFilesContext(ctx context.Context) ([]*starr.BackupFile, error)
		starr.APIer
	}
}

// Payload is the backups and corruption data we send to notifiarr.
type Payload struct {
	App   starr.App           `json:"app"`
	Int   int                 `json:"instance"`
	Name  string              `json:"name"`
	Files []*starr.BackupFile `json:"backups"`
}

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{
		Config:   config,
		lidarr:   make(map[int]string),
		prowlarr: make(map[int]string),
		radarr:   make(map[int]string),
		readarr:  make(map[int]string),
		sonarr:   make(map[int]string),
	}}
}

// Create sets up all the triggers.
func (a *Action) Create() {
	info := clientinfo.Get()
	a.cmd.makeBackupTriggersLidarr(info)
	a.cmd.makeBackupTriggersRadarr(info)
	a.cmd.makeBackupTriggersReadarr(info)
	a.cmd.makeBackupTriggersSonarr(info)
	a.cmd.makeBackupTriggersProwlarr(info)
	a.cmd.makeCorruptionTriggersLidarr(info)
	a.cmd.makeCorruptionTriggersRadarr(info)
	a.cmd.makeCorruptionTriggersReadarr(info)
	a.cmd.makeCorruptionTriggersSonarr(info)
	a.cmd.makeCorruptionTriggersProwlarr(info)
}
