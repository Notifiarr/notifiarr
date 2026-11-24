package website //nolint:testpackage // payload host uses unexported server fields.

import (
	"testing"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/stretchr/testify/require"
)

func TestCleanHostname(t *testing.T) { //nolint:paralleltest // mutates mnd.IsDocker
	original := mnd.IsDocker
	t.Cleanup(func() { mnd.IsDocker = original })

	require.Empty(t, CleanHostname(nil))

	info := &host.InfoStat{
		Hostname:      "keep-me-please-extra-junk",
		KernelVersion: "6.12.15-truenas",
	}

	mnd.IsDocker = false
	require.Equal(t, info.Hostname, CleanHostname(info))

	mnd.IsDocker = true
	require.Equal(t, "keep-me-please", CleanHostname(info))
	require.Equal(t, "short-name", CleanHostname(&host.InfoStat{
		Hostname:      "short-name",
		KernelVersion: "6.12.15-truenas",
	}))
	require.Equal(t, info.Hostname, CleanHostname(&host.InfoStat{
		Hostname:      info.Hostname,
		KernelVersion: "6.12.15-generic",
	}))
}

func TestClientHostname(t *testing.T) {
	t.Parallel()

	require.Equal(t, "box", (&server{}).clientHostname("box"))
	require.Equal(t, "box", (&server{config: &Config{}}).clientHostname("box"))
	require.Equal(t, "box", (&server{config: &Config{ClientName: "  "}}).clientHostname("box"))
	require.Equal(t, "library", (&server{config: &Config{ClientName: " library "}}).clientHostname("box"))
}

func TestPayloadHostUsesClientName(t *testing.T) { //nolint:paralleltest // mutates the website client
	original := site
	t.Cleanup(func() { site = original })

	site = &server{
		config:   &Config{ClientName: "library"},
		hostInfo: &host.InfoStat{Hostname: "random-container", HostID: "abc"},
	}

	info, err := GetHostInfo(t.Context())
	require.NoError(t, err)
	require.Equal(t, "random-container", info.Hostname)

	sent := site.payloadHost(t.Context())
	require.Equal(t, "library", sent.Hostname)
	require.Equal(t, "abc", sent.HostID)
	require.Equal(t, "random-container", site.hostInfo.Hostname)
	require.Equal(t, "random-container", info.Hostname)

	site.config.ClientName = ""

	sent = site.payloadHost(t.Context())
	require.Equal(t, "random-container", sent.Hostname)
}
