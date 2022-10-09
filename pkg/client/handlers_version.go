//nolint:lll
package client

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

// infoHandler is like the version handler except it doesn't poll all the apps.
func (c *Client) infoHandler(r *http.Request) (int, interface{}) {
	output := c.website.Info(r.Context())
	output["commands"] = c.triggers.Commands.List()

	if host, err := c.website.GetHostInfo(r.Context()); err != nil {
		output["hostError"] = err.Error()
	} else {
		output["host"] = host
	}

	return http.StatusOK, output
}

// versionHandler returns application run and build time data and application statuses: /api/version.
func (c *Client) versionHandler(r *http.Request) (int, interface{}) {
	output := c.website.Info(r.Context())

	instance, _ := strconv.Atoi(mux.Vars(r)["instance"])
	if app := mux.Vars(r)["app"]; app != "" && instance > 0 {
		output["appsStatus"] = c.appStatsForVersionInstance(r.Context(), app, instance)
	} else {
		output["appsStatus"] = c.appStatsForVersion(r.Context())
	}

	output["commands"] = c.triggers.Commands.List()

	if host, err := c.website.GetHostInfo(r.Context()); err != nil {
		output["hostError"] = err.Error()
	} else {
		output["host"] = host
	}

	return http.StatusOK, output
}

// appStatsForVersion loops each app and gets the version info.
func (c *Client) appStatsForVersion(ctx context.Context) map[string]interface{} {
	var (
		lid  = make([]*conTest, len(c.Config.Apps.Lidarr))
		prl  = make([]*conTest, len(c.Config.Apps.Prowlarr))
		rad  = make([]*conTest, len(c.Config.Apps.Radarr))
		read = make([]*conTest, len(c.Config.Apps.Readarr))
		son  = make([]*conTest, len(c.Config.Apps.Sonarr))
		plx  = []*conTest{}
		wg   sync.WaitGroup
	)

	getPlexVersion(ctx, &wg, c.Config.Apps.Plex, &plx, c.Config.Serial)
	getLidarrVersion(ctx, &wg, c.Config.Apps.Lidarr, lid, c.Config.Serial)
	getProwlarrVersion(ctx, &wg, c.Config.Apps.Prowlarr, prl, c.Config.Serial)
	getRadarrVersion(ctx, &wg, c.Config.Apps.Radarr, rad, c.Config.Serial)
	getReadarrVersion(ctx, &wg, c.Config.Apps.Readarr, read, c.Config.Serial)
	getSonarrVersion(ctx, &wg, c.Config.Apps.Sonarr, son, c.Config.Serial)
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
func (c *Client) appStatsForVersionInstance(ctx context.Context, app string, instance int) map[string]interface{} { //nolint:cyclop
	switch idx := instance - 1; app {
	case "lidarr":
		if instance <= len(c.Config.Apps.Lidarr) {
			stat, err := c.Config.Apps.Lidarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "radarr":
		if instance <= len(c.Config.Apps.Radarr) {
			stat, err := c.Config.Apps.Radarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "readarr":
		if instance <= len(c.Config.Apps.Readarr) {
			stat, err := c.Config.Apps.Readarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "sonarr":
		if instance <= len(c.Config.Apps.Sonarr) {
			stat, err := c.Config.Apps.Sonarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "prowlarr":
		if instance <= len(c.Config.Apps.Prowlarr) {
			stat, err := c.Config.Apps.Prowlarr[idx].GetSystemStatusContext(ctx)
			data.SaveWithID(app+"Status", idx, stat)

			return map[string]any{app: []*conTest{{Instance: instance, Up: err == nil, Status: stat}}}
		}
	case "plex":
		return map[string]any{app: plexVersionReply(c.Config.Plex.GetInfo(ctx))}
	case "tautulli":
		stat, err := c.Config.Apps.Tautulli.GetInfo(ctx)
		data.SaveWithID(app+"Status", 1, stat)

		return map[string]any{app: []*conTest{{Instance: 1, Up: err == nil, Status: stat}}}
	}

	return nil
}

func getLidarrVersion(ctx context.Context, wait *sync.WaitGroup, lidarrs []*apps.LidarrConfig, lid []*conTest, fg bool) {
	for idx, app := range lidarrs {
		if app.Enabled() {
			lid[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		if fg {
			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("lidarrStatus", idx, stat)

			lid[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
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

func getProwlarrVersion(ctx context.Context, wait *sync.WaitGroup, prowlarrs []*apps.ProwlarrConfig, prl []*conTest, fg bool) {
	for idx, app := range prowlarrs {
		if app.Enabled() {
			prl[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		if fg {
			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("prowlarrStatus", idx, stat)

			prl[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
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

func getRadarrVersion(ctx context.Context, wait *sync.WaitGroup, radarrs []*apps.RadarrConfig, rad []*conTest, fg bool) {
	for idx, app := range radarrs {
		if app.Enabled() {
			rad[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		if fg {
			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("radarrStatus", idx, stat)

			rad[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
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

func getReadarrVersion(ctx context.Context, wait *sync.WaitGroup, readarrs []*apps.ReadarrConfig, read []*conTest, fg bool) {
	for idx, app := range readarrs {
		if app.Enabled() {
			read[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		if fg {
			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("readarrStatus", idx, stat)

			read[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
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

func getSonarrVersion(ctx context.Context, wait *sync.WaitGroup, sonarrs []*apps.SonarrConfig, son []*conTest, fg bool) {
	for idx, app := range sonarrs {
		if app.Enabled() {
			son[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		if fg {
			stat, err := app.GetSystemStatusContext(ctx)
			data.SaveWithID("sonarrStatus", idx, stat)

			son[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
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

func getPlexVersion(ctx context.Context, wait *sync.WaitGroup, plexServer *apps.PlexConfig, plx *[]*conTest, fg bool) {
	if !plexServer.Enabled() {
		return
	}

	if fg {
		*plx = plexVersionReply(plexServer.GetInfo(ctx))
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
