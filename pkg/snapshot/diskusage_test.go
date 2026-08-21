package snapshot //nolint:testpackage

import (
	"errors"
	"strings"
	"testing"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errUsageMissing = errors.New("usage missing")

func TestCollectDiskUsageKeepsReadOnlyNFSBinds(t *testing.T) {
	t.Parallel()

	// Docker `:ro` bind mounts of NFS (and systemd automounts) commonly have
	// both `ro` and `nodev`. `nodev` is a security flag, not "no device".
	got := collectDiskUsage([]disk.PartitionStat{
		{
			Device: "/dev/mapper/ubuntu--vg-ubuntu--lv", Mountpoint: "/", Fstype: "ext4",
			Opts: []string{"rw", "relatime"},
		},
		{
			Device: "192.168.50.206:/mnt/array1/Movies", Mountpoint: "/buffalo01", Fstype: "nfs4",
			Opts: []string{"ro", "relatime", "nosuid", "nodev"},
		},
		{
			Device: "192.168.50.205:/nfs/server_mnt", Mountpoint: "/lenovo01", Fstype: "nfs4",
			Opts: []string{"ro", "relatime", "nosuid", "nodev"},
		},
	}, usageMap{
		"/":          {Fstype: "ext4", Total: 913_000_000_000, Free: 841_000_000_000, Used: 34_000_000_000},
		"/buffalo01": {Fstype: "nfs4", Total: 8_100_000_000_000, Free: 2_600_000_000_000, Used: 5_500_000_000_000},
		"/lenovo01":  {Fstype: "nfs4", Total: 8_000_000_000_000, Free: 8_000_000_000_000, Used: 4_900_000_000},
	}.get)

	require.Len(t, got, 3)
	assert.Equal(t, "/", got["/dev/mapper/ubuntu--vg-ubuntu--lv"].Device)
	assert.Equal(t, "/buffalo01", got["/buffalo01"].Device)
	assert.Equal(t, "192.168.50.206:/mnt/array1/Movies", got["/buffalo01"].DevicePath)
	assert.Equal(t, "/lenovo01", got["/lenovo01"].Device)
	assert.True(t, got["/buffalo01"].ReadOnly)
	assert.True(t, got["/lenovo01"].ReadOnly)
}

func TestCollectDiskUsageDropsSystemdAutofsDuplicates(t *testing.T) {
	t.Parallel()

	got := collectDiskUsage([]disk.PartitionStat{
		{
			Device: "systemd-1", Mountpoint: "/mnt/buffalo01", Fstype: "autofs",
			Opts: []string{"rw", "relatime"},
		},
		{
			Device: "192.168.50.206:/mnt/array1/Movies", Mountpoint: "/mnt/buffalo01", Fstype: "nfs4",
			Opts: []string{"rw", "relatime"},
		},
		{
			Device: "systemd-1", Mountpoint: "/mnt/lenovo01", Fstype: "autofs",
			Opts: []string{"rw", "relatime"},
		},
		{
			Device: "192.168.50.205:/nfs/server_mnt", Mountpoint: "/mnt/lenovo01", Fstype: "nfs4",
			Opts: []string{"rw", "relatime"},
		},
	}, usageMap{
		"/mnt/buffalo01": {Fstype: "nfs4", Total: 8_100_000_000_000, Free: 2_600_000_000_000, Used: 5_500_000_000_000},
		"/mnt/lenovo01":  {Fstype: "nfs4", Total: 8_000_000_000_000, Free: 8_000_000_000_000, Used: 4_900_000_000},
	}.get)

	require.Len(t, got, 2)
	assert.Equal(t, "192.168.50.206:/mnt/array1/Movies", got["/mnt/buffalo01"].DevicePath)
	assert.Equal(t, "192.168.50.205:/nfs/server_mnt", got["/mnt/lenovo01"].DevicePath)

	for _, part := range got {
		assert.NotEqual(t, "autofs", part.FSType)
		assert.False(t, strings.HasPrefix(part.DevicePath, "systemd-"))
	}
}

func TestCollectDiskUsageSkipsTmpfsEvenWhenUsageLooksLikeOverlay(t *testing.T) {
	t.Parallel()

	got := collectDiskUsage([]disk.PartitionStat{
		{
			Device: "/dev/mapper/ubuntu--vg-ubuntu--lv", Mountpoint: "/", Fstype: "ext4",
			Opts: []string{"rw", "relatime"},
		},
		{
			Device: "tmpfs", Mountpoint: "/tmp", Fstype: "tmpfs",
			Opts: []string{"rw", "nosuid", "nodev"},
		},
	}, usageMap{
		"/":    {Fstype: "ext4", Total: 913_000_000_000, Free: 841_000_000_000, Used: 34_000_000_000},
		"/tmp": {Fstype: "overlay", Total: 913_000_000_000, Free: 841_000_000_000, Used: 34_000_000_000},
	}.get)

	require.Len(t, got, 1)
	assert.Equal(t, "/", got["/dev/mapper/ubuntu--vg-ubuntu--lv"].Device)
}

func TestCollectDiskUsagePrefersVolumeMountOverDockerMetadata(t *testing.T) {
	t.Parallel()

	mapper := "/dev/mapper/ubuntu--vg-ubuntu--lv"
	got := collectDiskUsage([]disk.PartitionStat{
		{
			Device: mapper, Mountpoint: "/etc/hosts", Fstype: "ext4",
			Opts: []string{"rw", "relatime", "bind"},
		},
		{
			Device: mapper, Mountpoint: "/config", Fstype: "ext4",
			Opts: []string{"rw", "relatime", "bind"},
		},
		{
			Device: mapper, Mountpoint: "/logs", Fstype: "ext4",
			Opts: []string{"rw", "relatime", "bind"},
		},
	}, usageMap{
		"/etc/hosts": {Fstype: "ext4", Total: 913_000_000_000, Free: 841_000_000_000, Used: 34_000_000_000},
		"/config":    {Fstype: "ext4", Total: 913_000_000_000, Free: 841_000_000_000, Used: 34_000_000_000},
		"/logs":      {Fstype: "ext4", Total: 913_000_000_000, Free: 841_000_000_000, Used: 34_000_000_000},
	}.get)

	require.Len(t, got, 1)
	assert.Equal(t, "/logs", got[mapper].Device)
	assert.Equal(t, mapper, got[mapper].DevicePath)
}

func TestIsNetworkDevice(t *testing.T) {
	t.Parallel()

	assert.True(t, isNetworkDevice("192.168.50.205:/nfs/server_mnt"))
	assert.True(t, isNetworkDevice("//nas.lan/share"))
	assert.False(t, isNetworkDevice("/dev/sda1"))
	assert.False(t, isNetworkDevice("overlay"))
	assert.False(t, isNetworkDevice("systemd-1"))
}

func TestIsJunkMountKeepsRemovableMedia(t *testing.T) {
	t.Parallel()

	assert.False(t, isJunkMount("/run/media/user/USB"))
	assert.True(t, isJunkMount("/run/docker.sock"))
	assert.True(t, isJunkMount("/etc/hosts"))
}

func TestSkipPartitionDoesNotTreatNodevAsNoDevice(t *testing.T) {
	t.Parallel()

	part := disk.PartitionStat{
		Device:     "192.168.50.205:/nfs/server_mnt",
		Mountpoint: "/lenovo01",
		Fstype:     "nfs4",
		Opts:       []string{"ro", "nodev", "nosuid"},
	}
	assert.False(t, skipPartition(&part))
}

type usageMap map[string]*disk.UsageStat

func (u usageMap) get(mount string) (*disk.UsageStat, error) {
	stat, ok := u[mount]
	if !ok {
		return nil, errUsageMissing
	}

	return stat, nil
}
