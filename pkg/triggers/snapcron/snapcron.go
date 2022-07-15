package snapcron

import (
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const TrigSnapshot common.TriggerName = "Gathering and sending System Snapshot."

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
}

// New configures the library.
func New(config *common.Config) *Action {
	return &Action{cmd: &cmd{Config: config}}
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.create()
}

// Send a snapshot to the website.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(event, TrigSnapshot)
}

func (c *cmd) create() {
	var ticker *time.Ticker

	if c.Snapshot.Interval.Duration > 0 {
		ticker = time.NewTicker(c.Snapshot.Interval.Duration)
	}

	c.printLog()
	c.Add(&common.Action{
		Name: TrigSnapshot,
		Fn:   c.sendSnapshot,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})
}

func (c *cmd) printLog() {
	var ex string

	for key, val := range map[string]bool{
		"cpu, load, memory, uptime, users, temps": true,
		"raid":   c.Snapshot.Raid,
		"disks":  c.Snapshot.DiskUsage,
		"drives": c.Snapshot.DriveData,
		"iotop":  c.Snapshot.IOTop > 0,
		"pstop":  c.Snapshot.PSTop > 0,
		"mysql":  c.Snapshot.Plugins != nil && len(c.Snapshot.MySQL) > 0,
		"zfs":    len(c.Snapshot.ZFSPools) > 0,
		"sudo":   c.Snapshot.UseSudo && c.Snapshot.DriveData,
	} {
		if !val {
			continue
		}

		if ex != "" {
			ex += ", "
		}

		ex += key
	}

	if c.Snapshot.Interval.Duration == 0 {
		c.Printf("==> System Snapshot Collection Disabled, timeout: %v, configured: %s", c.Snapshot.Timeout, ex)
		return
	}

	c.Printf("==> System Snapshot Collection Started, interval: %v, timeout: %v, enabled: %s",
		c.Snapshot.Interval, c.Snapshot.Timeout, ex)
}

func (c *cmd) sendSnapshot(event website.EventType) {
	snapshot, errs, debug := c.Snapshot.GetSnapshot()
	for _, err := range errs {
		if err != nil {
			c.ErrorfNoShare("[%s requested] Snapshot: %v", event, err)
		}
	}

	// These debug messages are mostly just errors that we we expect to have.
	for _, err := range debug {
		if err != nil {
			c.Debugf("Snapshot: %v", err)
		}
	}

	c.SendData(&website.Request{
		Route:      website.SnapRoute,
		Event:      event,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("System Snapshot (interval: %v)", c.Snapshot.Interval),
		Payload:    &website.Payload{Snap: snapshot},
	})
}
