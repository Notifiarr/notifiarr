package client

import (
	"context"
	"net/http"
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/plex"
)

/* The version handler gets the version from a bunch of apps and returns them. */

type conTest struct {
	Instance int         `json:"instance"`
	Up       bool        `json:"up"`
	Status   interface{} `json:"systemStatus,omitempty"`
}

// versionHandler returns application run and build time data and application statuses: /api/version.
func (c *Client) versionHandler(r *http.Request) (int, interface{}) {
	output := c.website.Info()
	output["appsStatus"] = c.appStatsForVersion(r.Context())

	if host, err := c.website.GetHostInfoUID(); err != nil {
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

	getPlexVersion(ctx, &wg, c.Config.Plex, &plx)
	getLidarrVersion(ctx, &wg, c.Config.Apps.Lidarr, lid)
	getProwlarrVersion(ctx, &wg, c.Config.Apps.Prowlarr, prl)
	getRadarrVersion(ctx, &wg, c.Config.Apps.Radarr, rad)
	getReadarrVersion(ctx, &wg, c.Config.Apps.Readarr, read)
	getSonarrVersion(ctx, &wg, c.Config.Apps.Sonarr, son)
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

func getLidarrVersion(ctx context.Context, wait *sync.WaitGroup, lidarrs []*apps.LidarrConfig, lid []*conTest) {
	for idx, app := range lidarrs {
		wait.Add(1)

		go func(idx int, app *apps.LidarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			lid[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getProwlarrVersion(ctx context.Context, wait *sync.WaitGroup, prowlarrs []*apps.ProwlarrConfig, prl []*conTest) {
	for idx, app := range prowlarrs {
		wait.Add(1)

		go func(idx int, app *apps.ProwlarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			prl[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getRadarrVersion(ctx context.Context, wait *sync.WaitGroup, radarrs []*apps.RadarrConfig, rad []*conTest) {
	for idx, app := range radarrs {
		wait.Add(1)

		go func(idx int, app *apps.RadarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			rad[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getReadarrVersion(ctx context.Context, wait *sync.WaitGroup, readarrs []*apps.ReadarrConfig, read []*conTest) {
	for idx, app := range readarrs {
		wait.Add(1)

		go func(idx int, app *apps.ReadarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			read[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getSonarrVersion(ctx context.Context, wait *sync.WaitGroup, sonarrs []*apps.SonarrConfig, son []*conTest) {
	for idx, app := range sonarrs {
		wait.Add(1)

		go func(idx int, app *apps.SonarrConfig) {
			defer wait.Done()

			stat, err := app.GetSystemStatusContext(ctx)
			son[idx] = &conTest{Instance: idx + 1, Up: err == nil, Status: stat}
		}(idx, app)
	}
}

func getPlexVersion(ctx context.Context, wait *sync.WaitGroup, plexServer *plex.Server, plx *[]*conTest) {
	if !plexServer.Configured() {
		return
	}

	wait.Add(1)

	go func() {
		defer wait.Done()

		stat, err := plexServer.GetInfo(ctx)
		if stat == nil {
			stat = &plex.PMSInfo{}
		}

		*plx = []*conTest{{
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
	}()
}
