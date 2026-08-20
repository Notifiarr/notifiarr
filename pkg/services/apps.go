package services

import (
	"net/url"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
)

// PlexServerName is hard coded as the service name for Plex.
const PlexServerName = "Plex Server"

const (
	starrV3StatusURI = "/api/v3/system/status|X-API-Key:"
	starrV1StatusURI = "/api/v1/system/status|X-API-Key:"
)

type starrChecker interface {
	Enabled() bool
	Starr() apps.StarrApp
}

// AddApps turns app configs into service checks if they have a name.
func (s *Services) AddApps(apps *apps.Apps, mysql []snapshot.MySQLConfig) {
	svcs := []*Service{}
	svcs = collectStarrApps(svcs, apps.Lidarr, starrV1StatusURI)
	svcs = collectStarrApps(svcs, apps.Prowlarr, starrV1StatusURI)
	svcs = collectStarrApps(svcs, apps.Radarr, starrV3StatusURI)
	svcs = collectStarrApps(svcs, apps.Readarr, starrV1StatusURI)
	svcs = collectStarrApps(svcs, apps.Sonarr, starrV3StatusURI)
	svcs = collectDelugeApps(svcs, apps.Deluge)
	svcs = collectNZBGetApps(svcs, apps.NZBGet)
	svcs = collectQbittorrentApps(svcs, apps.Qbit)
	svcs = collectRtorrentApps(svcs, apps.Rtorrent)
	svcs = collectSabNZBApps(svcs, apps.SabNZB)
	svcs = collectXmissionApps(svcs, apps.Transmission)
	svcs = collectTautulliApps(svcs, apps.Tautulli)
	svcs = collectPlexApps(svcs, &apps.Plex)
	svcs = collectMySQLApps(svcs, mysql)

	for _, svc := range svcs {
		if err := svc.Validate(); err != nil {
			mnd.Log.Errorf("called", "Skipping invalid app service check %s: %v", svc.Name, err)
			continue
		}

		s.add(svc.ServiceConfig)
	}
}

func collectStarrApps[T starrChecker](svcs []*Service, list []T, statusURI string) []*Service {
	for _, app := range list {
		if !app.Enabled() {
			continue
		}

		starr := app.Starr()
		svcs = appendHTTPCheck(svcs, starr.ExtraConfig, starr.URL+statusURI+starr.APIKey, "200")
	}

	return svcs
}

func appendHTTPCheck(svcs []*Service, extra apps.ExtraConfig, value, expect string) []*Service {
	if extra.Name == "" || extra.Interval.Duration < 0 {
		return svcs
	}

	interval := extra.Interval
	if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
		interval.Duration = MinimumCheckInterval
	}

	return append(svcs, &Service{
		ServiceConfig: &ServiceConfig{
			validSSL: extra.ValidSSL,
			Name:     extra.Name,
			Type:     CheckHTTP,
			Value:    value,
			Expect:   expect,
			Timeout:  extra.Timeout,
			Interval: interval,
		},
	})
}

func collectDelugeApps(svcs []*Service, deluge []apps.Deluge) []*Service {
	for _, app := range deluge {
		if !app.Enabled() {
			continue
		}

		svcs = appendHTTPCheck(svcs, app.ExtraConfig, strings.TrimSuffix(app.URL, "/json"), "200")
	}

	return svcs
}

func collectNZBGetApps(svcs []*Service, nzbget []apps.NZBGet) []*Service {
	for _, app := range nzbget {
		if !app.Enabled() {
			continue
		}

		svcs = appendHTTPCheck(svcs, app.ExtraConfig, nzbgetCheckURL(app), "200")
	}

	return svcs
}

func nzbgetCheckURL(app apps.NZBGet) string {
	if strings.Contains(app.URL, "@") {
		return app.URL
	}

	user := url.PathEscape(app.User) + ":" + url.PathEscape(app.Pass) + "@"
	scheme := "http://"

	if strings.HasPrefix(app.URL, "https://") {
		scheme = "https://"
	}

	host := strings.TrimPrefix(strings.TrimPrefix(app.URL, "https://"), "http://")

	return scheme + user + host
}

func collectQbittorrentApps(svcs []*Service, qbit []apps.Qbit) []*Service {
	for _, app := range qbit {
		if !app.Enabled() {
			continue
		}

		svcs = appendHTTPCheck(svcs, app.ExtraConfig, app.URL, "200")
	}

	return svcs
}

func collectRtorrentApps(svcs []*Service, rtorrent []apps.Rtorrent) []*Service {
	for _, app := range rtorrent {
		if !app.Enabled() {
			continue
		}

		svcs = appendHTTPCheck(svcs, app.ExtraConfig, app.URL, "200,401")
	}

	return svcs
}

func collectSabNZBApps(svcs []*Service, sabnzb []apps.SabNZB) []*Service {
	for _, app := range sabnzb {
		if !app.Enabled() {
			continue
		}

		svcs = appendHTTPCheck(svcs, app.ExtraConfig,
			app.SabNZBConfig.URL+"/api?mode=version&apikey="+app.SabNZBConfig.APIKey, "200")
	}

	return svcs
}

func collectXmissionApps(svcs []*Service, xmission []apps.Xmission) []*Service {
	for _, app := range xmission {
		if !app.Enabled() {
			continue
		}

		expect := "401"
		if app.User == "" {
			expect = "409"
		}

		svcs = appendHTTPCheck(svcs, app.ExtraConfig, app.URL, expect)
	}

	return svcs
}

func collectTautulliApps(svcs []*Service, app apps.Tautulli) []*Service {
	if !app.Enabled() {
		return svcs
	}

	return appendHTTPCheck(svcs, app.ExtraConfig,
		app.TautulliConfig.URL+"/api/v2?cmd=status&apikey="+app.TautulliConfig.APIKey, "200")
}

func collectPlexApps(svcs []*Service, app *apps.Plex) []*Service {
	if !app.Enabled() {
		return svcs
	}

	extra := app.ExtraConfig
	extra.Name = PlexServerName

	return appendHTTPCheck(svcs, extra, app.PlexConfig.URL+"|X-Plex-Token:"+app.PlexConfig.Token, "200")
}

func collectMySQLApps(svcs []*Service, mysql []snapshot.MySQLConfig) []*Service { //nolint:cyclop
	if mysql == nil {
		return svcs
	}

	for _, app := range mysql {
		if app.Host == "" || app.Timeout.Duration < 0 {
			continue
		}

		if app.Timeout.Duration == 0 {
			app.Timeout.Duration = DefaultTimeout
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		host := strings.TrimSuffix(strings.TrimPrefix(app.Host, "@tcp("), ")")
		if app.Name == "" || host == "" || strings.HasPrefix(host, "@") {
			continue
		}

		if !strings.Contains(host, ":") {
			host += ":3306"
		}

		svcs = append(svcs, &Service{
			ServiceConfig: &ServiceConfig{
				Name:     app.Name,
				Type:     CheckTCP,
				Value:    host,
				Timeout:  app.Timeout,
				Interval: interval,
			},
		})
	}

	return svcs
}
