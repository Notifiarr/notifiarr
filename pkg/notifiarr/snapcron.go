package notifiarr

import "fmt"

func (c *Config) logSnapshotStartup() {
	var ex string

	for key, val := range map[string]bool{
		"cpu, load, memory, uptime, users, temps": true,
		"raid":   c.Snap.Raid,
		"disks":  c.Snap.DiskUsage,
		"drives": c.Snap.DriveData,
		"iotop":  c.Snap.IOTop > 0,
		"pstop":  c.Snap.PSTop > 0,
		"mysql":  c.Snap.Plugins != nil && len(c.Snap.MySQL) > 0,
		"zfs":    len(c.Snap.ZFSPools) > 0,
		"sudo":   c.Snap.UseSudo && c.Snap.DriveData,
	} {
		if !val {
			continue
		}

		if ex != "" {
			ex += ", "
		}

		ex += key
	}

	if c.Snap.Interval.Duration == 0 {
		c.Printf("==> System Snapshot Collection Disabled, timeout: %v, configured: %s", c.Snap.Timeout, ex)
		return
	}

	c.Printf("==> System Snapshot Collection Started, interval: %v, timeout: %v, enabled: %s",
		c.Snap.Interval, c.Snap.Timeout, ex)
}

func (t *Triggers) SendSnapshot(event EventType) {
	t.exec(event, TrigSnapshot)
}

func (c *Config) sendSnapshot(event EventType) {
	snapshot, errs, debug := c.Snap.GetSnapshot()
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

	c.QueueData(&SendRequest{
		Route:      SnapRoute,
		Event:      event,
		LogPayload: true,
		LogMsg:     fmt.Sprintf("System Snapshot (interval: %v)", c.Snap.Interval),
		Payload:    &Payload{Snap: snapshot},
	})
}
