package backups

import (
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"golift.io/starr"
)

// Action is the data needed by the backups package to perform backup ad corruption checks.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
}

// Errors returned by this package.
var (
	ErrNoDBInBackup = fmt.Errorf("no database file found in backup")
)

// Intervals at which these apps database backups are checked for corruption.
const (
	lidarrCorruptCheckDur   = 5*time.Hour + 10*time.Minute
	prowlarrCorruptCheckDur = 5*time.Hour + 20*time.Minute
	radarrCorruptCheckDur   = 5*time.Hour + 30*time.Minute
	readarrCorruptCheckDur  = 5*time.Hour + 40*time.Minute
	sonarrCorruptCheckDur   = 5*time.Hour + 50*time.Minute
	lidarrBackupCheckDur    = 6*time.Hour + 10*time.Minute
	prowlarrBackupCheckDur  = 6*time.Hour + 20*time.Minute
	radarrBackupCheckDur    = 6*time.Hour + 30*time.Minute
	readarrBackupCheckDur   = 6*time.Hour + 40*time.Minute
	sonarrBackupCheckDur    = 6*time.Hour + 50*time.Minute
	maxCheckTime            = 10 * time.Minute
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
	return &Action{cmd: &cmd{Config: config}}
}

// Create sets up all the triggers.
func (a *Action) Create() {
	a.cmd.makeBackupTriggersLidarr()
	a.cmd.makeBackupTriggersRadarr()
	a.cmd.makeBackupTriggersReadarr()
	a.cmd.makeBackupTriggersSonarr()
	a.cmd.makeBackupTriggersProwlarr()
	a.cmd.makeCorruptionTriggersLidarr()
	a.cmd.makeCorruptionTriggersRadarr()
	a.cmd.makeCorruptionTriggersReadarr()
	a.cmd.makeCorruptionTriggersSonarr()
	a.cmd.makeCorruptionTriggersProwlarr()
}
