package services

import (
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
	svcs = c.collectMySQLApps(svcs)

	return svcs
}

func (c *Config) collectLidarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Lidarr {
		if app.Interval.Duration == 0 {
			app.Interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v1/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: app.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectProwlarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Prowlarr {
		if app.Interval.Duration == 0 {
			app.Interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v1/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: app.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectRadarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Radarr {
		if app.Interval.Duration == 0 {
			app.Interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v3/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: app.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectReadarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Readarr {
		if app.Interval.Duration == 0 {
			app.Interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v1/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: app.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectSonarrApps(svcs []*Service) []*Service {
	for _, app := range c.Apps.Sonarr {
		if app.Interval.Duration == 0 {
			app.Interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api/v3/system/status?apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: app.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectDownloadApps(svcs []*Service) []*Service {
	// Deluge instanceapp.
	for _, app := range c.Apps.Deluge {
		if app.Interval.Duration == 0 {
			app.Interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    strings.TrimSuffix(app.Config.URL, "/json"),
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: app.Interval,
			})
		}
	}

	// Qbittorrent instanceapp.
	for _, app := range c.Apps.Qbit {
		if app.Interval.Duration == 0 {
			app.Interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: app.Interval,
			})
		}
	}

	// SabNBZd instanceapp.
	for _, app := range c.Apps.SabNZB {
		if app.Interval.Duration == 0 {
			app.Interval.Duration = DefaultCheckInterval
		}

		if app.Name != "" {
			svcs = append(svcs, &Service{
				Name:     app.Name,
				Type:     CheckHTTP,
				Value:    app.URL + "/api?mode=version&apikey=" + app.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: app.Timeout.Duration},
				Interval: app.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectTautulliApp(svcs []*Service) []*Service {
	// Tautulli instance (1).
	if app := c.Apps.Tautulli; app != nil && app.URL != "" && app.Name != "" {
		if app.Interval.Duration == 0 {
			app.Interval.Duration = DefaultCheckInterval
		}

		svcs = append(svcs, &Service{
			Name:     app.Name,
			Type:     CheckHTTP,
			Value:    app.URL + "/api/v2?cmd=status&apikey=" + app.APIKey,
			Expect:   "200",
			Timeout:  app.Timeout,
			Interval: app.Interval,
		})
	}

	return svcs
}

func (c *Config) collectMySQLApps(svcs []*Service) []*Service {
	if c.Plugins == nil {
		return svcs
	}

	for _, plugin := range c.Plugins.MySQL {
		if plugin.Interval.Duration == 0 {
			plugin.Interval.Duration = DefaultCheckInterval
		}

		if plugin.Timeout.Duration == 0 {
			plugin.Timeout.Duration = DefaultTimeout
		}

		host := strings.TrimLeft(strings.TrimRight(plugin.Host, ")"), "@tcp(")
		if plugin.Name == "" || host == "" || strings.HasPrefix(host, "@") {
			continue
		}

		if !strings.Contains(host, ":") {
			host += ":3306"
		}

		svcs = append(svcs, &Service{
			Name:     plugin.Name,
			Type:     CheckTCP,
			Value:    host,
			Timeout:  plugin.Timeout,
			Interval: plugin.Interval,
		})
	}

	return svcs
}
