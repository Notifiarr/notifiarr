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
	_ "github.com/go-sql-driver/mysql" //nolint:gci // We use mysql driver, this is how it's loaded.
	"golift.io/cnfg"
)

// MySQLConfig allows us to gather a process list for the snapshot.
type MySQLConfig struct {
	Name     string        `toml:"name"`
	Host     string        `toml:"host"`
	User     string        `toml:"user"`
	Pass     string        `toml:"pass"`
	Timeout  cnfg.Duration `toml:"timeout"`  // only used by service checks, snapshot timeout is used for mysql.
	Interval cnfg.Duration `toml:"interval"` // only used by service checks, snapshot interval is used for mysql.
}

// MySQLProcesses allows us to manipulate our list with methods.
type MySQLProcesses []*MySQLProcess

// MySQLProcess represents the data returned from SHOW PROCESS LIST.
type MySQLProcess struct {
	ID    int64      `json:"id"`
	User  string     `json:"user"`
	Host  string     `json:"host"`
	DB    NullString `json:"db"`
	Cmd   string     `json:"command"`
	Time  int64      `json:"time"`
	State string     `json:"state"`
	Info  NullString `json:"info"`
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
func (s *Snapshot) GetMySQL(ctx context.Context, servers []*MySQLConfig, limit int) (errs []error) {
	s.MySQL = make(map[string]*MySQLServerData)

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
	rows, err := dbase.QueryContext(ctx, "SHOW FULL PROCESSLIST")
	if err != nil {
		return nil, fmt.Errorf("getting processes: %w", err)
	} else if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("getting processes rows: %w", err)
	}
	defer rows.Close()

	var list MySQLProcesses

	for rows.Next() {
		var p MySQLProcess

		// for each row, scan the result into our tag composite object
		err := rows.Scan(&p.ID, &p.User, &p.Host, &p.DB, &p.Cmd, &p.Time, &p.State, &p.Info)
		if err != nil {
			return nil, fmt.Errorf("scanning process rows: %w", err)
		}

		if p.Info.Valid {
			p.Info.String = strings.Join(strings.Fields(p.Info.String), " ")
		}

		list = append(list, &p)
	}

	return list, nil
}

func scanMySQLStatus(ctx context.Context, dbase *sql.DB) (MySQLStatus, error) {
	var (
		list  = make(MySQLStatus)
		likes = []string{
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
		}
	)

	for _, name := range likes {
		rows, err := dbase.QueryContext(ctx, "SHOW GLOBAL STATUS LIKE '"+name+"%'")
		if err != nil {
			return nil, fmt.Errorf("getting global status: %w", err)
		} else if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("getting global status rows: %w", err)
		}

		for rows.Next() {
			var vname, value string

			if err := rows.Scan(&vname, &value); err != nil {
				return nil, fmt.Errorf("scanning global status rows: %w", err)
			}

			v, err := strconv.ParseFloat(value, mnd.Bits64)
			if err != nil || v == 0 {
				continue
			}

			list[vname] = v
		}

		rows.Close() //nolint:sqlclosecheck // not defered because we're in a loop.
	}

	return list, nil
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
