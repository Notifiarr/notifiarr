package notifiarr

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golang.org/x/net/context"
	"golift.io/starr"
	"golift.io/xtractr"
)

// Intervals at which these apps database backups are checked for corruption.
const (
	lidarrCorruptCheckDur  = 6*time.Hour + 10*time.Minute
	radarrCorruptCheckDur  = 5*time.Hour + 40*time.Minute
	readarrCorruptCheckDur = 6*time.Hour + 45*time.Minute
	sonarrCorruptCheckDur  = 5*time.Hour + 15*time.Minute
)

// Errors returned by this package.
var (
	ErrNoDBInBackup = fmt.Errorf("no database file found in backup")
)

// BackupInfo contains a pile of information about a Starr database (backup).
// This is the data sent to notifiarr.com.
type BackupInfo struct {
	App    string    `json:"app"`
	Int    int       `json:"instance"`
	File   string    `json:"file"`
	Name   string    `json:"name"`
	Ver    string    `json:"version"`
	Integ  string    `json:"integrity"`
	Quick  string    `json:"quick"`
	Rows   int       `json:"rows"`
	Size   int64     `json:"bytes"`
	Tables int64     `json:"tables"`
	Date   time.Time `json:"date"`
}

func (c *Config) makeCorruptionTriggers() {
	c.Trigger.corruptLidarr = &action{
		Fn:  c.sendLidarrCorruption,
		Msg: "Checking Lidarr instances for database backup corruption.",
		C:   make(chan EventType, 1),
		T:   time.NewTicker(lidarrCorruptCheckDur),
	}
	c.Trigger.corruptRadarr = &action{
		Fn:  c.sendRadarrCorruption,
		Msg: "Checking Radarr instances for database backup corruption.",
		C:   make(chan EventType, 1),
		T:   time.NewTicker(radarrCorruptCheckDur),
	}
	c.Trigger.corruptReadarr = &action{
		Fn:  c.sendReadarrCorruption,
		Msg: "Checking Readarr instances for database backup corruption.",
		C:   make(chan EventType, 1),
		T:   time.NewTicker(readarrCorruptCheckDur),
	}
	c.Trigger.corruptSonarr = &action{
		Fn:  c.sendSonarrCorruption,
		Msg: "Checking Sonarr instances for database backup corruption.",
		C:   make(chan EventType, 1),
		T:   time.NewTicker(sonarrCorruptCheckDur),
	}
}

func (t *Triggers) SendLidarrCorruption(event EventType) {
	if t.stop == nil {
		return
	}

	t.corruptLidarr.C <- event
}

func (t *Triggers) SendRadarrCorruption(event EventType) {
	if t.stop == nil {
		return
	}

	t.corruptRadarr.C <- event
}

func (t *Triggers) SendReadarrCorruption(event EventType) {
	if t.stop == nil {
		return
	}

	t.corruptReadarr.C <- event
}

func (t *Triggers) SendSonarrCorruption(event EventType) {
	if t.stop == nil {
		return
	}

	t.corruptSonarr.C <- event
}

func (c *Config) sendLidarrCorruption(event EventType) {
	for i, app := range c.Apps.Lidarr {
		app.Corrupt = c.sendAndLogAppCorruption(&checkInstanceCorruption{
			event: event,
			last:  app.Corrupt,
			name:  "Lidarr",
			int:   i + 1,
			app:   app,
			cName: app.Name,
		})
	}
}

func (c *Config) sendRadarrCorruption(event EventType) {
	for i, app := range c.Apps.Radarr {
		app.Corrupt = c.sendAndLogAppCorruption(&checkInstanceCorruption{
			event: event,
			last:  app.Corrupt,
			name:  "Radarr",
			int:   i + 1,
			app:   app,
			cName: app.Name,
		})
	}
}

func (c *Config) sendReadarrCorruption(event EventType) {
	for i, app := range c.Apps.Readarr {
		app.Corrupt = c.sendAndLogAppCorruption(&checkInstanceCorruption{
			event: event,
			last:  app.Corrupt,
			name:  "Readarr",
			int:   i + 1,
			app:   app,
			cName: app.Name,
		})
	}
}

