package website

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
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
		WelcomeMSG string  `json:"welcome"`
		Subscriber bool    `json:"subscriber"`
		Patron     bool    `json:"patron"`
		DevAllowed bool    `json:"devAllowed"`
		DateFormat PHPDate `json:"dateFormat"`
	} `json:"user"`
	Actions struct {
		Poll      bool             `json:"poll"`
		Plex      PlexConfig       `json:"plex"`      // Site Config for Plex.
		Apps      AppConfigs       `json:"apps"`      // Site Config for Starr.
		Dashboard DashConfig       `json:"dashboard"` // Site Config for Dashboard.
		Sync      SyncConfig       `json:"sync"`      // Site Config for TRaSH Sync.
		Gaps      GapsConfig       `json:"gaps"`      // Site Config for Radarr Gaps.
		Custom    []*CronConfig    `json:"custom"`    // Site config for Custom Crons.
		Snapshot  *snapshot.Config `json:"snapshot"`  // Site Config for System Snapshot.
	} `json:"actions"`
}

// CronConfig defines a custom GET timer from the website.
// Used to offload crons to clients.
type CronConfig struct {
	Name     string        `json:"name"`     // name of action.
	Interval cnfg.Duration `json:"interval"` // how often to GET this URI.
	URI      string        `json:"endpoint"` // endpoint for the URI.
	Desc     string        `json:"description"`
}

// SyncConfig is the configuration returned from the notifiarr website for CF/RP TraSH sync.
type SyncConfig struct {
	Interval        cnfg.Duration `json:"interval"`        // how often to fire in minutes.
	Radarr          int64         `json:"radarr"`          // items in sync
	RadarrInstances IntList       `json:"radarrInstances"` // which instance IDs we sync
	Sonarr          int64         `json:"sonarr"`          // items in sync
	SonarrInstances IntList       `json:"sonarrInstances"` // which instance IDs we sync
}

// DashConfig is the configuration returned from the notifiarr website for the dashboard configuration.
type DashConfig struct {
	Interval cnfg.Duration `json:"interval"` // how often to fire in minutes.
}

// AppConfig is the data that comes from the website for each Starr app.
type AppConfig struct {
	Instance int           `json:"instance"`
	Name     string        `json:"name"`
	Stuck    bool          `json:"stuck"`
	Corrupt  string        `json:"corrupt"`
	Backup   string        `json:"backup"`
	Interval cnfg.Duration `json:"interval"`
}

// AppConfigs is the configuration returned from the notifiarr website for Starr apps.
type AppConfigs struct {
	Lidarr   []*AppConfig `json:"lidarr"`
	Prowlarr []*AppConfig `json:"prowlarr"`
	Radarr   []*AppConfig `json:"radarr"`
	Readarr  []*AppConfig `json:"readarr"`
	Sonarr   []*AppConfig `json:"sonarr"`
}

// PlexConfig is the website-derived configuration for Plex.
type PlexConfig struct {
	Interval   cnfg.Duration `json:"interval"`
	TrackSess  bool          `json:"trackSessions"`
	AccountMap string        `json:"accountMap"`
	NoActivity bool          `json:"noActivity"`
	Delay      cnfg.Duration `json:"activityDelay"`
	Cooldown   cnfg.Duration `json:"cooldown"`
	SeriesPC   uint          `json:"seriesPc"`
	MoviesPC   uint          `json:"moviesPc"`
}

