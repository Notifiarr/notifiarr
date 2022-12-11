//nolint:lll,godot
package clientinfo

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/gorilla/mux"
	"golift.io/starr/lidarr"
	"golift.io/starr/prowlarr"
	"golift.io/starr/radarr"
	"golift.io/starr/readarr"
	"golift.io/starr/sonarr"
)

/* The version handler gets the version from a bunch of apps and returns them. */

type conTest struct {
	// The site-ID for the instance (1-index).
	Instance int `json:"instance"`
	// user-provided name of the instance.
	Name string `json:"name"`
	// Up is true if the instance is reachable.
	Up bool `json:"up"`
}

// LidarrConTest contains information about connected Lidarrs.
type LidarrConTest struct {
	conTest
	Status *lidarr.SystemStatus `json:"systemStatus,omitempty"`
}

// RadarrConTest contains information about connected Radarrs.
type RadarrConTest struct {
	conTest
	Status *radarr.SystemStatus `json:"systemStatus,omitempty"`
}

// ReadarrConTest contains information about connected Readarrs.
type ReadarrConTest struct {
	conTest
	Status *readarr.SystemStatus `json:"systemStatus,omitempty"`
}

// SonarrConTest contains information about connected Sonarrs.
type SonarrConTest struct {
	conTest
	Status *sonarr.SystemStatus `json:"systemStatus,omitempty"`
}

// ProwlarrConTest contains information about connected Prowlarrs.
type ProwlarrConTest struct {
	conTest
	Status *prowlarr.SystemStatus `json:"systemStatus,omitempty"`
}

// PlexConTest contains information about a connected Plex.
type PlexConTest struct {
	Status *PlexInfo `json:"systemStatus,omitempty"`
	conTest
}

// PlexInfo represents a small slice of the Plex Media Server Data.
// @Description Contains some very basic Plex data, including the name and version.
type PlexInfo struct {
	FriendlyName       string `json:"friendlyName"`
	Version            string `json:"version"`
	UpdatedAt          int64  `json:"updatedAt"`
	Platform           string `json:"platform"`
	PlatformVersion    string `json:"platformVersion"`
	Size               int64  `json:"size"`
	MyPlexSigninState  string `json:"myPlexSigninState"`
	MyPlexSubscription bool   `json:"myPlexSubscription"`
	PushNotifications  bool   `json:"pushNotifications"`
}

type TautulliConTest struct {
	conTest
	Status *tautulli.Info `json:"systemStatus,omitempty"`
}

// AppStatuses contains some integration up-statuses and versions.
type AppStatuses struct {
	Lidarr   []*LidarrConTest   `json:"lidarr,omitempty"`
	Radarr   []*RadarrConTest   `json:"radarr,omitempty"`
	Readarr  []*ReadarrConTest  `json:"readarr,omitempty"`
	Sonarr   []*SonarrConTest   `json:"sonarr,omitempty"`
	Prowlarr []*ProwlarrConTest `json:"prowlarr,omitempty"`
	Plex     []*PlexConTest     `json:"plex,omitempty"`
	Tautulli []*TautulliConTest `json:"tautulli,omitempty"`
}

// InfoHandler is like the version handler except it doesn't poll all the apps.
// @Description  Returns information about the client's configuration. This endpoint returns all the instance IDs (and instance names if present). Use the returned instance IDs with endpoints that accept an instance ID.
// @Summary      Retrieve client info.
// @Tags         Client
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=AppInfo} "contains all info except appStatus"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/info [get]
// @Security     ApiKeyAuth
//
//nolint:lll
func (c *Config) InfoHandler(r *http.Request) (int, interface{}) {
	return http.StatusOK, c.Info(r.Context(), false)
}

// VersionHandler returns application run and build time data and application statuses.
// @Description  Returns information about the client's configuration, and polls multiple applications for up-status and version.
// @Summary      Retrieve client info + starr/plex info.
// @Tags         Client
// @Produce      json
// @Success      200  {object} apps.Respond.apiResponse{message=AppInfo} "contains app info included appStatus"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/version [get]
// @Security     ApiKeyAuth
func (c *Config) VersionHandler(r *http.Request) (int, interface{}) {
	output := c.Info(r.Context(), false)
	output.AppsStatus = c.appStatsForVersion(r.Context())

	return http.StatusOK, output
}

