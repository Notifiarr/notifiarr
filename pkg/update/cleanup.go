package update

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
)

const keepBackups = 3

type file struct {
	name string
	when time.Time
}

type fileList []file

func (f fileList) Len() int {
	return len(f)
}

func (f fileList) Less(i, j int) bool {
	return f[i].when.UnixMicro() < f[j].when.UnixMicro()
}

func (f fileList) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (u *Command) cleanOldBackups() {
	dir := filepath.Dir(u.Path)
	pfx := strings.TrimSuffix(filepath.Base(u.Path), dotExe) + ".backup."

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	consider := fileList{}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), pfx) || len(entry.Name()) < len(pfx)+len(backupTimeFormat) {
			continue
		}

		date := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), pfx), dotExe)
		if when, err := time.Parse(backupTimeFormat, date); err == nil {
			consider = append(consider, file{name: entry.Name(), when: when})
		}
	}

	sort.Sort(consider)

	for idx, file := range consider {
		if len(consider)-idx <= keepBackups {
			return
		}

		err := os.Remove(filepath.Join(dir, file.name))
		u.Printf("[UPDATE] Deleted old backup file: %s%s (error: %v)", file.name, mnd.DurationAge(file.when), err)
	}
}
