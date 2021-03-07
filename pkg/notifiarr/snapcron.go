package notifiarr

import (
	"strings"
	"time"
)

func (c *Config) startSnapCron() {
	if c.Snap.Interval.Duration == 0 || c.stopSnap != nil {
		return
	}

	t := time.NewTicker(c.Snap.Interval.Duration)
	c.stopSnap = make(chan struct{})
	c.logStart()

	defer func() {
		t.Stop()
		close(c.stopSnap)
		c.stopSnap = nil
	}()

	for {
		select {
		case <-t.C:
			c.sendSnapshot()
		case <-c.stopSnap:
			return
		}
	}
}

func (c *Config) logStart() {
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

	for _, err := range debug {
		if err != nil {
			c.Debugf("Snapshot: %v", err)
		}
	}

	if body, err := c.SendData(c.URL, &Payload{Snap: snapshot}); err != nil {
		c.Errorf("Sending snapshot to %s: %v: %v", c.URL, err, string(body))
	} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
		c.Printf("Systems Snapshot sent to %s, sending again in %s, reply: %s", c.URL, c.Snap.Interval, fields[3])
	} else {
		c.Printf("Systems Snapshot sent to %s, sending again in %s, reply: %s", c.URL, c.Snap.Interval, string(body))
	}
}
