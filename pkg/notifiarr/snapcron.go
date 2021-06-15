package notifiarr

import (
	"strings"
)

func (c *Config) logSnapshotStartup() {
	var ex string

	for k, v := range map[string]bool{
		"raid":    c.Snap.Raid,
		"disks":   c.Snap.DiskUsage,
		"drives":  c.Snap.DriveData,
		"uptime":  c.Snap.Uptime,
		"cpumem":  c.Snap.CPUMem,
		"cputemp": c.Snap.CPUTemp,
		"zfs":     c.Snap.ZFSPools != nil,
		"sudo":    c.Snap.UseSudo && c.Snap.DriveData,
	} {
		if !v {
			continue
		}

		if ex != "" {
			ex += ", "
		}

		ex += k
	}

	c.Printf("==> System Snapshot Collection Started, interval: %v, timeout: %v, enabled: %s",
		c.Snap.Interval, c.Snap.Timeout, ex)
}

func (c *Config) sendSnapshot() {
	snapshot, errs, debug := c.Snap.GetSnapshot()
	for _, err := range errs {
		if err != nil {
			c.Errorf("Snapshot: %v", err)
		}
	}

	// These debug messages are mostly just errors that we we expect to have.
	for _, err := range debug {
		if err != nil {
			c.Debugf("Snapshot: %v", err)
		}
	}

	if _, _, body, err := c.SendData(c.URL, &Payload{Type: SnapCron, Snap: snapshot}); err != nil {
		c.Errorf("Sending snapshot to %s: %v: %v", c.URL, err, string(body))
	} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
		c.Printf("Systems Snapshot sent to %s, sending again in %s, reply: %s", c.URL, c.Snap.Interval, fields[3])
	} else {
		c.Printf("Systems Snapshot sent to %s, sending again in %s, reply: %s", c.URL, c.Snap.Interval, string(body))
	}
}
