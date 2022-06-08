package notifiarr

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/starr"
	"golift.io/xtractr"
)

// Intervals at which these apps database backups are checked for corruption.
const (
	lidarrCorruptCheckDur   = 5*time.Hour + 10*time.Minute
	prowlarrCorruptCheckDur = 5*time.Hour + 20*time.Minute
	radarrCorruptCheckDur   = 5*time.Hour + 30*time.Minute
	readarrCorruptCheckDur  = 5*time.Hour + 40*time.Minute
	sonarrCorruptCheckDur   = 5*time.Hour + 50*time.Minute
	maxCheckTime            = 10 * time.Minute
)

// Trigger Types.
const (
	TrigLidarrCorrupt   TriggerName = "Checking Lidarr for database backup corruption."
	TrigProwlarrCorrupt TriggerName = "Checking Prowlarr for database backup corruption."
	TrigRadarrCorrupt   TriggerName = "Checking Radarr for database backup corruption."
	TrigReadarrCorrupt  TriggerName = "Checking Readarr for database backup corruption."
	TrigSonarrCorrupt   TriggerName = "Checking Sonarr for database backup corruption."
)

// Errors returned by this package.
var (
	ErrNoDBInBackup = fmt.Errorf("no database file found in backup")
)

// BackupInfo contains a pile of information about a Starr database (backup).
// This is the data sent to notifiarr.com.
type BackupInfo struct {
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
// It's also used in the backups.go file.
type genericInstance struct {
	event EventType
	last  string      // app.Corrupt
	name  starr.App   // Lidarr, Radarr, ..
	cName string      // configured app name
	int   int         // instance ID: 1, 2, 3...
	app   interface { // all starr apps satisfy this interface. yay!
		GetBackupFiles() ([]*starr.BackupFile, error)
		starr.APIer
	}
}

func (c *Config) makeCorruptionTriggers() {
	c.Trigger.add(&action{
		Name: TrigLidarrCorrupt,
		Fn:   c.sendLidarrCorruption,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(lidarrCorruptCheckDur),
	}, &action{
		Name: TrigProwlarrCorrupt,
		Fn:   c.sendProwlarrCorruption,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(prowlarrCorruptCheckDur),
	}, &action{
		Name: TrigRadarrCorrupt,
		Fn:   c.sendRadarrCorruption,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(radarrCorruptCheckDur),
	}, &action{
		Name: TrigReadarrCorrupt,
		Fn:   c.sendReadarrCorruption,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(readarrCorruptCheckDur),
	}, &action{
		Name: TrigSonarrCorrupt,
		Fn:   c.sendSonarrCorruption,
		C:    make(chan EventType, 1),
		T:    time.NewTicker(sonarrCorruptCheckDur),
	})
}

func (t *Triggers) Corruption(event EventType, app starr.App) error {
	switch app { //nolint:exhaustive // we do not need them all here.
	default:
		return fmt.Errorf("%w: %s", ErrInvalidApp, app)
	case "":
		return fmt.Errorf("%w: <no app provided>", ErrInvalidApp)
	case "All":
		t.exec(event, TrigLidarrCorrupt)
		t.exec(event, TrigProwlarrCorrupt)
		t.exec(event, TrigRadarrCorrupt)
		t.exec(event, TrigReadarrCorrupt)
		t.exec(event, TrigSonarrCorrupt)
	case starr.Lidarr:
		t.exec(event, TrigLidarrCorrupt)
	case starr.Prowlarr:
		t.exec(event, TrigProwlarrCorrupt)
	case starr.Radarr:
		t.exec(event, TrigRadarrCorrupt)
	case starr.Readarr:
		t.exec(event, TrigReadarrCorrupt)
	case starr.Sonarr:
		t.exec(event, TrigSonarrCorrupt)
	}

	return nil
}

func (c *Config) sendLidarrCorruption(event EventType) {
	for i, app := range c.Apps.Lidarr {
		app.Corrupt = c.sendAndLogAppCorruption(&genericInstance{
			event: event,
			last:  app.Corrupt,
			name:  starr.Lidarr,
			int:   i + 1,
			app:   app.Lidarr,
			cName: app.Name,
		})
	}
}

func (c *Config) sendProwlarrCorruption(event EventType) {
	for i, app := range c.Apps.Prowlarr {
		app.Corrupt = c.sendAndLogAppCorruption(&genericInstance{
			event: event,
			last:  app.Corrupt,
			name:  starr.Prowlarr,
			int:   i + 1,
			app:   app.Prowlarr,
			cName: app.Name,
		})
	}
}

func (c *Config) sendRadarrCorruption(event EventType) {
	for i, app := range c.Apps.Radarr {
		app.Corrupt = c.sendAndLogAppCorruption(&genericInstance{
			event: event,
			last:  app.Corrupt,
			name:  starr.Radarr,
			int:   i + 1,
			app:   app.Radarr,
			cName: app.Name,
		})
	}
}

func (c *Config) sendReadarrCorruption(event EventType) {
	for i, app := range c.Apps.Readarr {
		app.Corrupt = c.sendAndLogAppCorruption(&genericInstance{
			event: event,
			last:  app.Corrupt,
			name:  starr.Readarr,
			int:   i + 1,
			app:   app.Readarr,
			cName: app.Name,
		})
	}
}

func (c *Config) sendSonarrCorruption(event EventType) {
	for i, app := range c.Apps.Sonarr {
		app.Corrupt = c.sendAndLogAppCorruption(&genericInstance{
			event: event,
			last:  app.Corrupt,
			name:  starr.Sonarr,
			int:   i + 1,
			app:   app.Sonarr,
			cName: app.Name,
		})
	}
}

func (c *Config) sendAndLogAppCorruption(input *genericInstance) string {
	ctx, cancel := context.WithTimeout(context.Background(), maxCheckTime)
	defer cancel()

	if input.last == mnd.Disabled || input.last == "" {
		c.Printf("[%s requested] Disabled: %s Backup File Corruption Check (%d), Last File: '%s'",
			input.event, input.name, input.int, input.last)
		return input.last
	}

	fileList, err := input.app.GetBackupFiles()
	if err != nil {
		c.Errorf("[%s requested] Getting %s Backup Files (%d): %v", input.event, input.name, input.int, err)
		return input.last
	} else if len(fileList) == 0 {
		c.Printf("[%s requested] %s has no backup files (%d)", input.event, input.name, input.int)
		return input.last
	}

	latest := fileList[0].Path
	if input.last == latest {
		c.Printf("[%s requested] %s Backup DB Check (%d): already checked latest file: %s",
			input.event, input.name, input.int, latest)
		return latest
	}

	backup, err := c.checkBackupFileCorruption(ctx, input, latest)
	if err != nil {
		// XXX: Send "error" to notifirr.com here?
		c.Errorf("[%s requested] Checking %s Backup File Corruption (%d): %s: %v",
			input.event, input.name, input.int, latest, err)
		return input.last
	}

	backup.App = input.name
	backup.Int = input.int
	backup.Name = input.cName
	backup.File = latest
	backup.Date = fileList[0].Time.Round(time.Second)

	if resp, err := c.SendData(CorruptRoute.Path(input.event), backup, true); err != nil {
		c.Errorf("[%s requested] Sending %s Backup File Corruption Info to Notifiarr (%d): %v: %s: "+
			"OK: ver:%s, integ:%s, quick:%s, tables:%d, size:%d. %s",
			input.event, input.name, input.int, err, latest, backup.Ver, backup.Integ,
			backup.Quick, backup.Tables, backup.Size, resp)
	} else {
		c.Printf("[%s requested] Checking %s Backup File Corruption (%d): %s: "+
			"OK: ver:%s, integ:%s, quick:%s, tables:%d, size:%d. %s",
			input.event, input.name, input.int, latest, backup.Ver, backup.Integ,
			backup.Quick, backup.Tables, backup.Size, resp)
	}

	return backup.Name
}

func (c *Config) checkBackupFileCorruption(
	ctx context.Context,
	input *genericInstance,
	remotePath string,
) (*BackupInfo, error) {
	// XXX: Set TMPDIR to configure this.
	folder, err := ioutil.TempDir("", "notifiarr_tmp_dir")
	if err != nil {
		return nil, fmt.Errorf("creating temporary folder: %w", err)
	}

	defer os.RemoveAll(folder) // clean up when we're done.
	c.Debugf("[%s requested] Downloading %s backup file (%d): %s", input.event, input.name, input.int, remotePath)

	fileName, err := input.saveBackupFile(ctx, remotePath, folder)
	if err != nil {
		return nil, err
	}

	c.Debugf("[%s requested] Extracting downloaded %s backup file (%d): %s", input.event, input.name, input.int, fileName)

	_, newFiles, err := xtractr.ExtractZIP(&xtractr.XFile{
		FilePath:  fileName,
		OutputDir: folder,
		FileMode:  mnd.Mode0600,
		DirMode:   mnd.Mode0750,
	})
	if err != nil {
		return nil, fmt.Errorf("extracting backup zip file: %w", err)
	}

	for _, filePath := range newFiles {
		if path.Ext(filePath) == ".db" {
			c.Debugf("[%s requested] Checking %s backup sqlite3 file (%d): %s",
				input.event, input.name, input.int, filePath)
			return input.checkCorruptSQLite(ctx, filePath)
		}
	}

	return nil, ErrNoDBInBackup
}

func (c *genericInstance) saveBackupFile(
	ctx context.Context,
	remotePath,
	localPath string,
) (string, error) {
	reader, status, err := c.app.GetBody(ctx, remotePath, nil)
	if err != nil {
		return "", fmt.Errorf("getting http response body: %w", err)
	}
	defer reader.Close()

	if status >= http.StatusMultipleChoices && status <= http.StatusPermanentRedirect {
		if err := c.app.Login(ctx); err != nil {
			return "", fmt.Errorf("(%d) %w: you may need to set a username and password to download backup files: %s",
				status, err, remotePath)
		}

		// Try again after logging in.
		reader, status, err = c.app.GetBody(ctx, remotePath, nil)
		if err != nil {
			return "", fmt.Errorf("getting http response body: %w", err)
		}
		defer reader.Close()
	}

	if status != http.StatusOK {
		return "", fmt.Errorf("(%d) %w: %s", status, ErrNon200, remotePath)
	}

	file, err := ioutil.TempFile(localPath, "starr_"+path.Base(remotePath)+".*."+path.Ext(remotePath))
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %w", err)
	}
	defer file.Close()

	size, err := io.Copy(file, reader)
	if err != nil {
		return "", fmt.Errorf("writing temporary file: %d, %w", size, err)
	}

	return file.Name(), nil
}

