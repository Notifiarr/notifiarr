package website

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/shirou/gopsutil/v4/host"
)

// hostInfoNoError will return nil if there is an error, otherwise a copy of the host info.
func (s *server) hostInfoNoError() *host.InfoStat {
	if s.hostInfo == nil {
		return nil
	}

	return &host.InfoStat{
		Hostname:             s.hostInfo.Hostname,
		Uptime:               uint64(time.Now().Unix()) - s.hostInfo.BootTime,
		BootTime:             s.hostInfo.BootTime,
		OS:                   s.hostInfo.OS,
		Platform:             s.hostInfo.Platform,
		PlatformFamily:       s.hostInfo.PlatformFamily,
		PlatformVersion:      s.hostInfo.PlatformVersion,
		KernelVersion:        s.hostInfo.KernelVersion,
		KernelArch:           s.hostInfo.KernelArch,
		VirtualizationSystem: s.hostInfo.VirtualizationSystem,
		VirtualizationRole:   s.hostInfo.VirtualizationRole,
		HostID:               s.hostInfo.HostID,
	}
}

// GetHostInfo attempts to make a unique machine identifier...
func GetHostInfo(ctx context.Context) (*host.InfoStat, error) { //nolint:cyclop
	if site.hostInfo != nil {
		return site.hostInfoNoError(), nil
	}

	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: GetHostInfo")
	defer mnd.Log.Trace(reqID, "end: GetHostInfo")

	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting host info: %w", err)
	}

	syn, err := snapshot.GetSynology(ctx, false) // false makes it not do snapshot data, just get syno info.
	if err == nil {
		// This method writes synology data into hostInfo.
		syn.SetInfo(hostInfo)
	}

	if hostInfo.Platform == "" &&
		(hostInfo.VirtualizationSystem == "docker" || mnd.IsDocker) {
		hostInfo.Platform = "Docker " + hostInfo.KernelVersion
		hostInfo.PlatformFamily = "Docker"
	}

	const (
		trueNasJunkLen   = 17
		trueNasJunkParts = 2
	)
	// TrueNAS adds junk to the hostname.
	if mnd.IsDocker && strings.HasSuffix(hostInfo.KernelVersion, "truenas") && len(hostInfo.Hostname) > trueNasJunkLen {
		if splitHost := strings.Split(hostInfo.Hostname, "-"); len(splitHost) > trueNasJunkParts {
			hostInfo.Hostname = strings.Join(splitHost[:len(splitHost)-trueNasJunkParts], "-")
		}
	}

	if hid := site.config.HostID; hid != "" {
		hostInfo.HostID = hid
	}

	// This only happens once.
	site.hostInfo = hostInfo

	return site.hostInfoNoError(), nil // return a copy.
}
