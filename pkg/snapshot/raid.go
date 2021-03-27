package snapshot

import (
	"bytes"
	"context"
	"io/ioutil"
	"os/exec"
	"strings"
)

func (s *Snapshot) getRaidData(ctx context.Context, useSudo, run bool) error {
	if !run {
		return nil
	}

	s.Raid = &RaidData{}
	s.getRaidMDstat()

	return s.getRaidMegaCLI(ctx, useSudo)
}

/* four drive raid1:
$ cat /proc/mdstat
Personalities : [raid1] [linear] [multipath] [raid0] [raid6] [raid5] [raid4] [raid10]
md1 : active raid1 sdd2[3] sdb2[1] sdc2[2] sda2[0]
      536738816 blocks super 1.2 [4/4] [UUUU]
      bitmap: 3/4 pages [12KB], 65536KB chunk
unused devices: <none>
*/

func (s *Snapshot) getRaidMDstat() {
	b, _ := ioutil.ReadFile("/proc/mdstat")
	// Remove the first line "Personalities" and replace the rest of the newlines with spaces.
	if i := bytes.IndexByte(b, '\n'); i != -1 && len(b) > i+1 {
		b = b[i+1:]
	}

	s.Raid.MDstat = string(b)
}

/*
[root@server]# MegaCli -LDInfo -Lall -aALL
Adapter 0 -- Virtual Drive Information:
Virtual Drive: 0 (Target Id: 0)
Name                :
RAID Level          : Primary-5, Secondary-0, RAID Level Qualifier-3
Size                : 50.937 TB
Sector Size         : 512
Is VD emulated      : Yes
Parity Size         : 7.276 TB
State               : Degraded
Strip Size          : 128 KB
Number Of Drives    : 8
Span Depth          : 1
Default Cache Policy: WriteBack, ReadAhead, Cached, Write Cache OK if Bad BBU
Current Cache Policy: WriteBack, ReadAhead, Cached, Write Cache OK if Bad BBU
Default Access Policy: Read/Write
Current Access Policy: Read/Write
Disk Cache Policy   : Enabled
Encryption Type     : None
Is VD Cached: No
*/

func (s *Snapshot) getRaidMegaCLI(ctx context.Context, useSudo bool) error {
	// The megacli code is barely tested.
	megacli, err := exec.LookPath("MegaCli")
	for _, s := range []string{"MegaCli64", "megacli", "megacli64"} {
		if err == nil {
			break
		}

		megacli, err = exec.LookPath(s)
	}

	if err != nil {
		// we dont return an error if megacli does not exist.
		return nil //nolint:nilerr
	}

	cmd, stdout, wg, err := readyCommand(ctx, useSudo, megacli, "-LDInfo", "-Lall", "-aALL")
	if err != nil {
		return err
	}

	s.Raid.MegaCLI = make(map[string]string)

	go func() {
		for stdout.Scan() {
			if split := strings.Split(strings.TrimSpace(stdout.Text()), ":"); len(split) == 2 { // nolint:gomnd
				s.Raid.MegaCLI[strings.TrimSpace(split[0])] = strings.TrimSpace(split[1])
			}
		}
		wg.Done()
	}()

	return runCommand(cmd, wg)
}