// VersionHandlerInstance returns application run and build time data and the status for the requested instance.
// @Description  Returns information about the client's configuration, and polls 1 application instance for up-status and version.
// @Description  Plex and Tautulli only support app instance 1.
// @Summary      Retrieve client info + 1 app's info.
// @Tags         Client
// @Produce      json
// @Param        app      path string  true  "Application" Enums(lidarr, prowlarr, radarr, readarr, sonarr, plex, tautulli)
// @Param        instance path int64   true  "Application instance (1-index)."
// @Success      200  {object} apps.Respond.apiResponse{message=AppInfo} "contains app info included appStatus"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/version/{app}/{instance} [get]
// @Security     ApiKeyAuth
func (c *Config) VersionHandlerInstance(r *http.Request) (int, interface{}) {
	output := c.Info(r.Context(), false)
	instance, _ := strconv.Atoi(mux.Vars(r)["instance"])
	output.AppsStatus = c.appStatsForVersionInstance(r.Context(), mux.Vars(r)["app"], instance)

	return http.StatusOK, output
}

// appStatsForVersion loops each app and gets the version info.
func (c *Config) appStatsForVersion(ctx context.Context) *AppStatuses {
	var (
		lid  = make([]*LidarrConTest, len(c.Apps.Lidarr))
		prl  = make([]*ProwlarrConTest, len(c.Apps.Prowlarr))
		rad  = make([]*RadarrConTest, len(c.Apps.Radarr))
		read = make([]*ReadarrConTest, len(c.Apps.Readarr))
		son  = make([]*SonarrConTest, len(c.Apps.Sonarr))
		plx  = []*PlexConTest{}
		wg   sync.WaitGroup
	)

	getPlexVersion(ctx, &wg, c.Apps.Plex, &plx)
	getLidarrVersion(ctx, &wg, c.Apps.Lidarr, lid)
	getProwlarrVersion(ctx, &wg, c.Apps.Prowlarr, prl)
	getRadarrVersion(ctx, &wg, c.Apps.Radarr, rad)
	getReadarrVersion(ctx, &wg, c.Apps.Readarr, read)
	getSonarrVersion(ctx, &wg, c.Apps.Sonarr, son)
	wg.Wait()

	return &AppStatuses{
		Lidarr:   lid,
		Radarr:   rad,
		Readarr:  read,
		Sonarr:   son,
		Prowlarr: prl,
		Plex:     plx,
	}
}

// appStatsForVersionInstance handles a single-app version request.
func (c *Config) appStatsForVersionInstance(ctx context.Context, app string, instance int) *AppStatuses { //nolint:cyclop
	switch idx := instance - 1; app {
	case "lidarr":
		if instance <= len(c.Apps.Lidarr) {
			stat, err := c.Apps.Lidarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return &AppStatuses{Lidarr: []*LidarrConTest{{conTest{Instance: instance, Up: err == nil, Name: c.Apps.Lidarr[idx].Name}, stat}}}
		}
	case "radarr":
		if instance <= len(c.Apps.Radarr) {
			stat, err := c.Apps.Radarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return &AppStatuses{Radarr: []*RadarrConTest{{conTest{Instance: instance, Up: err == nil, Name: c.Apps.Radarr[idx].Name}, stat}}}
		}
	case "readarr":
		if instance <= len(c.Apps.Readarr) {
			stat, err := c.Apps.Readarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return &AppStatuses{Readarr: []*ReadarrConTest{{conTest{Instance: instance, Up: err == nil, Name: c.Apps.Readarr[idx].Name}, stat}}}
		}
	case "sonarr":
		if instance <= len(c.Apps.Sonarr) {
			stat, err := c.Apps.Sonarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return &AppStatuses{Sonarr: []*SonarrConTest{{conTest{Instance: instance, Up: err == nil, Name: c.Apps.Sonarr[idx].Name}, stat}}}
		}
	case "prowlarr":
		if instance <= len(c.Apps.Prowlarr) {
			stat, err := c.Apps.Prowlarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return &AppStatuses{Prowlarr: []*ProwlarrConTest{{conTest{Instance: instance, Up: err == nil, Name: c.Apps.Prowlarr[idx].Name}, stat}}}
		}
	case "plex":
		stat := &AppStatuses{Plex: plexVersionReply(c.Apps.Plex.GetInfo(ctx))}
		stat.Plex[0].Name = c.Apps.Plex.Server.Name()

		return stat
	case "tautulli":
		stat, err := c.Apps.Tautulli.GetInfo(ctx)
		data.SaveWithID(app+"Status", 1, stat)

		return &AppStatuses{Tautulli: []*TautulliConTest{{conTest{Instance: 1, Up: err == nil, Name: c.Apps.Tautulli.Name}, stat}}}
	}

	return nil
}

