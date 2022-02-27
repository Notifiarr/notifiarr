package notifiarr

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/update"
	"github.com/shirou/gopsutil/v3/host"
	"golift.io/cnfg"
	"golift.io/version"
)

// ClientInfo is the client's startup data received from the website.
type ClientInfo struct {
	User struct {
		WelcomeMSG string `json:"welcome"`
		Subscriber bool   `json:"subscriber"`
		Patron     bool   `json:"patron"`
	} `json:"user"`
	Actions struct {
		Poll      bool             `json:"poll"`
		Plex      *plex.Server     `json:"plex"`      // optional
		Apps      appConfigs       `json:"apps"`      // unused yet!
		Dashboard dashConfig       `json:"dashboard"` // now in use.
		Sync      syncConfig       `json:"sync"`      // in use (cfsync)
		Gaps      gapsConfig       `json:"gaps"`      // radarr collection gaps
		Custom    []*timerConfig   `json:"custom"`    // custom GET timers
		Snapshot  *snapshot.Config `json:"snapshot"`  // optional
	} `json:"actions"`
}

// ServiceConfig comes from the services package. It's only used for display on the website.
type ServiceConfig struct {
	Interval cnfg.Duration   `json:"interval"`
	Parallel uint            `json:"parallel"`
	Disabled bool            `json:"disabled"`
	Checks   []*ServiceCheck `json:"checks"`
}

// ServiceCheck comes from the services package. It's only used for display on the website.
type ServiceCheck struct {
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Expect   string        `json:"expect"`
	Timeout  cnfg.Duration `json:"timeout"`
	Interval cnfg.Duration `json:"interval"`
}

// IntList has a method to abstract lookups.
type IntList []int

// Has returns true if the list has an instance ID.
func (l IntList) Has(instance int) bool {
	for _, i := range l {
		if instance == i {
			return true
		}
	}

	return false
}

// String returns the message text for a client info response.
func (c *ClientInfo) String() string {
	if c == nil {
		return "<nil>"
	}

	return c.User.WelcomeMSG
}

// IsSub returns true if the client is a subscriber. False otherwise.
func (c *ClientInfo) IsSub() bool {
	return c != nil && c.User.Subscriber
}

// IsPatron returns true if the client is a patron. False otherwise.
func (c *ClientInfo) IsPatron() bool {
	return c != nil && c.User.Patron
}

// GetClientInfo returns an error if the API key is wrong. Returns client info otherwise.
func (c *Config) GetClientInfo() (*ClientInfo, error) {
	c.extras.ciMutex.Lock()
	defer c.extras.ciMutex.Unlock()

	if c.extras.clientInfo != nil {
		return c.extras.clientInfo, nil
	}

	body, err := c.SendData(ClientRoute.Path("reload"), c.Info(), true)
	if err != nil {
		return nil, fmt.Errorf("sending client info: %w", err)
	}

	clientInfo := ClientInfo{}
	if err = json.Unmarshal(body.Details.Response, &clientInfo); err != nil {
		return &clientInfo, fmt.Errorf("parsing response: %w, %s", err, string(body.Details.Response))
	}

	// Only set this if there was no error.
	c.extras.clientInfo = &clientInfo

	return c.extras.clientInfo, nil
}

// Info is used for JSON input for our outgoing client info.
func (c *Config) Info() map[string]interface{} {
	numPlex := 0 // maybe one day we'll support more than 1 plex.
	if c.Plex.Configured() {
		numPlex = 1
	}

	numTautulli := 0 // maybe one day we'll support more than 1 tautulli.
	if c.Apps.Tautulli != nil && c.Apps.Tautulli.URL != "" && c.Apps.Tautulli.APIKey != "" {
		numTautulli = 1
	}

	return map[string]interface{}{
		"client": map[string]interface{}{
			"arch":      runtime.GOARCH,
			"buildDate": version.BuildDate,
			"goVersion": version.GoVersion,
			"os":        runtime.GOOS,
			"revision":  version.Revision,
			"version":   version.Version,
			"uptimeSec": time.Since(version.Started).Round(time.Second).Seconds(),
			"started":   version.Started,
			"docker":    mnd.IsDocker,
			"gui":       ui.HasGUI(),
		},
		"num": map[string]interface{}{
			"deluge":   len(c.Apps.Deluge),
			"lidarr":   len(c.Apps.Lidarr),
			"plex":     numPlex,
			"prowlarr": len(c.Apps.Prowlarr),
			"qbit":     len(c.Apps.Qbit),
			"radarr":   len(c.Apps.Radarr),
			"readarr":  len(c.Apps.Readarr),
			"tautulli": numTautulli,
			"sabnzbd":  len(c.Apps.SabNZB),
			"sonarr":   len(c.Apps.Sonarr),
		},
		"config": map[string]interface{}{
			"globalTimeout": c.Timeout.String(),
			"retries":       c.Retries,
			"apps":          c.getAppConfigs(),
		},
		"internal": map[string]interface{}{
			"stuckDur": stuckDur.String(),
			"pollDur":  pollDur.String(),
		},
		"services": c.Services,
	}
}

