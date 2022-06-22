package snapcron

import (
	"fmt"

	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

const TrigSnapshot common.TriggerName = "Gathering and sending System Snapshot."

type Config struct {
	*common.Config
}

func (c *Config) Create() {
	c.printLog()
	c.Add(&common.Action{
		Name: TrigSnapshot,
		Fn:   c.sendSnapshot,
		C:    make(chan website.EventType, 1),
		T:    common.TickerOrNil(c.Snapshot.Interval.Duration),
	})
}

func (c *Config) printLog() {
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

func (c *Config) SendSnapshot(event website.EventType) {
	c.Exec(event, TrigSnapshot)
}

func (c *Config) sendSnapshot(event website.EventType) {
	snapshot, errs, debug := c.Snapshot.GetSnapshot()
	for _, err := range errs {
		if err != nil {
			c.Errorf("[%s requested] Snapshot: %v", event, err)
		}
	}

	// These debug messages are mostly just errors that we we expect to have.
	for _, err := range debug {
		if err != nil {
			c.Debugf("Snapshot: %v", err)
		}
	}

	c.QueueData(&website.SendRequest{
		Route:      website.SnapRoute,
		Event:      event,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("System Snapshot (interval: %v)", c.Snapshot.Interval),
		Payload:    &website.Payload{Snap: snapshot},
	})
}
