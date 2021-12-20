// Package snapshot generates system reports and sends them to notifiarr.com.
// The reports include zfs data, cpu, memory, mdadm info, megaraid arrays,
// smart status, mounted volume (disk) usage, cpu temp, other temps, uptime,
// drive age/health, logged on user count, etc. Works across most platforms.
// These snapshots are posted to a user's Chatroom on request.
package snapshot

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"golift.io/cnfg"
	"golift.io/version"
)

// DefaultTimeout is used when one is not provided.
const DefaultTimeout = 30 * time.Second

const (
	minimumTimeout  = 5 * time.Second
	maximumTimeout  = time.Minute
	minimumInterval = time.Minute
	defaultMyLimit  = 10
)

// Config determines which checks to run, etc.
//nolint:lll
type Config struct {
	Timeout   cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`                           // total run time allowed.
	Interval  cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`                        // how often to send snaps (cron).
	ZFSPools  []string      `toml:"zfs_pools" xml:"zfs_pool" json:"zfsPools"`                       // zfs pools to monitor.
	UseSudo   bool          `toml:"use_sudo" xml:"use_sudo" json:"useSudo"`                         // use sudo for smartctl commands.
	Raid      bool          `toml:"monitor_raid" xml:"monitor_raid" json:"monitorRaid"`             // include mdstat and/or megaraid.
	DriveData bool          `toml:"monitor_drives" xml:"monitor_drives" json:"monitorDrives"`       // smartctl commands.
	DiskUsage bool          `toml:"monitor_space" xml:"monitor_space" json:"monitorSpace"`          // get disk usage.
	AllDrives bool          `toml:"all_drives" xml:"all_drives" json:"allDrives"`                   // usage for all drives?
	Uptime    bool          `toml:"monitor_uptime" xml:"monitor_uptime" json:"monitorUptime"`       // all system stats.
	CPUMem    bool          `toml:"monitor_cpuMemory" xml:"monitor_cpuMemory" json:"monitorCpuMem"` // cpu perct and memory used/free.
	CPUTemp   bool          `toml:"monitor_cpuTemp" xml:"monitor_cpuTemp" json:"monitorCpuTemp"`    // not everything supports temps.
	IOTop     int           `toml:"iotop" xml:"iotop" json:"ioTop"`                                 // number of processes to include from ioTop
	PSTop     int           `toml:"pstop" xml:"pstop" json:"psTop"`                                 // number of processes to include from top (cpu usage)
	MyTop     int           `toml:"mytop" xml:"mytop" json:"myTop"`                                 // number of processes to include from mysql servers.
	*Plugins
	// Debug     bool          `toml:"debug" xml:"debug" json:"debug"`
}

// Plugins is optional configuration for "plugins".
type Plugins struct {
	MySQL []*MySQLConfig `toml:"mysql" xml:"mysql" json:"mysql"`
}

// Errors this package generates.
var (
	ErrPlatformUnsup = fmt.Errorf("the requested metric is not available on this platform, " +
		"if you know how to collect it, please open an issue on the github repo")
	ErrNonZeroExit = fmt.Errorf("cmd exited non-zero")
)

// Snapshot is the output data sent to Notifiarr.
type Snapshot struct {
	Version string
	Uptime  time.Duration
	System  struct {
		*host.InfoStat
		Username string             `json:"username"`
		CPU      float64            `json:"cpuPerc"`
		MemFree  uint64             `json:"memFree"`
		MemUsed  uint64             `json:"memUsed"`
		MemTotal uint64             `json:"memTotal"`
		Temps    map[string]float64 `json:"temperatures,omitempty"`
		Users    int                `json:"users"`
		*load.AvgStat
		CPUTime cpu.TimesStat `json:"cpuTime"`
	} `json:"system"`
	Raid       *RaidData                      `json:"raid,omitempty"`
	DriveAges  map[string]int                 `json:"driveAges,omitempty"`
	DriveTemps map[string]int                 `json:"driveTemps,omitempty"`
	DiskUsage  map[string]*Partition          `json:"diskUsage,omitempty"`
	DiskHealth map[string]string              `json:"driveHealth,omitempty"`
	IOTop      *IOTopData                     `json:"ioTop,omitempty"`
	IOStat     *IoStatDisks                   `json:"ioStat,omitempty"`
	IOStat2    map[string]disk.IOCountersStat `json:"ioStat2,omitempty"`
	Processes  Processes                      `json:"processes,omitempty"`
	ZFSPool    map[string]*Partition          `json:"zfsPools,omitempty"`
	MySQL      map[string]*MySQLServerData    `json:"mysql,omitempty"`
}

