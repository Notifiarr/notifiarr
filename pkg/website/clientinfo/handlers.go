//nolint:lll,godot
package clientinfo

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/gorilla/mux"
)

/* The version handler gets the version from a bunch of apps and returns them. */

type conTest struct {
	Instance int         `json:"instance"`
	Up       bool        `json:"up"`
	Status   interface{} `json:"systemStatus,omitempty"`
}

// InfoHandler is like the version handler except it doesn't poll all the apps.
// @Summary      Returns information about the client's configuration.
// @Description  Retrieve client info.
// @Tags         client
// @Produce      json
// @Success      200  {object} AppInfo "contains all info except appStatus"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/info [get]
func (c *Config) InfoHandler(r *http.Request) (int, interface{}) {
	return http.StatusOK, c.Info(r.Context())
}

// VersionHandler returns application run and build time data and application statuses.
// @Summary      Returns information about the client's configuration, and polls multiple applications for up-status and version.
// @Description  Retrieve client info.
// @Tags         client
// @Produce      json
// @Success      200  {object} AppInfo "contains app info included appStatus"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/version [get]
func (c *Config) VersionHandler(r *http.Request) (int, interface{}) {
	output := c.Info(r.Context())

	instance, _ := strconv.Atoi(mux.Vars(r)["instance"])
	if app := mux.Vars(r)["app"]; app != "" && instance > 0 {
		output.AppStatus = c.appStatsForVersionInstance(r.Context(), app, instance)
	} else {
		output.AppStatus = c.appStatsForVersion(r.Context())
	}

	return http.StatusOK, output
}

// appStatsForVersion loops each app and gets the version info.
func (c *Config) appStatsForVersion(ctx context.Context) map[string]interface{} {
	var (
		lid  = make([]*conTest, len(c.Apps.Lidarr))
		prl  = make([]*conTest, len(c.Apps.Prowlarr))
		rad  = make([]*conTest, len(c.Apps.Radarr))
		read = make([]*conTest, len(c.Apps.Readarr))
		son  = make([]*conTest, len(c.Apps.Sonarr))
		plx  = []*conTest{}
		wg   sync.WaitGroup
	)

	getPlexVersion(ctx, &wg, c.Apps.Plex, &plx)
	getLidarrVersion(ctx, &wg, c.Apps.Lidarr, lid)
	getProwlarrVersion(ctx, &wg, c.Apps.Prowlarr, prl)
	getRadarrVersion(ctx, &wg, c.Apps.Radarr, rad)
	getReadarrVersion(ctx, &wg, c.Apps.Readarr, read)
	getSonarrVersion(ctx, &wg, c.Apps.Sonarr, son)
	wg.Wait()

	return map[string]interface{}{
		"lidarr":   lid,
		"radarr":   rad,
		"readarr":  read,
		"sonarr":   son,
		"prowlarr": prl,
		"plex":     plx,
	}
}

// appStatsForVersionInstance handles a single-app version request.
func (c *Config) appStatsForVersionInstance(ctx context.Context, app string, instance int) map[string]interface{} { //nolint:cyclop
	switch idx := instance - 1; app {
	case "lidarr":
		if instance <= len(c.Apps.Lidarr) {
			stat, err := c.Apps.Lidarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "radarr":
		if instance <= len(c.Apps.Radarr) {
			stat, err := c.Apps.Radarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "readarr":
		if instance <= len(c.Apps.Readarr) {
			stat, err := c.Apps.Readarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "sonarr":
		if instance <= len(c.Apps.Sonarr) {
			stat, err := c.Apps.Sonarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "prowlarr":
		if instance <= len(c.Apps.Prowlarr) {
			stat, err := c.Apps.Prowlarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "plex":
		return map[string]any{app: plexVersionReply(c.Apps.Plex.GetInfo(ctx))}
	case "tautulli":
		stat, err := c.Apps.Tautulli.GetInfo(ctx)
		data.SaveWithID(app+"Status", 1, stat)

		return map[string]any{app: []*conTest{{Instance: 1, Up: err == nil, Status: stat}}}
	}

	return nil
}

func getLidarrVersion(ctx context.Context, wait *sync.WaitGroup, lidarrs []*apps.LidarrConfig, lid []*conTest) {
	for idx, app := range lidarrs {
		if app.Enabled() {
			lid[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		wait.Add(1)

		go func(idx int, app *apps.LidarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("lidarrStatus", idx, stat)

			lid[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getProwlarrVersion(ctx context.Context, wait *sync.WaitGroup, prowlarrs []*apps.ProwlarrConfig, prl []*conTest) {
	for idx, app := range prowlarrs {
		if app.Enabled() {
			prl[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		wait.Add(1)

		go func(idx int, app *apps.ProwlarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("prowlarrStatus", idx, stat)

			prl[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getRadarrVersion(ctx context.Context, wait *sync.WaitGroup, radarrs []*apps.RadarrConfig, rad []*conTest) {
	for idx, app := range radarrs {
		if app.Enabled() {
			rad[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		wait.Add(1)

		go func(idx int, app *apps.RadarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("radarrStatus", idx, stat)

			rad[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getReadarrVersion(ctx context.Context, wait *sync.WaitGroup, readarrs []*apps.ReadarrConfig, read []*conTest) {
	for idx, app := range readarrs {
		if app.Enabled() {
			read[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		wait.Add(1)

		go func(idx int, app *apps.ReadarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("readarrStatus", idx, stat)

			read[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getSonarrVersion(ctx context.Context, wait *sync.WaitGroup, sonarrs []*apps.SonarrConfig, son []*conTest) {
	for idx, app := range sonarrs {
		if app.Enabled() {
			son[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		wait.Add(1)

		go func(idx int, app *apps.SonarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("sonarrStatus", idx, stat)

			son[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getPlexVersion(ctx context.Context, wait *sync.WaitGroup, plexServer *apps.PlexConfig, plx *[]*conTest) {
	if !plexServer.Enabled() {
		return
	}

	wait.Add(1)

	go func() {
		defer wait.Done()
		*plx = plexVersionReply(plexServer.GetInfo(ctx)) //nolint:wsl
	}()
}

func plexVersionReply(stat *plex.PMSInfo, err error) []*conTest {
	if stat == nil {
		stat = &plex.PMSInfo{}
	} else {
		data.Save("plexStatus", stat)
	}

	return []*conTest{{
		Instance: 1,
		Up:       err == nil,
		Status: map[string]interface{}{
			"friendlyName":             stat.FriendlyName,
			"version":                  stat.Version,
			"updatedAt":                stat.UpdatedAt,
			"platform":                 stat.Platform,
			"platformVersion":          stat.PlatformVersion,
			"size":                     stat.Size,
			"myPlexSigninState":        stat.MyPlexSigninState,
			"myPlexSubscription":       stat.MyPlexSubscription,
			"pushNotifications":        stat.PushNotifications,
			"streamingBrainVersion":    stat.StreamingBrainVersion,
			"streamingBrainABRVersion": stat.StreamingBrainABRVersion,
		},
	}}
}
