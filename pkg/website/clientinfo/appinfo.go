package clientinfo

import (
	"context"
	"fmt"
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

type AppInfo struct {
	Client    *AppInfoClient      `json:"client"` // contains running client information.
	Num       map[string]int      `json:"num"`    // contains configured application counters.
	Config    AppInfoConfig       `json:"config"` // contains running configuration information.
	Commands  []*cmdconfig.Config `json:"commands"`
	Host      *host.InfoStat      `json:"host"`      // contains host info.
	HostError string              `json:"hostError"` // returned if hostinfo has an error.
	AppStatus map[string]any      `json:"appStatus"` // only returned on version endpoint.
}

type AppInfoClient struct {
	Arch      string    `jsno:"arch"`
	BuildDate string    `jsno:"buildDate"`
	GoVersion string    `jsno:"goVersion"`
	OS        string    `jsno:"os"`
	Revision  string    `jsno:"revision"`
	Version   string    `jsno:"version"`
	UptimeSec int64     `jsno:"uptimeSec"`
	Started   time.Time `jsno:"started"`
	Docker    bool      `jsno:"docker"`
	HasGUI    bool      `jsno:"hasGUI"`
}

type AppInfoConfig struct {
	WebsiteTimeout string      `json:"websiteTimeout"`
	Retries        int         `json:"retries"`
	Apps           *AppConfigs `json:"apps"`
}

type AppConfigs struct {
	Lidarr   []*AppInfoAppConfig `json:"lidarr"`
	Prowlarr []*AppInfoAppConfig `json:"prowlarr"`
	Radarr   []*AppInfoAppConfig `json:"radarr"`
	Readarr  []*AppInfoAppConfig `json:"readarr"`
	Sonarr   []*AppInfoAppConfig `json:"sonarr"`
	Tautulli *AppInfoTautulli    `json:"tautulli"`
}

type AppInfoAppConfig struct {
	Instance int    `json:"instance"`
	Name     string `json:"name"`
}

type AppInfoTautulli struct {
	Users map[string]string `json:"users"`
}

// Info is used for JSON input for our outgoing app info.
func (c *Config) Info(ctx context.Context) *AppInfo {
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

	return &AppInfo{
		Client: &AppInfoClient{
			Arch:      runtime.GOARCH,
			BuildDate: version.BuildDate,
			GoVersion: version.GoVersion,
			OS:        runtime.GOOS,
			Revision:  version.Revision,
			Version:   version.Version,
			UptimeSec: int64(time.Since(version.Started).Round(time.Second).Seconds()),
			Started:   version.Started,
			Docker:    mnd.IsDocker,
			HasGUI:    ui.HasGUI(),
		},
		Num: map[string]int{
			"nzbget":   len(c.Apps.NZBGet),
			"deluge":   len(c.Apps.Deluge),
			"lidarr":   len(c.Apps.Lidarr),
			"plex":     numPlex,
			"prowlarr": len(c.Apps.Prowlarr),
			"qbit":     len(c.Apps.Qbit),
			"rtorrent": len(c.Apps.Rtorrent),
			"radarr":   len(c.Apps.Radarr),
			"readarr":  len(c.Apps.Readarr),
			"tautulli": numTautulli,
			"sabnzbd":  len(c.Apps.SabNZB),
			"sonarr":   len(c.Apps.Sonarr),
		},
		Config: AppInfoConfig{
			WebsiteTimeout: c.Server.Config.Timeout.String(),
			Retries:        c.Server.Config.Retries,
			Apps:           c.getAppConfigs(ctx),
		},
		// Commands:  c.Actions.Commands.List(),
		Host:      host,
		HostError: err.Error(),
	}
}

func (c *Config) getAppConfigs(ctx context.Context) *AppConfigs {
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

	if u, err := c.tautulliUsers(ctx); err != nil {
		c.Error("Getting Tautulli Users:",
			strings.ReplaceAll(c.Apps.Tautulli.APIKey, "<redacted>", err.Error()))
	} else {
		apps.Tautulli = &AppInfoTautulli{
			Users: u.MapEmailName(),
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
