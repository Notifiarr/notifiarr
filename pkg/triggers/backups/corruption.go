package backups

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
	"golift.io/starr"
	"golift.io/xtractr"
)

// Corruption initializes a corruption check for all instances of the provided app.
func (a *Action) Corruption(input *common.ActionInput, app starr.App) error {
	switch app { //nolint:exhaustive // We only check starr apps.
	default:
		return fmt.Errorf("%w: %s", common.ErrInvalidApp, app)
	case "":
		return fmt.Errorf("%w: <no app provided>", common.ErrInvalidApp)
	case "All":
		a.cmd.Exec(input, TrigLidarrCorrupt)
		a.cmd.Exec(input, TrigProwlarrCorrupt)
		a.cmd.Exec(input, TrigRadarrCorrupt)
		a.cmd.Exec(input, TrigReadarrCorrupt)
		a.cmd.Exec(input, TrigSonarrCorrupt)
	case starr.Lidarr:
		a.cmd.Exec(input, TrigLidarrCorrupt)
	case starr.Prowlarr:
		a.cmd.Exec(input, TrigProwlarrCorrupt)
	case starr.Radarr:
		a.cmd.Exec(input, TrigRadarrCorrupt)
	case starr.Readarr:
		a.cmd.Exec(input, TrigReadarrCorrupt)
	case starr.Sonarr:
		a.cmd.Exec(input, TrigSonarrCorrupt)
	}

	return nil
}

func (c *cmd) makeCorruptionTriggersLidarr(info *clientinfo.ClientInfo) {
	action := &common.Action{
		Name: TrigLidarrCorrupt,
		Fn:   c.sendLidarrCorruption,
		C:    make(chan *common.ActionInput, 1),
	}
	defer c.Add(action)

	if info == nil {
		return
	}

	for idx, app := range c.Apps.Lidarr {
		if app.Enabled() {
			c.lidarr[idx] = info.Actions.Apps.Lidarr.Corrupt(idx + 1) // mandatory
			if c.lidarr[idx] != mnd.Disabled {
				randomTime := time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Second +
					time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Minute
				action.D = cnfg.Duration{Duration: checkInterval + randomTime}
			}
		}
	}
}

func (c *cmd) makeCorruptionTriggersProwlarr(info *clientinfo.ClientInfo) {
	action := &common.Action{
		Name: TrigProwlarrCorrupt,
		Fn:   c.sendProwlarrCorruption,
		C:    make(chan *common.ActionInput, 1),
	}
	defer c.Add(action)

	if info == nil {
		return
	}

	for idx, app := range c.Apps.Prowlarr {
		if app.Enabled() {
			c.prowlarr[idx] = info.Actions.Apps.Prowlarr.Corrupt(idx + 1) // mandatory
			if c.prowlarr[idx] != mnd.Disabled {
				randomTime := time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Second +
					time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Minute
				action.D = cnfg.Duration{Duration: checkInterval + randomTime}
			}
		}
	}
}

func (c *cmd) makeCorruptionTriggersRadarr(info *clientinfo.ClientInfo) {
	action := &common.Action{
		Name: TrigRadarrCorrupt,
		Fn:   c.sendRadarrCorruption,
		C:    make(chan *common.ActionInput, 1),
	}
	defer c.Add(action)

	if info == nil {
		return
	}

	for idx, app := range c.Apps.Radarr {
		if app.Enabled() {
			c.radarr[idx] = info.Actions.Apps.Radarr.Corrupt(idx + 1) // mandatory
			if c.radarr[idx] != mnd.Disabled {
				randomTime := time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Second +
					time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Minute
				action.D = cnfg.Duration{Duration: checkInterval + randomTime}
			}
		}
	}
}

func (c *cmd) makeCorruptionTriggersReadarr(info *clientinfo.ClientInfo) {
	action := &common.Action{
		Name: TrigReadarrCorrupt,
		Fn:   c.sendReadarrCorruption,
		C:    make(chan *common.ActionInput, 1),
	}
	defer c.Add(action)

	if info == nil {
		return
	}

	for idx, app := range c.Apps.Readarr {
		if app.Enabled() {
			c.readarr[idx] = info.Actions.Apps.Readarr.Corrupt(idx + 1) // mandatory
			if c.readarr[idx] != mnd.Disabled {
				randomTime := time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Second +
					time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Minute
				action.D = cnfg.Duration{Duration: checkInterval + randomTime}
			}
		}
	}
}

func (c *cmd) makeCorruptionTriggersSonarr(info *clientinfo.ClientInfo) {
	action := &common.Action{
		Name: TrigSonarrCorrupt,
		Fn:   c.sendSonarrCorruption,
		C:    make(chan *common.ActionInput, 1),
	}
	defer c.Add(action)

	if info == nil {
		return
	}

	for idx, app := range c.Apps.Sonarr {
		if app.Enabled() {
			c.sonarr[idx] = info.Actions.Apps.Sonarr.Corrupt(idx + 1)
			if c.sonarr[idx] != mnd.Disabled {
				randomTime := time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Second +
					time.Duration(c.Config.Rand().Intn(randomMinutes))*time.Minute
				action.D = cnfg.Duration{Duration: checkInterval + randomTime}
			}
		}
	}
}

func (c *cmd) sendLidarrCorruption(ctx context.Context, input *common.ActionInput) {
	for idx, app := range c.Apps.Lidarr {
		c.lidarr[idx] = c.sendAndLogAppCorruption(ctx, &genericInstance{
			event: input.Type,
			last:  c.lidarr[idx],
			name:  starr.Lidarr,
			int:   idx + 1,
			app:   app.Lidarr,
			cName: app.Name,
			skip:  !app.Enabled(),
		})
	}
}

