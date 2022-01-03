package notifiarr

func (c *Config) logSnapshotStartup() {
	if c.Snap.Interval.Duration < 1 {
		return
	}

	var ex string

	for k, v := range map[string]bool{
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

	resp, err := c.SendData(SnapRoute.Path(event), &Payload{Snap: snapshot}, true)
	if err != nil {
		c.Errorf("[%s requested] Sending snapshot to Notifiarr: %v", event, err)
	} else {
		c.Printf("[%s requested] System Snapshot sent to Notifiarr, cron interval: %s. %s", event, c.Snap.Interval, resp)
	}
}