func (c *genericInstance) checkCorruptSQLite(
	ctx context.Context,
	filePath string,
) (*BackupInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("checking db file: %w", err)
	}

	conn, err := sql.Open("sqlite", filePath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite DB: %w", err)
	}
	defer conn.Close()

	backup := &BackupInfo{
		Name:   filePath,
		Size:   fileInfo.Size(),
		Tables: c.getSQLLiteRowInt64(ctx, conn, "SELECT count(*) FROM sqlite_master WHERE type = 'table'"),
	}
	backup.Ver, _ = c.getSQLLiteRowString(ctx, conn, "select sqlite_version()")
	backup.Integ, backup.Rows = c.getSQLLiteRowString(ctx, conn, "PRAGMA integrity_check")
	backup.Quick, _ = c.getSQLLiteRowString(ctx, conn, "PRAGMA quick_check")

	return backup, nil
}

func (c *genericInstance) getSQLLiteRowString(
	ctx context.Context,
	conn *sql.DB,
	sql string,
) (string, int) {
	text := "<no data returned>"
	count := 0

	rows, err := conn.QueryContext(ctx, sql)
	if err != nil {
		return fmt.Sprintf("%s: running DB query: %v", text, err), 0
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return fmt.Sprintf("%s: reading DB rows: %v", text, err), 0
	}

	for rows.Next() {
		if err := rows.Scan(&text); err != nil {
			return fmt.Sprintf("%s: reading DB query: %v", text, err), 0
		}

		count++
	}

	return text, count
}

func (c *genericInstance) getSQLLiteRowInt64(
	ctx context.Context,
	conn *sql.DB,
	sql string,
) int64 {
	rows, err := conn.QueryContext(ctx, sql)
	if err != nil {
		return 0
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return 0
	}

	if rows.Next() {
		var i int64
		if err := rows.Scan(&i); err != nil {
			return 0
		}

		return i
	}

	return 0
}