func getLidarrVersion(ctx context.Context, wait *sync.WaitGroup, lidarrs []*apps.LidarrConfig, lid []*LidarrConTest) {
	for idx, app := range lidarrs {
		if app.Enabled() {
			lid[idx] = &LidarrConTest{conTest: conTest{Instance: idx + 1, Up: false, Name: app.Name}}
		}

		wait.Add(1)

		go func(idx int, app *apps.LidarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("lidarrStatus", idx, stat)

			lid[idx] = &LidarrConTest{conTest: conTest{Instance: idx + 1, Up: err == nil, Name: app.Name}, Status: stat}
		}(idx, app)
	}
}

func getProwlarrVersion(ctx context.Context, wait *sync.WaitGroup, prowlarrs []*apps.ProwlarrConfig, prl []*ProwlarrConTest) {
	for idx, app := range prowlarrs {
		if app.Enabled() {
			prl[idx] = &ProwlarrConTest{conTest: conTest{Instance: idx + 1, Up: false, Name: app.Name}}
		}

		wait.Add(1)

		go func(idx int, app *apps.ProwlarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("prowlarrStatus", idx, stat)

			prl[idx] = &ProwlarrConTest{conTest: conTest{Instance: idx + 1, Up: err == nil, Name: app.Name}, Status: stat}
		}(idx, app)
	}
}

func getRadarrVersion(ctx context.Context, wait *sync.WaitGroup, radarrs []*apps.RadarrConfig, rad []*RadarrConTest) {
	for idx, app := range radarrs {
		if app.Enabled() {
			rad[idx] = &RadarrConTest{conTest: conTest{Instance: idx + 1, Up: false, Name: app.Name}}
		}

		wait.Add(1)

		go func(idx int, app *apps.RadarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("radarrStatus", idx, stat)

			rad[idx] = &RadarrConTest{conTest: conTest{Instance: idx + 1, Up: err == nil, Name: app.Name}, Status: stat}
		}(idx, app)
	}
}

func getReadarrVersion(ctx context.Context, wait *sync.WaitGroup, readarrs []*apps.ReadarrConfig, read []*ReadarrConTest) {
	for idx, app := range readarrs {
		if app.Enabled() {
			read[idx] = &ReadarrConTest{conTest: conTest{Instance: idx + 1, Up: false, Name: app.Name}}
		}

		wait.Add(1)

		go func(idx int, app *apps.ReadarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("readarrStatus", idx, stat)

			read[idx] = &ReadarrConTest{conTest: conTest{Instance: idx + 1, Up: err == nil, Name: app.Name}, Status: stat}
		}(idx, app)
	}
}

func getSonarrVersion(ctx context.Context, wait *sync.WaitGroup, sonarrs []*apps.SonarrConfig, son []*SonarrConTest) {
	for idx, app := range sonarrs {
		if app.Enabled() {
			son[idx] = &SonarrConTest{conTest: conTest{Instance: idx + 1, Up: false, Name: app.Name}}
		}

		wait.Add(1)

		go func(idx int, app *apps.SonarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("sonarrStatus", idx, stat)

			son[idx] = &SonarrConTest{conTest: conTest{Instance: idx + 1, Up: err == nil, Name: app.Name}, Status: stat}
		}(idx, app)
	}
}

func getPlexVersion(ctx context.Context, wait *sync.WaitGroup, plexServer *apps.PlexConfig, plx *[]*PlexConTest) {
	if !plexServer.Enabled() {
		return
	}

	wait.Add(1)

	go func() {
		defer wait.Done()
		*plx = plexVersionReply(plexServer.GetInfo(ctx)) //nolint:wsl
	}()
}

func plexVersionReply(stat *plex.PMSInfo, err error) []*PlexConTest {
	if stat == nil {
		stat = &plex.PMSInfo{}
	} else {
		data.Save("plexStatus", stat)
	}

	return []*PlexConTest{{
		&PlexInfo{
			FriendlyName:       stat.FriendlyName,
			Version:            stat.Version,
			UpdatedAt:          stat.UpdatedAt,
			Platform:           stat.Platform,
			PlatformVersion:    stat.PlatformVersion,
			Size:               stat.Size,
			MyPlexSigninState:  stat.MyPlexSigninState,
			MyPlexSubscription: stat.MyPlexSubscription,
			PushNotifications:  stat.PushNotifications,
		},
		conTest{Instance: 1, Up: err == nil},
	}}
}
