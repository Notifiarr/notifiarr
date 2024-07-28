package snapshot

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	_ "github.com/go-sql-driver/mysql" // We use mysql driver, this is how it's loaded.
	"golift.io/cnfg"
)

// MySQLConfig allows us to gather a process list for the snapshot.
type MySQLConfig struct {
	Name    string        `json:"name"    toml:"name"    xml:"name"`
	Host    string        `json:"host"    toml:"host"    xml:"host"`
	User    string        `json:"-"       toml:"user"    xml:"user"`
	Pass    string        `json:"-"       toml:"pass"    xml:"pass"`
	Timeout cnfg.Duration `json:"timeout" toml:"timeout" xml:"timeout"`
	// Only used by service checks, snapshot interval is used for mysql.
	Interval cnfg.Duration `json:"interval" toml:"interval" xml:"interval"`
}

// MySQLProcesses allows us to manipulate our list with methods.
type MySQLProcesses []*MySQLProcess

// MySQLProcess represents the data returned from SHOW PROCESS LIST.
type MySQLProcess struct {
	ID       int64      `json:"id"`
	User     string     `json:"user"`
	Host     string     `json:"host"`
	DB       NullString `json:"db"`
	Cmd      string     `json:"command"`
	Time     int64      `json:"time"`
	State    string     `json:"state"`
	Info     NullString `json:"info"`
	Progress float64    `json:"progress"` // mariadb
}

type NullString struct {
	sql.NullString
}

type MySQLStatus map[string]interface{}

type MySQLServerData struct {
	Name      string         `json:"name"`
	Processes MySQLProcesses `json:"processes"`
	GStatus   MySQLStatus    `json:"globalstatus"`
}

// MarshalJSON makes the output from sql.NullString not suck.
func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte(`"NULL"`), nil
	}

	return json.Marshal(strings.TrimSpace(n.String)) //nolint:wrapcheck
}

// GetMySQL grabs the process list from a bunch of servers.
func (s *Snapshot) GetMySQL(ctx context.Context, servers []*MySQLConfig, limit int) []error {
	s.MySQL = make(map[string]*MySQLServerData)

	var errs []error

	for _, server := range servers {
		if server.Host == "" {
			continue
		}

		procs, status, err := getMySQL(ctx, server)
		if err != nil {
			errs = append(errs, err)
		}

		s.MySQL[server.Host] = &MySQLServerData{
			Name:      server.Name,
			Processes: procs,
			GStatus:   status,
		}
	}

	for _, v := range s.MySQL {
		sort.Sort(v.Processes)
		v.Processes.Shrink(limit)
	}

	return errs
}

func getMySQL(ctx context.Context, mysql *MySQLConfig) (MySQLProcesses, MySQLStatus, error) {
	hostID := mysql.Host
	if mysql.Name != "" {
		hostID = mysql.Name
	}

	host := "@tcp(" + mysql.Host + ")"
	if strings.HasPrefix(mysql.Host, "@") {
		host = mysql.Host
	}

	dbase, err := sql.Open("mysql", mysql.User+":"+mysql.Pass+host+"/")
	if err != nil {
		return nil, nil, fmt.Errorf("mysql server %s: connecting: %w", hostID, err)
	}
	defer dbase.Close()

	list, err := scanMySQLProcessList(ctx, dbase)
	if err != nil {
		return list, nil, fmt.Errorf("mysql server %s: %w", hostID, err)
	}

	status, err := scanMySQLStatus(ctx, dbase)
	if err != nil {
		return list, nil, fmt.Errorf("mysql server %s: %w", hostID, err)
	}

	return list, status, nil
}

func scanMySQLProcessList(ctx context.Context, dbase *sql.DB) (MySQLProcesses, error) {
	mnd.Apps.Add("MySQL&&Process List Queries", 1)

	rows, err := dbase.QueryContext(ctx, "SHOW FULL PROCESSLIST") //nolint:execinquery
	if err != nil {
		mnd.Apps.Add("MySQL&&Errors", 1)
		return nil, fmt.Errorf("getting processes: %w", err)
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		mnd.Apps.Add("MySQL&&Errors", 1)
		return nil, fmt.Errorf("getting processes rows: %w", err)
	}

	var list MySQLProcesses

	const mysqlColCount = 8

	for rows.Next() {
		var pid MySQLProcess
		// for each row, scan the result into our tag composite object.
		if cols, _ := rows.Columns(); len(cols) == mysqlColCount {
			// mysql only has 8 columns
			err = rows.Scan(&pid.ID, &pid.User, &pid.Host, &pid.DB, &pid.Cmd, &pid.Time, &pid.State, &pid.Info)
		} else {
			// mariadb returns 9 columns (adds progress).
			err = rows.Scan(&pid.ID, &pid.User, &pid.Host, &pid.DB, &pid.Cmd, &pid.Time, &pid.State, &pid.Info, &pid.Progress)
		}

		if err != nil {
			mnd.Apps.Add("MySQL&&Errors", 1)
			return nil, fmt.Errorf("scanning process rows: %w", err)
		}

		if pid.Info.Valid {
			pid.Info.String = strings.Join(strings.Fields(pid.Info.String), " ")
		}

		list = append(list, &pid)
	}

	return list, nil
}

func scanMySQLStatus(ctx context.Context, dbase *sql.DB) (MySQLStatus, error) {
	list := make(MySQLStatus)

	for _, name := range []string{
		"Aborted",
		"Bytes",
		"Connection",
		"Created",
		"Handler",
		"Innodb",
		"Key",
		"Open",
		"Q",
		"Slow",
		"Sort",
		"Uptime",
		"Table",
		"Threads",
	} {
		if err := list.processStatus(ctx, dbase, name); err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (m MySQLStatus) processStatus(ctx context.Context, dbase *sql.DB, name string) error {
	mnd.Apps.Add("MySQL&&Global Status Queries", 1)

	rows, err := dbase.QueryContext(ctx, "SHOW GLOBAL STATUS LIKE '"+name+"%'")
	if err != nil {
		mnd.Apps.Add("MySQL&&Errors", 1)
		return fmt.Errorf("getting global status: %w", err)
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		mnd.Apps.Add("MySQL&&Errors", 1)
		return fmt.Errorf("getting global status rows: %w", err)
	}

	for rows.Next() {
		var vname, value string

		if err := rows.Scan(&vname, &value); err != nil {
			mnd.Apps.Add("MySQL&&Errors", 1)
			return fmt.Errorf("scanning global status rows: %w", err)
		}

		v, err := strconv.ParseFloat(value, mnd.Bits64)
		if err != nil || v == 0 {
			continue
		}

		m[vname] = v
	}

	return nil
}

// Len allows us to sort MySQLProcesses.
func (s MySQLProcesses) Len() int {
	return len(s)
}

// Swap allows us to sort MySQLProcesses.
func (s MySQLProcesses) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less allows us to sort MySQLProcesses.
func (s MySQLProcesses) Less(i, j int) bool {
	return s[i].Time > s[j].Time
}

// Shrink a process list.
func (s *MySQLProcesses) Shrink(size int) {
	if size == 0 {
		size = defaultMyLimit
	}

	if s == nil {
		return
	}

	if len(*s) > size {
		*s = (*s)[:size]
	}
}