// GapsConfig is the configuration returned from the notifiarr website for Radarr Collection Gaps.
type GapsConfig struct {
	Instances IntList       `json:"instances"`
	Interval  cnfg.Duration `json:"interval"`
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

func (s *Server) HaveClientInfo() bool {
	s.ciMutex.RLock()
	defer s.ciMutex.RUnlock()

	return s.clientInfo != nil
}

// GetClientInfo returns an error if the API key is wrong. Returns client info otherwise.
func (s *Server) GetClientInfo() (*ClientInfo, error) {
	if s.HaveClientInfo() {
		return s.clientInfo, nil
	}

	body, err := s.GetData(&Request{
		Route:      ClientRoute,
		Event:      EventStart,
		Payload:    s.Info(),
		LogPayload: true,
	})
	if err != nil {
		return nil, fmt.Errorf("sending client info: %w", err)
	}

	clientInfo := ClientInfo{}
	if err = json.Unmarshal(body.Details.Response, &clientInfo); err != nil {
		return &clientInfo, fmt.Errorf("parsing response: %w, %s", err, string(body.Details.Response))
	}

	s.ciMutex.Lock()
	defer s.ciMutex.Unlock()
	// Only set this if there was no error.
	s.clientInfo = &clientInfo

	return s.clientInfo, nil
}

// Info is used for JSON input for our outgoing client info.
func (s *Server) Info() map[string]interface{} {
	numPlex := 0 // maybe one day we'll support more than 1 plex.
	if s.config.Plex.Configured() {
		numPlex = 1
	}

	numTautulli := 0 // maybe one day we'll support more than 1 tautulli.
	if s.config.Apps.Tautulli != nil && s.config.Apps.Tautulli.URL != "" && s.config.Apps.Tautulli.APIKey != "" {
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
			"nzbget":   len(s.config.Apps.NZBGet),
			"deluge":   len(s.config.Apps.Deluge),
			"lidarr":   len(s.config.Apps.Lidarr),
			"plex":     numPlex,
			"prowlarr": len(s.config.Apps.Prowlarr),
			"qbit":     len(s.config.Apps.Qbit),
			"rtorrent": len(s.config.Apps.Rtorrent),
			"radarr":   len(s.config.Apps.Radarr),
			"readarr":  len(s.config.Apps.Readarr),
			"tautulli": numTautulli,
			"sabnzbd":  len(s.config.Apps.SabNZB),
			"sonarr":   len(s.config.Apps.Sonarr),
		},
		"config": map[string]interface{}{
			"globalTimeout": s.config.Timeout.String(),
			"retries":       s.config.Retries,
			"apps":          s.getAppConfigs(),
		},
	}
}

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
func (s *Server) GetHostInfo() (*host.InfoStat, error) { //nolint:cyclop
	if s.hostInfo != nil {
		return s.hostInfoNoError(), nil
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

	if s.config.HostID != "" {
		hostInfo.HostID = s.config.HostID
	}

	// This only happens once.
	s.hostInfo = hostInfo

	return s.hostInfoNoError(), nil // return a copy.
}

func (s *Server) PollForReload(event EventType) {
	body, err := s.GetData(&Request{
		Route:      ClientRoute,
		Event:      EventPoll,
		Payload:    s.Info(),
		LogPayload: true,
	})
	if err != nil {
		s.config.Errorf("[%s requested] Polling Notifiarr: %v", event, err)
		return
	}

	var v struct {
		Reload     bool      `json:"reload"`
		LastSync   time.Time `json:"lastSync"`
		LastChange time.Time `json:"lastChange"`
	}

	if err = json.Unmarshal(body.Details.Response, &v); err != nil {
		s.config.Errorf("[%s requested] Polling Notifiarr: %v", event, err)
		return
	}

	if v.Reload {
		s.config.Printf("[%s requested] Website indicated new configurations; reloading to pick them up!"+
			" Last Sync: %v, Last Change: %v, Diff: %v", event, v.LastSync, v.LastChange, v.LastSync.Sub(v.LastChange))
		s.config.Sighup <- &update.Signal{Text: "poll triggered reload"}
	} else if s.clientInfo == nil {
		s.config.Printf("[%s requested] API Key checked out, reloading to pick up configuration from website!", event)
		s.config.Sighup <- &update.Signal{Text: "client info reload"}
	}
}

func (s *Server) getAppConfigs() map[string]interface{} {
	apps := make(map[string][]map[string]interface{})
	add := func(i int, name string) map[string]interface{} {
		return map[string]interface{}{
			"name":     name,
			"instance": i + 1,
		}
	}

	for i, app := range s.config.Apps.Lidarr {
		apps["lidarr"] = append(apps["lidarr"], add(i, app.Name))
	}

	for i, app := range s.config.Apps.Prowlarr {
		apps["prowlarr"] = append(apps["prowlarr"], add(i, app.Name))
	}

	for i, app := range s.config.Apps.Radarr {
		apps["radarr"] = append(apps["radarr"], add(i, app.Name))
	}

	for i, app := range s.config.Apps.Readarr {
		apps["readarr"] = append(apps["readarr"], add(i, app.Name))
	}

	for i, app := range s.config.Apps.Sonarr {
		apps["sonarr"] = append(apps["sonarr"], add(i, app.Name))
	}

	// We do this so more apps can be added later (Tautulli).
	reApps := make(map[string]interface{})
	for k, v := range apps {
		reApps[k] = v
	}

	if u, err := s.config.Apps.Tautulli.GetUsers(context.Background()); err != nil {
		s.config.Error("Getting Tautulli Users:", err)
	} else {
		reApps["tautulli"] = map[string]interface{}{"users": u.MapEmailName()}
	}

	return reApps
}