func (c *cmd) sendProwlarrCorruption(ctx context.Context, input *common.ActionInput) {
	for idx, app := range c.Apps.Prowlarr {
		c.prowlarr[idx] = c.sendAndLogAppCorruption(ctx, &genericInstance{
			event: input.Type,
			last:  c.prowlarr[idx],
			name:  starr.Prowlarr,
			int:   idx + 1,
			app:   app.Prowlarr,
			cName: app.Name,
			skip:  !app.Enabled(),
		})
	}
}

func (c *cmd) sendRadarrCorruption(ctx context.Context, input *common.ActionInput) {
	for idx, app := range c.Apps.Radarr {
		c.radarr[idx] = c.sendAndLogAppCorruption(ctx, &genericInstance{
			event: input.Type,
			last:  c.radarr[idx],
			name:  starr.Radarr,
			int:   idx + 1,
			app:   app.Radarr,
			cName: app.Name,
			skip:  !app.Enabled(),
		})
	}
}

func (c *cmd) sendReadarrCorruption(ctx context.Context, input *common.ActionInput) {
	for idx, app := range c.Apps.Readarr {
		c.readarr[idx] = c.sendAndLogAppCorruption(ctx, &genericInstance{
			event: input.Type,
			last:  c.readarr[idx],
			name:  starr.Readarr,
			int:   idx + 1,
			app:   app.Readarr,
			cName: app.Name,
			skip:  !app.Enabled(),
		})
	}
}

func (c *cmd) sendSonarrCorruption(ctx context.Context, input *common.ActionInput) {
	for idx, app := range c.Apps.Sonarr {
		c.sonarr[idx] = c.sendAndLogAppCorruption(ctx, &genericInstance{
			event: input.Type,
			last:  c.sonarr[idx],
			name:  starr.Sonarr,
			int:   idx + 1,
			app:   app.Sonarr,
			cName: app.Name,
			skip:  !app.Enabled(),
		})
	}
}

func (c *cmd) sendAndLogAppCorruption(ctx context.Context, input *genericInstance) string { //nolint:cyclop
	if input.skip {
		c.Debugf("Skipping corruption check on %s: %s (%d), instance disabled.", input.name, input.cName, input.int)
		return input.last
	}

	if (input.last == mnd.Disabled || input.last == "") && input.event == website.EventCron {
		c.Debugf("Skipping corruption check on %s: %s (%d), corruption checking disabled, last: %s",
			input.name, input.cName, input.int, input.last)
		return input.last
	}

	fileList, err := input.app.GetBackupFilesContext(ctx)
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
		return input.last
	}

	backup, err := c.checkBackupFileCorruption(ctx, input, latest)
	if err != nil {
		c.Errorf("[%s requested] Checking %s Backup File Corruption (%d): %s: %v (last file: %s)",
			input.event, input.name, input.int, latest, err, input.last)
		return input.last
	}

	backup.App = input.name
	backup.Int = input.int
	backup.Name = input.cName
	backup.File = latest
	backup.Date = fileList[0].Time.Round(time.Second)

	c.SendData(&website.Request{
		Route:      website.CorruptRoute,
		Event:      input.event,
		LogPayload: true,
		LogMsg: fmt.Sprintf("%s Backup File Corruption Info (%d): %s: OK: ver:%s, integ:%s, quick:%s, tables:%d, size:%d",
			input.name, input.int, latest, backup.Ver, backup.Integ, backup.Quick, backup.Tables, backup.Size),
		Payload: backup,
	})

	if input.last == mnd.Disabled || input.last == "" {
		return input.last
	}

	return latest
}

func (c *cmd) checkBackupFileCorruption(
	ctx context.Context,
	input *genericInstance,
	remotePath string,
) (*Info, error) {
	folder, err := os.MkdirTemp("", "notifiarr_tmp_dir")
	if err != nil {
		const moreInfo = "click here for help with this: https://notifiarr.wiki/en/Client/Configuration#tmp-not-found"
		return nil, fmt.Errorf("creating temporary folder: %w - %s", err, moreInfo)
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
	resp, err := c.app.Get(ctx, starr.Request{URI: remotePath})
	if err != nil && !errors.Is(err, starr.ErrInvalidStatusCode) {
		return "", fmt.Errorf("getting http response body: %w", err)
	}

	if !errors.Is(err, starr.ErrInvalidStatusCode) {
		defer resp.Body.Close()
	} else if err := c.app.Login(ctx); err != nil {
		return "", fmt.Errorf("%w: you may need to set a username and password to download backup files: %s", err, remotePath)
	} else {
		// Try again after logging in.
		resp, err = c.app.Get(ctx, starr.Request{URI: remotePath})
		if err != nil {
			return "", fmt.Errorf("getting http response body after logging in: %w", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("(%s) %w: %s", resp.Status, website.ErrNon200, remotePath)
	}

	file, err := os.CreateTemp(localPath, "starr_"+path.Base(remotePath)+".*."+path.Ext(remotePath))
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %w", err)
	}
	defer file.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("writing temporary file: %d, %w", size, err)
	}

	return file.Name(), nil
}

func (c *genericInstance) checkCorruptSQLite(
	ctx context.Context,
	filePath string,
) (*Info, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("checking db file: %w", err)
	}

	conn, err := sql.Open("sqlite", filePath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite DB: %w", err)
	}
	defer conn.Close()

	backup := &Info{
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
