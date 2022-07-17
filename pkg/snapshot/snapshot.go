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
	"errors"
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
	Timeout   cnfg.Duration `toml:"timeout" xml:"timeout" json:"timeout"`                     // total run time allowed.
	Interval  cnfg.Duration `toml:"interval" xml:"interval" json:"interval"`                  // how often to send snaps (cron).
	ZFSPools  []string      `toml:"zfs_pools" xml:"zfs_pool" json:"zfsPools"`                 // zfs pools to monitor.
	UseSudo   bool          `toml:"use_sudo" xml:"use_sudo" json:"useSudo"`                   // use sudo for smartctl commands.
	Raid      bool          `toml:"monitor_raid" xml:"monitor_raid" json:"monitorRaid"`       // include mdstat and/or megaraid.
	DriveData bool          `toml:"monitor_drives" xml:"monitor_drives" json:"monitorDrives"` // smartctl commands.
	DiskUsage bool          `toml:"monitor_space" xml:"monitor_space" json:"monitorSpace"`    // get disk usage.
	AllDrives bool          `toml:"all_drives" xml:"all_drives" json:"allDrives"`             // usage for all drives?
	IOTop     int           `toml:"iotop" xml:"iotop" json:"ioTop"`                           // number of processes to include from ioTop
	PSTop     int           `toml:"pstop" xml:"pstop" json:"psTop"`                           // number of processes to include from top (cpu usage)
	MyTop     int           `toml:"mytop" xml:"mytop" json:"myTop"`                           // number of processes to include from mysql servers.
	*Plugins
	// Debug     bool          `toml:"debug" xml:"debug" json:"debug"`
}

// Plugins is optional configuration for "plugins".
type Plugins struct {
	Nvidia *NvidiaConfig  `toml:"nvidia" xml:"nvidia" json:"nvidia"`
	MySQL  []*MySQLConfig `toml:"mysql" xml:"mysql" json:"mysql"`
}

// Errors this package generates.
var (
	ErrPlatformUnsup = fmt.Errorf("the requested metric is not available on this platform, " +
		"if you know how to collect it, please open an issue on the github repo")
	ErrNonZeroExit = fmt.Errorf("cmd exited non-zero")
)

// Snapshot is the output data sent to Notifiarr.
type Snapshot struct {
	Version string `json:"version"`
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
	Nvidia     []*NvidiaOutput                `json:"nvidia,omitempty"`
}

// RaidData contains raid information from mdstat and/or megacli.
type RaidData struct {
	MDstat  string     `json:"mdstat,omitempty"`
	MegaCLI []*MegaCLI `json:"megacli,omitempty"`
}

// Partition is used for ZFS pools as well as normal Disk arrays.
type Partition struct {
	Device string `json:"name"`
	Total  uint64 `json:"total"`
	Free   uint64 `json:"free"`
	Used   uint64 `json:"used"`
}

// Validate makes sure the snapshot configuration is valid.
func (c *Config) Validate() {
	switch {
	case c.Timeout.Duration == 0:
		c.Timeout.Duration = DefaultTimeout
	case c.Timeout.Duration < minimumTimeout:
		c.Timeout.Duration = minimumTimeout
	case c.Timeout.Duration > maximumTimeout:
		c.Timeout.Duration = maximumTimeout
	}

	if c.Interval.Duration != 0 && c.Interval.Duration < minimumInterval {
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

	snap := &Snapshot{Version: version.Version + "-" + version.Revision}
	errs, debug := c.getSnapshot(ctx, snap)

	return snap, errs, debug
}

func (c *Config) getSnapshot(ctx context.Context, snap *Snapshot) ([]error, []error) {
	errs := snap.GetProcesses(ctx, c.PSTop)
	errs = append(errs, snap.GetCPUSample(ctx))

	if err := snap.GetLocalData(ctx); len(err) != 0 {
		errs = append(errs, err...)
	}

	if syn, err := GetSynology(); err != nil && !errors.Is(err, ErrNotSynology) {
		errs = append(errs, err)
	} else if syn != nil {
		syn.SetInfo(snap.System.InfoStat)
	}

	if err := snap.getDisksUsage(ctx, c.DiskUsage, c.AllDrives); len(err) != 0 {
		errs = append(errs, err...)
	}

	var debug []error

	if err := snap.getDriveData(ctx, c.DriveData, c.UseSudo); len(err) != 0 {
		debug = append(debug, err...) // these can be noisy, so debug/hide them.
	}

	if err := snap.GetMySQL(ctx, c.Plugins.MySQL, c.MyTop); len(err) != 0 {
		errs = append(errs, err...)
	}

	errs = append(errs, snap.GetMemoryUsage(ctx))
	errs = append(errs, snap.getZFSPoolData(ctx, c.ZFSPools))
	errs = append(errs, snap.getRaidData(ctx, c.UseSudo, c.Raid))
	errs = append(errs, snap.getSystemTemps(ctx))
	errs = append(errs, snap.getIOTop(ctx, c.UseSudo, c.IOTop))
	errs = append(errs, snap.getIoStat(ctx, c.DiskUsage && mnd.IsLinux))
	errs = append(errs, snap.getIoStat2(ctx, c.DiskUsage))
	errs = append(errs, snap.GetNvidia(ctx, c.Nvidia))

	return errs, debug
}

/*******************************************************/
/*********************** HELPERS ***********************/
/*******************************************************/

// readyCommand gets a command ready for output capture.
func readyCommand(
	ctx context.Context,
	useSudo bool,
	run string,
	args ...string,
) (*exec.Cmd, *bufio.Scanner, *sync.WaitGroup, error) {
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
func runCommand(cmd *exec.Cmd, waitg *sync.WaitGroup) error {
	waitg.Add(1)

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	err := cmd.Run() //nolint:ifshort

	waitg.Wait()

	if err != nil {
		return fmt.Errorf("%v %w: %s", cmd.Args, err, stderr)
	}

	return nil
}
