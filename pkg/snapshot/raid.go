package snapshot

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func (s *Snapshot) getRaidData(ctx context.Context, useSudo, run bool) error {
	if !run {
		return nil
	}

	s.Raid = &RaidData{}
	s.getRaidMDstat()

	return s.getRaidMegaCLI(ctx, useSudo)
}

// getRaidMDstat parses this:
/* four drive raid1:
$ cat /proc/mdstat
Personalities : [raid1] [linear] [multipath] [raid0] [raid6] [raid5] [raid4] [raid10]
md1 : active raid1 sdd2[3] sdb2[1] sdc2[2] sda2[0]
      536738816 blocks super 1.2 [4/4] [UUUU]
      bitmap: 3/4 pages [12KB], 65536KB chunk
unused devices: <none>.
*/
func (s *Snapshot) getRaidMDstat() {
	data, _ := os.ReadFile("/proc/mdstat")
	// Remove the first line "Personalities".
	if i := bytes.IndexByte(data, '\n'); i != -1 && len(data) > i+1 {
		data = data[i+1:]
	}

	s.Raid.MDstat = string(data)
}

/* getRaidMegaCLI parses this:
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
Is VD Cached: No.
*/

// MegaCLI represents the megaraid cli output.
type MegaCLI struct {
	Drive   string            `json:"drive"`
	Target  string            `json:"target"`
	Adapter string            `json:"adapter"`
	Data    map[string]string `json:"data"`
}

func (s *Snapshot) getRaidMegaCLI(ctx context.Context, useSudo bool) error {
	megacli, err := exec.LookPath("MegaCli64")
	for _, s := range []string{"MegaCli", "megacli", "megacli64"} {
		if err == nil {
			break
		}

		megacli, err = exec.LookPath(s)
	}

	if err != nil {
		// we dont return an error if megacli does not exist.
		return nil //nolint:nilerr
	}

	cmd, stdout, waitg, err := readyCommand(ctx, useSudo, megacli, "-LDInfo", "-Lall", "-aALL")
	if err != nil {
		return err
	}

	go s.scanMegaCLI(stdout, waitg)

	return runCommand(cmd, waitg)
}

func (s *Snapshot) scanMegaCLI(stdout *bufio.Scanner, waitg *sync.WaitGroup) {
	var (
		adapter string
		current *MegaCLI
	)

	for stdout.Scan() {
		text := stdout.Text()
		if strings.HasPrefix(text, "Adapter ") {
			if split := strings.Fields(text); len(split) > 1 {
				adapter = split[1]
			}

			continue
		}

		if strings.HasPrefix(text, "Virtual Drive:") {
			if current != nil {
				s.Raid.MegaCLI = append(s.Raid.MegaCLI, current)
			}

			split := strings.Fields(text)
			current = &MegaCLI{
				Drive:   split[2],
				Target:  strings.TrimRight(split[5], ")"),
				Adapter: adapter,
				Data:    make(map[string]string),
			}

			continue
		}

		if split := strings.SplitN(strings.TrimSpace(text), ":", 2); len(split) == 2 && current != nil { //nolint:mnd
			current.Data[strings.TrimSpace(split[0])] = strings.TrimSpace(split[1])
		}
	}

	if current != nil {
		// Append the final (last) VD.
		s.Raid.MegaCLI = append(s.Raid.MegaCLI, current)
	}

	waitg.Done()
}
