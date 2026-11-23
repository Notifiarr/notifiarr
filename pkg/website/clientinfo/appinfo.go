package clientinfo

import (
	"context"
	"errors"
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
	"github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig"
	"github.com/Notifiarr/notifiarr/pkg/ui"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/shirou/gopsutil/v4/host"
	"golift.io/version"
)

// Config is the data needed to send and retrieve client info. This data is sent to the website.
type Config struct {
	// Actions *triggers.Actions
	*apps.Apps
	CmdList   []*cmdconfig.Config
	Endpoints []*epconfig.Endpoint
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
	// Endpoints is the list of available endpoint url passthrough configs.
	Endpoints []*epconfig.Endpoint `json:"endpoints"`
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
	Lidarr       []*AppInfoAppConfig `json:"lidarr"`
	Prowlarr     []*AppInfoAppConfig `json:"prowlarr"`
	Radarr       []*AppInfoAppConfig `json:"radarr"`
	Readarr      []*AppInfoAppConfig `json:"readarr"`
	Sonarr       []*AppInfoAppConfig `json:"sonarr"`
	Deluge       []*AppInfoAppConfig `json:"deluge"`
	NZBGet       []*AppInfoAppConfig `json:"nzbget"`
	Qbit         []*AppInfoAppConfig `json:"qbit"`
	Rtorrent     []*AppInfoAppConfig `json:"rtorrent"`
	SabNZB       []*AppInfoAppConfig `json:"sabnzbd"`
	Transmission []*AppInfoAppConfig `json:"transmission"`
	Plex         *AppInfoAppConfig   `json:"plex,omitempty"`
	Tautulli     *AppInfoTautulli    `json:"tautulli"`
}

// AppInfoAppConfig Maps an instance to a name and/or other properties.
type AppInfoAppConfig struct {
	// The site-ID for the instance (1-index).
	Instance int `json:"instance"`
	// Instance name as configured in the client.
	Name string `json:"name"`
}

// AppInfoTautulli contains the configured Tautulli instance and its user map.
type AppInfoTautulli struct {
	// Instance is 1 when Tautulli is configured.
	Instance int `json:"instance,omitempty"`
	// Name is the instance name configured in the client.
	Name string `json:"name,omitempty"`
	// Tautulli userID -> email map. Omitted on startup.
	Users map[string]string `json:"users,omitempty"`
}

// Info is used for JSON input for our outgoing app info.
func (c *Config) Info(ctx context.Context, startup bool) *AppInfo { //nolint:funlen
	ctx = mnd.SetID(ctx)
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: Config.Info", startup)
	defer mnd.Log.Trace(reqID, "end: Config.Info", startup)

	numPlex := 0 // maybe one day we'll support more than 1 plex.
	if c.Plex.Enabled() {
		numPlex = 1
	}

	numTautulli := 0 // maybe one day we'll support more than 1 tautulli.
	if c.Tautulli.Enabled() {
		numTautulli = 1
	}

	host, err := website.GetHostInfo(ctx)
	if err == nil {
		err = errors.New("") //nolint:err113
	}

	webconf := website.GetConfig()
	split := strings.Split(webconf.BindAddr, ":")

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
			Listen:    GetOutboundIP(ctx) + ":" + port,
			Tunnel:    true, // no toggle for this.
		},
		Num: map[string]int{
			"nzbget":       len(c.NZBGet),
			"deluge":       len(c.Deluge),
			"lidarr":       len(c.Lidarr),
			"plex":         numPlex,
			"prowlarr":     len(c.Prowlarr),
			"qbit":         len(c.Qbit),
			"rtorrent":     len(c.Rtorrent),
			"transmission": len(c.Transmission),
			"radarr":       len(c.Radarr),
			"readarr":      len(c.Readarr),
			"tautulli":     numTautulli,
			"sabnzbd":      len(c.SabNZB),
			"sonarr":       len(c.Sonarr),
		},
		Config: AppInfoConfig{
			WebsiteTimeout: webconf.Timeout.String(),
			Retries:        webconf.Retries,
			Apps:           c.getAppConfigs(ctx, startup),
		},
		Commands:  c.CmdList,
		Endpoints: c.Endpoints,
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

	for i, app := range c.Lidarr {
		apps.Lidarr = append(apps.Lidarr, add(i, app.Name))
	}

	for i, app := range c.Prowlarr {
		apps.Prowlarr = append(apps.Prowlarr, add(i, app.Name))
	}

	for i, app := range c.Radarr {
		apps.Radarr = append(apps.Radarr, add(i, app.Name))
	}

	for i, app := range c.Readarr {
		apps.Readarr = append(apps.Readarr, add(i, app.Name))
	}

	for i, app := range c.Sonarr {
		apps.Sonarr = append(apps.Sonarr, add(i, app.Name))
	}

	c.addDownloaderConfigs(apps, add)

	if c.Plex.Enabled() {
		// Server.Name() is the Plex server name, not the configured instance name.
		apps.Plex = add(0, c.Plex.ExtraConfig.Name)
	}

	if c.Tautulli.Enabled() {
		apps.Tautulli = &AppInfoTautulli{
			Instance: 1,
			Name:     c.Tautulli.Name,
		}
	}

	if !startup {
		c.addTautulliUsers(ctx, apps)
	}

	return apps
}

func (c *Config) addDownloaderConfigs(apps *AppConfigs, add func(int, string) *AppInfoAppConfig) {
	for i, app := range c.Deluge {
		apps.Deluge = append(apps.Deluge, add(i, app.Name))
	}

	for i, app := range c.NZBGet {
		apps.NZBGet = append(apps.NZBGet, add(i, app.Name))
	}

	for i, app := range c.Qbit {
		apps.Qbit = append(apps.Qbit, add(i, app.Name))
	}

	for i, app := range c.Rtorrent {
		apps.Rtorrent = append(apps.Rtorrent, add(i, app.Name))
	}

	for i, app := range c.SabNZB {
		apps.SabNZB = append(apps.SabNZB, add(i, app.Name))
	}

	for i, app := range c.Transmission {
		apps.Transmission = append(apps.Transmission, add(i, app.Name))
	}
}

func (c *Config) addTautulliUsers(ctx context.Context, apps *AppConfigs) {
	users, err := c.tautulliUsers(ctx)
	if err != nil {
		mnd.Log.Errorf(mnd.GetID(ctx), "Getting Tautulli Users: %v", err)
		return
	}

	if apps.Tautulli == nil {
		apps.Tautulli = &AppInfoTautulli{}
	}

	apps.Tautulli.Users = users.MapIDName()
}

func (c *Config) tautulliUsers(ctx context.Context) (*tautulli.Users, error) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: Config.tautulliUsers")
	defer mnd.Log.Trace(reqID, "end: Config.tautulliUsers")

	const tautulliUsersKey = "tautulliUsers"
	cacheUsers := data.Get(tautulliUsersKey)

	if cacheUsers != nil && cacheUsers.Data != nil && time.Since(cacheUsers.Time) < 10*time.Minute {
		users, _ := cacheUsers.Data.(*tautulli.Users)
		return users, nil
	}

	users, err := c.Tautulli.GetUsers(ctx)
	if err != nil {
		return users, fmt.Errorf("tautulli failed: %w", err)
	}

	data.Save(tautulliUsersKey, users)

	return users, nil
}

func GetOutboundIP(ctx context.Context) string {
	dialer := &net.Dialer{Timeout: 1 * time.Second}

	conn, err := dialer.DialContext(ctx, "udp", "1.1.1.1:437")
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
