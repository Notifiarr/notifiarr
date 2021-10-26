package snapshot

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	// We use mysql driver, this is how it's loaded.
	_ "github.com/go-sql-driver/mysql"
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

// MarshalJSON makes the output from sql.NullString not suck.
func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte(`"NULL"`), nil
	}

	return []byte(`"` + n.String + `"`), nil
}

// GetMySQL grabs the process list from a bunch of servers.
func (s *Snapshot) GetMySQL(ctx context.Context, servers []*MySQLConfig, limit int) (errs []error) {
	s.MySQL = make(map[string]MySQLProcesses)

	for _, server := range servers {
		if server.Host == "" {
			continue
		}

		data, err := getMySQL(ctx, server)
		if err != nil {
			errs = append(errs, err)
		}

		if server.Name != "" {
			s.MySQL[server.Name] = data
		} else {
			s.MySQL[server.Host] = data
		}
	}

	for _, v := range s.MySQL {
		sort.Sort(v)
		v.Shrink(limit)
	}

	return errs
}

func getMySQL(ctx context.Context, s *MySQLConfig) (MySQLProcesses, error) {
	id := s.Host
	if s.Name != "" {
		id = s.Name
	}

	host := "@tcp(" + s.Host + ")"
	if strings.HasPrefix(s.Host, "@") {
		host = s.Host
	}

	db, err := sql.Open("mysql", s.User+":"+s.Pass+host+"/")
	if err != nil {
		return nil, fmt.Errorf("mysql server %s: connecting: %w", id, err)
	}
	defer db.Close()

	list, err := scanMySQLProcessList(ctx, db)
	if err != nil {
		return list, fmt.Errorf("mysql server %s: %w", id, err)
	}

	return list, nil
}

func scanMySQLProcessList(ctx context.Context, db *sql.DB) (MySQLProcesses, error) {
	rows, err := db.QueryContext(ctx, "SHOW PROCESSLIST")
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
			return nil, fmt.Errorf("scanning rows: %w", err)
		}

		list = append(list, &p)
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
