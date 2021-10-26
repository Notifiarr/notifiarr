package services

import (
	"strings"

	"golift.io/cnfg"
)

// collectApps turns app configs into service checks if they have a name.
func (c *Config) collectApps() []*Service {
	svcs := []*Service{}
	svcs = c.collectLidarrApps(svcs)
	svcs = c.collectRadarrApps(svcs)
	svcs = c.collectReadarrApps(svcs)
	svcs = c.collectSonarrApps(svcs)
	svcs = c.collectDownloadApps(svcs)
	svcs = c.collectTautulliApp(svcs)
	svcs = c.collectMySQLApps(svcs)

	return svcs
}

func (c *Config) collectLidarrApps(svcs []*Service) []*Service {
	for _, a := range c.Apps.Lidarr {
		if a.Interval.Duration == 0 {
			a.Interval.Duration = DefaultCheckInterval
		}

		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:     a.Name,
				Type:     CheckHTTP,
				Value:    a.URL + "/api/v1/system/status?apikey=" + a.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: a.Timeout.Duration},
				Interval: a.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectRadarrApps(svcs []*Service) []*Service {
	for _, a := range c.Apps.Radarr {
		if a.Interval.Duration == 0 {
			a.Interval.Duration = DefaultCheckInterval
		}

		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:     a.Name,
				Type:     CheckHTTP,
				Value:    a.URL + "/api/v3/system/status?apikey=" + a.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: a.Timeout.Duration},
				Interval: a.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectReadarrApps(svcs []*Service) []*Service {
	for _, a := range c.Apps.Readarr {
		if a.Interval.Duration == 0 {
			a.Interval.Duration = DefaultCheckInterval
		}

		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:     a.Name,
				Type:     CheckHTTP,
				Value:    a.URL + "/api/v1/system/status?apikey=" + a.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: a.Timeout.Duration},
				Interval: a.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectSonarrApps(svcs []*Service) []*Service {
	for _, a := range c.Apps.Sonarr {
		if a.Interval.Duration == 0 {
			a.Interval.Duration = DefaultCheckInterval
		}

		if a.Name != "" {
			svcs = append(svcs, &Service{
				Name:     a.Name,
				Type:     CheckHTTP,
				Value:    a.URL + "/api/v3/system/status?apikey=" + a.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: a.Timeout.Duration},
				Interval: a.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectDownloadApps(svcs []*Service) []*Service {
	// Deluge instances.
	for _, d := range c.Apps.Deluge {
		if d.Interval.Duration == 0 {
			d.Interval.Duration = DefaultCheckInterval
		}

		if d.Name != "" {
			svcs = append(svcs, &Service{
				Name:     d.Name,
				Type:     CheckHTTP,
				Value:    d.Config.URL,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: d.Timeout.Duration},
				Interval: d.Interval,
			})
		}
	}

	// Qbittorrent instances.
	for _, q := range c.Apps.Qbit {
		if q.Interval.Duration == 0 {
			q.Interval.Duration = DefaultCheckInterval
		}

		if q.Name != "" {
			svcs = append(svcs, &Service{
				Name:     q.Name,
				Type:     CheckHTTP,
				Value:    q.URL,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: q.Timeout.Duration},
				Interval: q.Interval,
			})
		}
	}

	// SabNBZd instances.
	for _, s := range c.Apps.SabNZB {
		if s.Interval.Duration == 0 {
			s.Interval.Duration = DefaultCheckInterval
		}

		if s.Name != "" {
			svcs = append(svcs, &Service{
				Name:     s.Name,
				Type:     CheckHTTP,
				Value:    s.URL + "/api?mode=version&apikey=" + s.APIKey,
				Expect:   "200",
				Timeout:  cnfg.Duration{Duration: s.Timeout.Duration},
				Interval: s.Interval,
			})
		}
	}

	return svcs
}

func (c *Config) collectTautulliApp(svcs []*Service) []*Service {
	// Tautulli instance (1).
	if t := c.Apps.Tautulli; t != nil && t.URL != "" && t.Name != "" {
		if t.Interval.Duration == 0 {
			t.Interval.Duration = DefaultCheckInterval
		}

		svcs = append(svcs, &Service{
			Name:     t.Name,
			Type:     CheckHTTP,
			Value:    t.URL + "/api/v2?cmd=status&apikey=" + t.APIKey,
			Expect:   "200",
			Timeout:  t.Timeout,
			Interval: t.Interval,
		})
	}

	return svcs
}

func (c *Config) collectMySQLApps(svcs []*Service) []*Service {
	if c.Plugins == nil {
		return svcs
	}

	for _, m := range c.Plugins.MySQL {
		if m.Interval.Duration == 0 {
			m.Interval.Duration = DefaultCheckInterval
		}

		host := strings.TrimLeft(strings.TrimRight(m.Host, ")"), "@tcp(")
		if !strings.Contains(host, ":") {
			host += ":3306"
		}

		if m.Name != "" {
			svcs = append(svcs, &Service{
				Name:     m.Name,
				Type:     CheckTCP,
				Value:    host,
				Timeout:  m.Timeout,
				Interval: m.Interval,
			})
		}
	}

	return svcs
}
