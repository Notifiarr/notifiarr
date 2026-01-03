package services

import (
	"net/url"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"golift.io/cnfg"
)

// PlexServerName is hard coded as the service name for Plex.
const PlexServerName = "Plex Server"

const (
	starrV3StatusURI = "/api/v3/system/status|X-API-Key:"
	starrV1StatusURI = "/api/v1/system/status|X-API-Key:"
)

// AddApps turns app configs into service checks if they have a name.
func (s *Services) AddApps(apps *apps.Apps, mysql []snapshot.MySQLConfig) {
	svcs := []*Service{}
	svcs = collectLidarrApps(svcs, apps.Lidarr)
	svcs = collectProwlarrApps(svcs, apps.Prowlarr)
	svcs = collectRadarrApps(svcs, apps.Radarr)
	svcs = collectReadarrApps(svcs, apps.Readarr)
	svcs = collectSonarrApps(svcs, apps.Sonarr)
	svcs = collectDelugeApps(svcs, apps.Deluge)
	svcs = collectNZBGetApps(svcs, apps.NZBGet)
	svcs = collectQbittorrentApps(svcs, apps.Qbit)
	svcs = collectRtorrentApps(svcs, apps.Rtorrent)
	svcs = collectSabNZBApps(svcs, apps.SabNZB)
	svcs = collectXmissionApps(svcs, apps.Transmission)
	svcs = collectTautulliApps(svcs, apps.Tautulli)
	svcs = collectPlexApps(svcs, &apps.Plex)
	svcs = collectMySQLApps(svcs, mysql)
	now := time.Now()

	for _, svc := range svcs {
		svc.ServiceConfig.validated = true
		svc.log = s.log
		svc.State = StateUnknown
		svc.Since = now
		s.add(svc.ServiceConfig)
	}
}

func collectLidarrApps(svcs []*Service, lidarr []apps.Lidarr) []*Service {
	for _, app := range lidarr {
		if !app.Enabled() || app.Name == "" {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    app.URL + starrV1StatusURI + app.APIKey,
					Expect:   "200",
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectProwlarrApps(svcs []*Service, prowlarr []apps.Prowlarr) []*Service {
	for _, app := range prowlarr {
		if !app.Enabled() || app.Name == "" {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    app.URL + starrV1StatusURI + app.APIKey,
					Expect:   "200",
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectRadarrApps(svcs []*Service, radarr []apps.Radarr) []*Service {
	for _, app := range radarr {
		if !app.Enabled() || app.Name == "" {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    app.URL + starrV3StatusURI + app.APIKey,
					Expect:   "200",
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectReadarrApps(svcs []*Service, readarr []apps.Readarr) []*Service {
	for _, app := range readarr {
		if !app.Enabled() || app.Name == "" {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    app.URL + starrV1StatusURI + app.APIKey,
					Expect:   "200",
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectSonarrApps(svcs []*Service, sonarr []apps.Sonarr) []*Service {
	for _, app := range sonarr {
		if !app.Enabled() || app.Name == "" {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    app.URL + starrV3StatusURI + app.APIKey,
					Expect:   "200",
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectDelugeApps(svcs []*Service, deluge []apps.Deluge) []*Service {
	// Deluge instanceapp.
	for _, app := range deluge {
		if !app.Enabled() || app.Name == "" {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    strings.TrimSuffix(app.Config.URL, "/json"),
					Expect:   "200",
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectNZBGetApps(svcs []*Service, nzbget []apps.NZBGet) []*Service {
	// NZBGet instances.
	for _, app := range nzbget {
		if !app.Enabled() || app.Name == "" {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		prefix := "" // add auth to the url here. woo, hacky, but it works!

		if !strings.Contains(app.URL, "@") {
			user := url.PathEscape(app.User) + ":" + url.PathEscape(app.Pass) + "@"
			if prefix = "http://" + user; strings.HasPrefix(app.URL, "https://") {
				prefix = "https://" + user
			}
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    prefix + strings.TrimPrefix(strings.TrimPrefix(app.URL, "https://"), "http://"),
					Expect:   "200",
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectQbittorrentApps(svcs []*Service, qbit []apps.Qbit) []*Service {
	// Qbittorrent instanceapp.
	for _, app := range qbit {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    app.URL,
					Expect:   "200",
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectRtorrentApps(svcs []*Service, rtorrent []apps.Rtorrent) []*Service {
	// rTorrent instanceapp.
	for _, app := range rtorrent {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    app.URL,
					Expect:   "200,401", // could not find a 200...
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectSabNZBApps(svcs []*Service, sabnzb []apps.SabNZB) []*Service {
	// SabNBZd instanceapp.
	for _, app := range sabnzb {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    app.SabNZB.URL + "/api?mode=version&apikey=" + app.SabNZB.APIKey,
					Expect:   "200",
					Timeout:  cnfg.Duration{Duration: app.SabNZB.Timeout},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectXmissionApps(svcs []*Service, xmission []apps.Xmission) []*Service {
	// Transmission instances.
	for _, app := range xmission {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
			interval.Duration = MinimumCheckInterval
		}

		expect := "401"
		if app.User == "" {
			expect = "409"
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				ServiceConfig: &ServiceConfig{
					validSSL: app.ValidSSL,
					Name:     app.Name,
					Type:     CheckHTTP,
					Value:    app.URL,
					Expect:   expect, // no 200 from RPC endpoint.
					Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
					Interval: interval,
				},
			})
		}
	}

	return svcs
}

func collectTautulliApps(svcs []*Service, app apps.Tautulli) []*Service {
	// Tautulli instance (1).
	if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
		return svcs
	}

	interval := app.Interval
	if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
		interval.Duration = MinimumCheckInterval
	}

	svcs = append(svcs, &Service{
		ServiceConfig: &ServiceConfig{
			validSSL: app.ValidSSL,
			Name:     app.Name,
			Type:     CheckHTTP,
			Value:    app.Tautulli.URL + "/api/v2?cmd=status&apikey=" + app.Tautulli.APIKey,
			Expect:   "200",
			Timeout:  app.ExtraConfig.Timeout,
			Interval: interval,
		},
	})

	return svcs
}

func collectPlexApps(svcs []*Service, app *apps.Plex) []*Service {
	if !app.Enabled() || app.Interval.Duration < 0 {
		return svcs
	}

	interval := app.Interval
	if interval.Duration != 0 && interval.Duration < MinimumCheckInterval {
		interval.Duration = MinimumCheckInterval
	}

	svcs = append(svcs, &Service{
		ServiceConfig: &ServiceConfig{
			validSSL: app.ValidSSL,
			Name:     PlexServerName,
			Type:     CheckHTTP,
			Value:    app.Server.URL + "|X-Plex-Token:" + app.Server.Token,
			Expect:   "200",
			Timeout:  app.Timeout,
			Interval: interval,
		},
	})

	return svcs
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

		host := strings.TrimLeft(strings.TrimRight(app.Host, ")"), "@tcp(")
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