// HostInfoNoError will return nil if there is an error, otherwise a copy of the host info.
func (c *Config) HostInfoNoError() *host.InfoStat {
	if c.extras.hostInfo == nil {
		return nil
	}

	return &host.InfoStat{
		Hostname:             c.extras.hostInfo.Hostname,
		Uptime:               uint64(time.Now().Unix()) - c.extras.hostInfo.BootTime,
		BootTime:             c.extras.hostInfo.BootTime,
		OS:                   c.extras.hostInfo.OS,
		Platform:             c.extras.hostInfo.Platform,
		PlatformFamily:       c.extras.hostInfo.PlatformFamily,
		PlatformVersion:      c.extras.hostInfo.PlatformVersion,
		KernelVersion:        c.extras.hostInfo.KernelVersion,
		KernelArch:           c.extras.hostInfo.KernelArch,
		VirtualizationSystem: c.extras.hostInfo.VirtualizationSystem,
		VirtualizationRole:   c.extras.hostInfo.VirtualizationRole,
		HostID:               c.extras.hostInfo.HostID,
	}
}

// GetHostInfoUID attempts to make a unique machine identifier...
func (c *Config) GetHostInfoUID() (*host.InfoStat, error) {
	if c.extras.hostInfo != nil {
		return c.HostInfoNoError(), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:gomnd
	defer cancel()

	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting host info: %w", err)
	}

	syn, err := snapshot.GetSynology()
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
	if mnd.IsDocker && strings.HasSuffix(hostInfo.KernelVersion, "truenas") && len(hostInfo.Hostname) > 17 {
		hostInfo.Hostname = hostInfo.Hostname[:len(hostInfo.Hostname)-17]
	}

	c.extras.hostInfo = hostInfo

	return c.HostInfoNoError(), nil // return a copy.
}

func (c *Config) pollForReload(event EventType) {
	body, err := c.SendData(ClientRoute.Path(EventPoll), c.Info(), true)
	if err != nil {
		c.Errorf("[%s requested] Polling Notifiarr: %v", event, err)
		return
	}

	var v struct {
		Reload     bool      `json:"reload"`
		LastSync   time.Time `json:"lastSync"`
		LastChange time.Time `json:"lastChange"`
	}

	if err = json.Unmarshal(body.Details.Response, &v); err != nil {
		c.Errorf("[%s requested] Polling Notifiarr: %v", event, err)
		return
	}

	if v.Reload {
		c.Printf("[%s requested] Website indicated new configurations; reloading to pick them up!"+
			" Last Sync: %v, Last Change: %v, Diff: %v", event, v.LastSync, v.LastChange, v.LastSync.Sub(v.LastChange))
		c.Sighup <- &update.Signal{Text: "poll triggered reload"}
	} else if c.clientInfo == nil {
		c.Printf("[%s requested] API Key checked out, reloading to pick up configuration from website!", event)
		c.Sighup <- &update.Signal{Text: "client info reload"}
	}
}

func (c *Config) getAppConfigs() map[string]interface{} {
	apps := make(map[string][]map[string]interface{})
	add := func(i int, name string) map[string]interface{} {
		return map[string]interface{}{
			"name":     name,
			"instance": i + 1,
		}
	}

	for i, app := range c.Apps.Lidarr {
		apps["lidarr"] = append(apps["lidarr"], add(i, app.Name))
	}

	for i, app := range c.Apps.Prowlarr {
		apps["prowlarr"] = append(apps["prowlarr"], add(i, app.Name))
	}

	for i, app := range c.Apps.Radarr {
		apps["radarr"] = append(apps["radarr"], add(i, app.Name))
	}

	for i, app := range c.Apps.Readarr {
		apps["readarr"] = append(apps["readarr"], add(i, app.Name))
	}

	for i, app := range c.Apps.Sonarr {
		apps["sonarr"] = append(apps["sonarr"], add(i, app.Name))
	}

	// We do this so more apps can be added later.
	reApps := make(map[string]interface{})
	for k, v := range apps {
		reApps[k] = v
	}

	if u, err := c.Apps.Tautulli.GetUsers(); err != nil {
		c.Error("Getting Tautulli Users:", err)
	} else {
		reApps["tautulli"] = map[string]interface{}{"users": u.MapEmailName()}
	}

	return reApps
}