func (c *Config) sendSonarrCorruption(event EventType) {
	for i, app := range c.Apps.Sonarr {
		app.Corrupt = c.sendAndLogAppCorruption(&checkInstanceCorruption{
			event: event,
			last:  app.Corrupt,
			name:  "Sonarr",
			int:   i + 1,
			app:   app,
			cName: app.Name,
		})
	}
}

// checkInstanceCorruption is used to abstract all starr apps to reusable methods.
type checkInstanceCorruption struct {
	event EventType
	last  string      // app.Corrupt
	name  string      // Lidarr, Radarr, ..
	cName string      // configured app name
	int   int         // instance ID: 1, 2, 3...
	app   interface { // all starr apps satisfy this interface. yay!
		GetBackupFiles() ([]*starr.BackupFile, error)
		starr.APIer
	}
}

func (c *Config) sendAndLogAppCorruption(input *checkInstanceCorruption) string {
	if input.last == mnd.Disabled || input.last == "" {
		c.Debugf("[%s requested] Disabled: %s Backup File Corruption Check (%d)", input.event, input.name, input.int)
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

	backup, err := input.checkFileCorruption(latest)
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

func (c *checkInstanceCorruption) checkFileCorruption(remotePath string) (*BackupInfo, error) {
	// XXX: Set TMPDIR to configure this.
	folder, err := ioutil.TempDir("", "starr")
	if err != nil {
		return nil, fmt.Errorf("creating temporary folder: %w", err)
	}

	defer os.RemoveAll(folder) // clean up when we're done.

	fileName, err := c.saveBackupFile(remotePath, folder)
	if err != nil {
		return nil, err
	}

	_, newFiles, err := xtractr.ExtractZIP(&xtractr.XFile{
		FilePath:  fileName,
		OutputDir: folder,
		FileMode:  mnd.Mode0600,
		DirMode:   mnd.Mode0750,
	})
	if err != nil {
		return nil, fmt.Errorf("extracting backup zip file: %w", err)
	}

	return c.checkCorruptFiles(newFiles)
}

func (c *checkInstanceCorruption) saveBackupFile(remotePath, localPath string) (string, error) {
	reader, status, err := c.app.GetBody(context.Background(), remotePath, nil)
	if err != nil {
		return "", fmt.Errorf("getting http response body: %w", err)
	}
	defer reader.Close()

	if status >= http.StatusMultipleChoices && status <= http.StatusPermanentRedirect {
		if err := c.app.Login(); err != nil {
			return "", fmt.Errorf("(%d) %w: you may need to set a username and password to download backup files: %s",
				status, err, remotePath)
		}

		// Try again after logging in.
		reader, status, err = c.app.GetBody(context.Background(), remotePath, nil)
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

func (c *checkInstanceCorruption) checkCorruptFiles(fileList []string) (*BackupInfo, error) {
	for _, filePath := range fileList {
		if path.Ext(filePath) == ".db" {
			return c.checkCorruptSQLite(filePath)
		}
	}

	return nil, ErrNoDBInBackup
}

func (c *checkInstanceCorruption) checkCorruptSQLite(filePath string) (*BackupInfo, error) {
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
		Tables: c.getSQLLiteRowInt64(conn, "SELECT count(*) FROM sqlite_master WHERE type = 'table'"),
	}
	backup.Ver, _ = c.getSQLLiteRowString(conn, "select sqlite_version()")
	backup.Integ, backup.Rows = c.getSQLLiteRowString(conn, "PRAGMA integrity_check")
	backup.Quick, _ = c.getSQLLiteRowString(conn, "PRAGMA quick_check")

	return backup, nil
}

func (c *checkInstanceCorruption) getSQLLiteRowString(conn *sql.DB, sql string) (string, int) {
	text := "<no data returned>"
	count := 0

	rows, err := conn.Query(sql)
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

func (c *checkInstanceCorruption) getSQLLiteRowInt64(conn *sql.DB, sql string) int64 {
	rows, err := conn.Query(sql)
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
