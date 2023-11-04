package clientinfo

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/shirou/gopsutil/v3/host"
	"golift.io/version"
)

// Config is the data needed to send and retrieve client info.
type Config struct {
	// Actions *triggers.Actions
	*apps.Apps
	*website.Server
	CmdList []*cmdconfig.Config
}

// AppInfo contains exported info about this app and its host.
type AppInfo struct {
	// Client contains running client information.
	Client *AppInfoClient `json:"client"`
	// Num contains configured application counters.
	Num map[string]int `json:"num"`
	// Config contains running configuration information.
	Config AppInfoConfig `json:"config"`
	// Commands is the list of available commands.
	Commands []*cmdconfig.Config `json:"commands"`
	// Host contains host info.
	Host *host.InfoStat `json:"host"`
	// HostError has data if hostinfo has an error.
	HostError string `json:"hostError"`
	// AppsStatus is only returned on the version endpoint.
	AppsStatus *AppStatuses `json:"appsStatus"`
}

// AppInfoClient contains the client's exported host info.
type AppInfoClient struct {
	// Architecture.
	Arch string `json:"arch"`
	// Application Build Date.
	BuildDate string `json:"buildDate"`
	// Branch application built from.
	Branch string `json:"branch"`
	// Go Version app built with.
	GoVersion string `json:"goVersion"`
	// OS app is running on.
	OS string `json:"os"`
	// Application Revision (part of the version).
	Revision string `json:"revision"`
	// Application Version.
	Version string `json:"version"`
	// Uptime in seconds.
	UptimeSec int64 `json:"uptimeSec"`
	// Application start time.
	Started time.Time `json:"started"`
	// Running in docker?
	Docker bool `json:"docker"`
	// Application has a GUI? (windows/mac only)
	HasGUI bool `json:"hasGui"`
	// Listen is the IP and port the client has configured.
	Listen string `json:"listen"`
	// Application supports tunnelling.
	Tunnel bool `json:"tunnel"`
}

// AppInfoConfig contains exported running configuration information for this app.
type AppInfoConfig struct {
	WebsiteTimeout string      `json:"websiteTimeout"`
	Retries        int         `json:"retries"`
	Apps           *AppConfigs `json:"apps"`
}

// AppConfigs contains exported configurations for various integrations.
type AppConfigs struct {
	Lidarr   []*AppInfoAppConfig `json:"lidarr"`
	Prowlarr []*AppInfoAppConfig `json:"prowlarr"`
	Radarr   []*AppInfoAppConfig `json:"radarr"`
	Readarr  []*AppInfoAppConfig `json:"readarr"`
	Sonarr   []*AppInfoAppConfig `json:"sonarr"`
	Tautulli *AppInfoTautulli    `json:"tautulli"`
}

// AppInfoAppConfig Maps an instance to a name and/or other properties.
type AppInfoAppConfig struct {
	// The site-ID for the instance (1-index).
	Instance int `json:"instance"`
	// Instance name as configured in the client.
	Name string `json:"name"`
}

// AppInfoTautulli contains the Tautulli user map, fetched from Tautulli.
type AppInfoTautulli struct {
	// Tautulli userID -> email map.
	Users map[string]string `json:"users"`
}

// Info is used for JSON input for our outgoing app info.
func (c *Config) Info(ctx context.Context, startup bool) *AppInfo { //nolint:funlen
	numPlex := 0 // maybe one day we'll support more than 1 plex.
	if c.Apps.Plex.Enabled() {
		numPlex = 1
	}

	numTautulli := 0 // maybe one day we'll support more than 1 tautulli.
	if c.Apps.Tautulli.Enabled() {
		numTautulli = 1
	}

	host, err := c.GetHostInfo(ctx)
	if err == nil {
		err = fmt.Errorf("") //nolint:goerr113
	}

	split := strings.Split(c.Config.BindAddr, ":")

	port := split[0]
	if len(split) > 1 {
		port = split[1]
	}

	return &AppInfo{
		Client: &AppInfoClient{
			Arch:      runtime.GOARCH,
			BuildDate: version.BuildDate,
			Branch:    version.Branch,
			GoVersion: version.GoVersion,
			OS:        runtime.GOOS,
			Revision:  version.Revision,
			Version:   version.Version,
			UptimeSec: int64(time.Since(version.Started).Round(time.Second).Seconds()),
			Started:   version.Started,
			Docker:    mnd.IsDocker,
			HasGUI:    ui.HasGUI(),
			Listen:    GetOutboundIP() + ":" + port,
			Tunnel:    true, // no toggle for this.
		},
		Num: map[string]int{
			"nzbget":       len(c.Apps.NZBGet),
			"deluge":       len(c.Apps.Deluge),
			"lidarr":       len(c.Apps.Lidarr),
			"plex":         numPlex,
			"prowlarr":     len(c.Apps.Prowlarr),
			"qbit":         len(c.Apps.Qbit),
			"rtorrent":     len(c.Apps.Rtorrent),
			"transmission": len(c.Apps.Transmission),
			"radarr":       len(c.Apps.Radarr),
			"readarr":      len(c.Apps.Readarr),
			"tautulli":     numTautulli,
			"sabnzbd":      len(c.Apps.SabNZB),
			"sonarr":       len(c.Apps.Sonarr),
		},
		Config: AppInfoConfig{
			WebsiteTimeout: c.Server.Config.Timeout.String(),
			Retries:        c.Server.Config.Retries,
			Apps:           c.getAppConfigs(ctx, startup),
		},
		Commands:  c.CmdList,
		Host:      host,
		HostError: err.Error(),
	}
}

func (c *Config) getAppConfigs(ctx context.Context, startup bool) *AppConfigs {
	apps := new(AppConfigs)
	add := func(i int, name string) *AppInfoAppConfig {
		return &AppInfoAppConfig{
			Name:     name,
			Instance: i + 1,
		}
	}

	for i, app := range c.Apps.Lidarr {
		apps.Lidarr = append(apps.Lidarr, add(i, app.Name))
	}

	for i, app := range c.Apps.Prowlarr {
		apps.Prowlarr = append(apps.Prowlarr, add(i, app.Name))
	}

	for i, app := range c.Apps.Radarr {
		apps.Radarr = append(apps.Radarr, add(i, app.Name))
	}

	for i, app := range c.Apps.Readarr {
		apps.Readarr = append(apps.Readarr, add(i, app.Name))
	}

	for i, app := range c.Apps.Sonarr {
		apps.Sonarr = append(apps.Sonarr, add(i, app.Name))
	}

	if !startup {
		if u, err := c.tautulliUsers(ctx); err != nil {
			c.Error("Getting Tautulli Users:", err)
		} else {
			apps.Tautulli = &AppInfoTautulli{Users: u.MapIDName()}
		}
	}

	return apps
}

func (c *Config) tautulliUsers(ctx context.Context) (*tautulli.Users, error) {
	const tautulliUsersKey = "tautulliUsers"
	cacheUsers := data.Get(tautulliUsersKey)

	if cacheUsers != nil && cacheUsers.Data != nil && time.Since(cacheUsers.Time) < 10*time.Minute {
		users, _ := cacheUsers.Data.(*tautulli.Users)
		return users, nil
	}

	users, err := c.Apps.Tautulli.GetUsers(ctx)
	if err != nil {
		return users, fmt.Errorf("tautulli failed: %w", err)
	}

	data.Save(tautulliUsersKey, users)

	return users, nil
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "1.1.1.1:437")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return conn.LocalAddr().String()
	}

	return localAddr.IP.String()
}
