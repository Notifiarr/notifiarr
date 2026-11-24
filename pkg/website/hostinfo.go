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
func GetHostInfo(ctx context.Context) (*host.InfoStat, error) {
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

	// TrueNAS adds junk to the hostname.
	hostInfo.Hostname = CleanHostname(hostInfo)

	if hid := site.config.HostID; hid != "" {
		hostInfo.HostID = hid
	}

	// This only happens once.
	site.hostInfo = hostInfo

	return site.hostInfoNoError(), nil // return a copy.
}

const (
	trueNasJunkLen   = 17
	trueNasJunkParts = 2
)

// CleanHostname returns the system hostname reported when no client name is configured.
// TrueNAS Docker hosts append a random suffix; that suffix is removed.
func CleanHostname(info *host.InfoStat) string {
	if info == nil {
		return ""
	}

	hostname := info.Hostname
	if mnd.IsDocker && strings.HasSuffix(info.KernelVersion, "truenas") && len(hostname) > trueNasJunkLen {
		if splitHost := strings.Split(hostname, "-"); len(splitHost) > trueNasJunkParts {
			return strings.Join(splitHost[:len(splitHost)-trueNasJunkParts], "-")
		}
	}

	return hostname
}

// payloadHost is the host object included in payloads sent to Notifiarr.com.
// ClientName replaces the system hostname so a container recreate does not rename the client.
func (s *server) payloadHost(ctx context.Context) *host.InfoStat {
	info, err := GetHostInfo(ctx)
	if err != nil {
		mnd.Log.Errorf(mnd.GetID(ctx), "Host Info Unknown: %v", err)
		return info
	}

	if info != nil {
		info.Hostname = s.clientHostname(info.Hostname)
	}

	return info
}

func (s *server) clientHostname(system string) string {
	if s == nil || s.config == nil {
		return system
	}

	if name := strings.TrimSpace(s.config.ClientName); name != "" {
		return name
	}

	return system
}
