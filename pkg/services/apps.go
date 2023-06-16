package services

import (
	"net/url"
	"strings"

	"golift.io/cnfg"
)

// collectApps turns app configs into service checks if they have a name.
func (c *Config) collectApps() []*Service {
	svcs := []*Service{}
	svcs = c.collectLidarrApps(svcs)
	svcs = c.collectProwlarrApps(svcs)
	svcs = c.collectRadarrApps(svcs)
	svcs = c.collectReadarrApps(svcs)
	svcs = c.collectSonarrApps(svcs)
	svcs = c.collectDownloadApps(svcs)
	svcs = c.collectTautulliApp(svcs)
	svcs = c.collectPlexApp(svcs)
	svcs = c.collectMySQLApps(svcs)

	return svcs
}

func (c *Config) collectLidarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Lidarr {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v1/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	return svcs
}

func (c *Config) collectProwlarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Prowlarr {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v1/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	return svcs
}

func (c *Config) collectRadarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Radarr {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v3/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	return svcs
}

func (c *Config) collectReadarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Readarr {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v1/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	return svcs
}

func (c *Config) collectSonarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Sonarr {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v3/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	return svcs
}

//nolint:funlen,cyclop,gocognit,gocyclo // split this one up.
func (c *Config) collectDownloadApps(svcs []*Service) []*Service {
	// Deluge instanceapp.
	for _, app := range c.Apps.Deluge {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    strings.TrimSuffix(app.Config.URL, "/json"),
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	// NZBGet instances.
	for _, app := range c.Apps.NZBGet {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		prefix := "" // add auth to the url here. woo, hacky, but it works!

		if !strings.Contains(app.Config.URL, "@") {
			user := url.PathEscape(app.User) + ":" + url.PathEscape(app.Pass) + "@"
			if prefix = "http://" + user; strings.HasPrefix(app.Config.URL, "https://") {
				prefix = "https://" + user
			}
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    prefix + strings.TrimPrefix(strings.TrimPrefix(app.Config.URL, "https://"), "http://"),
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	// Qbittorrent instanceapp.
	for _, app := range c.Apps.Qbit {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	// rTorrent instanceapp.
	for _, app := range c.Apps.Rtorrent {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL,
				Expect:   "200,401", // could not find a 200...
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	// SabNBZd instanceapp.
	for _, app := range c.Apps.SabNZB {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api?mode=version&apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	// Transmission instances.
	for _, app := range c.Apps.Transmission {
		if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
			continue
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		expect := "401"
		if app.User == "" {
			expect = "409"
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL,
				Expect:   expect, // no 200 from RPC endpoint.
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: interval,
				validSSL: app.ValidSSL,
			})
		}
	}

	return svcs
}

func (c *Config) collectTautulliApp(svcs []*Service) []*Service {
	// Tautulli instance (1).
	app := c.Apps.Tautulli
	if !app.Enabled() || app.Name == "" || app.Interval.Duration < 0 {
		return svcs
	}

	interval := app.Interval
	if interval.Duration == 0 {
		interval.Duration = DefaultCheckInterval
	}

	svcs = append(svcs, &Service{
		Name:     app.Name,
		Type:     CheckHTTP,
		Value:    app.URL + "/api/v2?cmd=status&apikey=" + app.APIKey,
		Expect:   "200",
		Timeout:  app.Timeout,
		Interval: interval,
		validSSL: app.ValidSSL,
	})

	return svcs
}

func (c *Config) collectPlexApp(svcs []*Service) []*Service {
	app := c.Apps.Plex
	if !app.Enabled() || app.Interval.Duration < 0 {
		return svcs
	}

	interval := app.Interval
	if interval.Duration == 0 {
		interval.Duration = DefaultCheckInterval
	}

	svcs = append(svcs, &Service{
		Name:     "Plex Server",
		Type:     CheckHTTP,
		Value:    app.URL + "?X-Plex-Token=" + app.Token,
		Expect:   "200",
		Timeout:  app.Timeout,
		Interval: interval,
		validSSL: app.ValidSSL,
	})

	return svcs
}

func (c *Config) collectMySQLApps(svcs []*Service) []*Service { //nolint:cyclop
	if c.Plugins == nil {
		return svcs
	}

	for _, app := range c.Plugins.MySQL {
		if app.Host == "" || app.Timeout.Duration < 0 || app.Interval.Duration < 0 {
			continue
		}

		if app.Timeout.Duration == 0 {
			app.Timeout.Duration = DefaultTimeout
		}

		interval := app.Interval
		if interval.Duration == 0 {
			interval.Duration = DefaultCheckInterval
		}

		host := strings.TrimLeft(strings.TrimRight(app.Host, ")"), "@tcp(")
		if app.Name == "" || host == "" || strings.HasPrefix(host, "@") {
			continue
		}

		if !strings.Contains(host, ":") {
			host += ":3306"
		}

		svcs = append(svcs, &Service{
			Name:     app.Name,
			Type:     CheckTCP,
			Value:    host,
			Timeout:  app.Timeout,
			Interval: interval,
		})
	}

	return svcs
}