// RaidData contains raid information from mdstat and/or megacli.
type RaidData struct {
	MDstat  string            `json:"mdstat,omitempty"`
	MegaCLI map[string]string `json:"megacli,omitempty"`
}

// Partition is used for ZFS pools as well as normal Disk arrays.
type Partition struct {
	Device string `json:"name"`
	Total  uint64 `json:"total"`
	Free   uint64 `json:"free"`
	Used   uint64 `json:"used"`
}

// Validate makes sure the snapshot configuration is valid.
func (c *Config) Validate() { //nolint:cyclop
	switch {
	case c.Timeout.Duration == 0:
		c.Timeout.Duration = DefaultTimeout
	case c.Timeout.Duration < minimumTimeout:
		c.Timeout.Duration = minimumTimeout
	case c.Timeout.Duration > maximumTimeout:
		c.Timeout.Duration = maximumTimeout
	}

	if c.Interval.Duration == 0 {
		return
	} else if c.Interval.Duration < minimumInterval {
		c.Interval.Duration = minimumInterval
	}

	if mnd.IsDocker || mnd.IsWindows {
		c.UseSudo = false
	}

	if mnd.IsDocker || !mnd.IsLinux {
		c.IOTop = 0
	}
}

// GetSnapshot returns a system snapshot based on requested data in the config.
func (c *Config) GetSnapshot() (*Snapshot, []error, []error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout.Duration)
	defer cancel()

	s := &Snapshot{
		Version: version.Version + "-" + version.Revision,
		Uptime:  time.Since(version.Started),
	}
	errs, debug := c.getSnapshot(ctx, s)

	return s, errs, debug
}

func (c *Config) getSnapshot(ctx context.Context, s *Snapshot) ([]error, []error) {
	errs := s.GetProcesses(ctx, c.PSTop)
	errs = append(errs, s.GetCPUSample(ctx, c.CPUMem))

	if err := s.GetLocalData(ctx, c.Uptime); len(err) != 0 {
		errs = append(errs, err...)
	}

	if syn, err := GetSynology(c.Uptime); err != nil {
		errs = append(errs, err)
	} else if syn != nil {
		syn.SetInfo(s.System.InfoStat)
	}

	if err := s.getDisksUsage(ctx, c.DiskUsage, c.AllDrives); len(err) != 0 {
		errs = append(errs, err...)
	}

	var debug []error

	if err := s.getDriveData(ctx, c.DriveData, c.UseSudo); len(err) != 0 {
		debug = append(debug, err...) // these can be noisy, so debug/hide them.
	}

	if err := s.GetMySQL(ctx, c.Plugins.MySQL, c.MyTop); len(err) != 0 {
		errs = append(errs, err...)
	}

	errs = append(errs, s.GetMemoryUsage(ctx, c.CPUMem))
	errs = append(errs, s.getZFSPoolData(ctx, c.ZFSPools))
	errs = append(errs, s.getRaidData(ctx, c.UseSudo, c.Raid))
	errs = append(errs, s.getSystemTemps(ctx, c.CPUTemp))
	errs = append(errs, s.getIOTop(ctx, c.UseSudo, c.IOTop))
	errs = append(errs, s.getIoStat(ctx, c.DiskUsage && mnd.IsLinux))
	errs = append(errs, s.getIoStat2(ctx, c.DiskUsage))

	return errs, debug
}

/*******************************************************/
/*********************** HELPERS ***********************/
/*******************************************************/

// readyCommand gets a command ready for output capture.
func readyCommand(ctx context.Context, useSudo bool, run string, args ...string) (
	*exec.Cmd, *bufio.Scanner, *sync.WaitGroup, error) {
	cmdPath, err := exec.LookPath(run)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s missing! %w", run, err)
	}

	if args == nil { // avoid nil pointer deref.
		args = []string{}
	}

	if useSudo {
		args = append([]string{"-n", cmdPath}, args...)

		if cmdPath, err = exec.LookPath("sudo"); err != nil {
			return nil, nil, nil, fmt.Errorf("sudo missing! %w", err)
		}
	}

	cmd := exec.CommandContext(ctx, cmdPath, args...)
	sysCallSettings(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s stdout error: %w", cmdPath, err)
	}

	return cmd, bufio.NewScanner(stdout), &sync.WaitGroup{}, nil
}

// runCommand executes the readied command and waits for the output loop to finish.
func runCommand(cmd *exec.Cmd, wg *sync.WaitGroup) error {
	wg.Add(1)

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	err := cmd.Run() //nolint:ifshort

	wg.Wait()

	if err != nil {
		return fmt.Errorf("%v %w: %s", cmd.Args, err, stderr)
	}

	return nil
}
