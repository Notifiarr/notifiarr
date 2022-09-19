//nolint:lll
package client

import (
	"context"
	"net/http"
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
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
	output["appsStatus"] = c.appStatsForVersion(r.Context())
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

	getPlexVersion(ctx, &wg, c.Config.Plex, &plx, c.Config.Serial)
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

func getLidarrVersion(ctx context.Context, wait *sync.WaitGroup, lidarrs []*apps.LidarrConfig, lid []*conTest, fg bool) {
	for idx, app := range lidarrs {
		if app.Enabled() {
			lid[idx] = &conTest{Instance: idx + 1, Up: false, Status: mnd.Disabled}
		}

		if fg {
			stat, err := app.GetSystemStatusContext(ctx)
			lid[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
		}

		wait.Add(1)

		go func(idx int, app *apps.LidarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
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
			prl[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
		}

		wait.Add(1)

		go func(idx int, app *apps.ProwlarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
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
			rad[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
		}

		wait.Add(1)

		go func(idx int, app *apps.RadarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
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
			read[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
		}

		wait.Add(1)

		go func(idx int, app *apps.ReadarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
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
			son[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
			continue //nolint:wsl
		}

		wait.Add(1)

		go func(idx int, app *apps.SonarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
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
