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
func (s *Server) hostInfoNoError() *host.InfoStat {
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
func (s *Server) GetHostInfo(ctx context.Context) (*host.InfoStat, error) { //nolint:cyclop
	if s.hostInfo != nil {
		return s.hostInfoNoError(), nil
	}

	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting host info: %w", err)
	}

	syn, err := snapshot.GetSynology(false) // false makes it not do snapshot data, just get syno info.
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

	if s.Config.HostID != "" {
		hostInfo.HostID = s.Config.HostID
	}

	// This only happens once.
	s.hostInfo = hostInfo

	return s.hostInfoNoError(), nil // return a copy.
}
