package services_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/cnfg"
	"golift.io/nzbget"
)

func TestCollectMySQLAppsPreservesHostnames(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mysql := []snapshot.MySQLConfig{
		{Name: "privatebin", Host: "privatebin-mariadb", Timeout: cnfg.Duration{Duration: time.Second}},
		{Name: "loopback", Host: "@tcp(127.0.0.1:3306)", Timeout: cnfg.Duration{Duration: time.Second}},
	}

	svc := services.New(&services.Config{})
	svc.AddApps(&apps.Apps{}, mysql)

	results := svc.GetResults()
	assert.Len(results, 2, "expected 2 service checks")

	got := resultsByName(results)
	assert.Equal("privatebin-mariadb:3306", got["privatebin"].Check, "privatebin host name mismatch")
	assert.Equal("127.0.0.1:3306", got["loopback"].Check, "tcp host value mismatch")
	assert.Equal(services.CheckTCP, got["privatebin"].Type)
	assert.Equal(services.StateUnknown, got["privatebin"].State)
}

func collectorApps() *apps.Apps { //nolint:funlen
	shortInterval := time.Second
	plexCfg := plex.Config{URL: "http://plex.example", Token: "plextok"}
	tautCfg := tautulli.Config{URL: "http://tautulli.example", APIKey: "tautkey"}
	sabCfg := sabnzbd.Config{URL: "http://sab.example", APIKey: "sabkey"}

	return &apps.Apps{
		Lidarr:   []apps.Lidarr{{StarrApp: starrApp("Lidarr", "http://lidarr.example", "lidkey", shortInterval)}},
		Prowlarr: []apps.Prowlarr{{StarrApp: starrApp("Prowlarr", "http://prowlarr.example", "prowlkey", 0)}},
		Radarr:   []apps.Radarr{{StarrApp: starrApp("Radarr", "http://radarr.example", "radkey", 0)}},
		Readarr:  []apps.Readarr{{StarrApp: starrApp("Readarr", "http://readarr.example", "readkey", 0)}},
		Sonarr:   []apps.Sonarr{{StarrApp: starrApp("Sonarr", "http://sonarr.example", "sonkey", 0)}},
		Deluge: []apps.Deluge{{
			ExtraConfig: extraConfig("Deluge", 0),
			URL:         "http://deluge.example/json",
		}},
		NZBGet: []apps.NZBGet{{
			NZBGetConfig: apps.NZBGetConfig{
				ExtraConfig: extraConfig("NZBGet", 0),
				Config:      nzbget.Config{URL: "http://nzbget.example", User: "nzb user", Pass: "p@ss"},
			},
		}},
		Qbit: []apps.Qbit{{
			ExtraConfig: extraConfig("qBittorrent", 0),
			URL:         "http://qbit.example",
		}},
		Rtorrent: []apps.Rtorrent{{
			ExtraConfig: extraConfig("rTorrent", 0),
			URL:         "http://rtorrent.example",
		}},
		SabNZB: []apps.SabNZB{{
			SabNZBConfig: apps.SabNZBConfig{
				ExtraConfig: extraConfig("SABnzbd", 0),
				Config:      &sabCfg,
			},
			SabNZB: sabnzbd.New(sabCfg, &http.Client{Timeout: time.Second}),
		}},
		Transmission: []apps.Xmission{
			{
				URL:         "http://transmission.example",
				User:        "user",
				ExtraConfig: extraConfig("TransmissionAuth", 0),
			},
			{
				URL:         "http://transmission-open.example",
				ExtraConfig: extraConfig("TransmissionOpen", 0),
			},
		},
		Tautulli: apps.Tautulli{
			TautulliConfig: apps.TautulliConfig{
				ExtraConfig: extraConfig("Tautulli", shortInterval),
				Config:      tautCfg,
			},
			Tautulli: tautulli.New(tautCfg, http.DefaultClient),
		},
		Plex: apps.Plex{
			PlexConfig: apps.PlexConfig{
				Config:      plexCfg,
				ExtraConfig: extraConfig("", 0),
			},
			Server: *plex.New(&plexCfg, nil),
		},
	}
}

func TestAddAppsCollectors(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	svc := services.New(&services.Config{})
	svc.AddApps(collectorApps(), nil)

	got := resultsByName(svc.GetResults())
	assert.Equal(14, svc.SvcCount())

	assert.Equal("http://lidarr.example/api/v1/system/status|X-API-Key:lidkey", got["Lidarr"].Check)
	assert.Equal(services.MinimumCheckInterval, got["Lidarr"].IntervalDur, "short intervals bump to the minimum")
	assert.Equal("http://prowlarr.example/api/v1/system/status|X-API-Key:prowlkey", got["Prowlarr"].Check)
	assert.Equal("http://radarr.example/api/v3/system/status|X-API-Key:radkey", got["Radarr"].Check)
	assert.Equal("http://readarr.example/api/v1/system/status|X-API-Key:readkey", got["Readarr"].Check)
	assert.Equal("http://sonarr.example/api/v3/system/status|X-API-Key:sonkey", got["Sonarr"].Check)
	assert.Equal("http://deluge.example", got["Deluge"].Check, "deluge /json suffix is stripped")
	assert.Equal("http://nzb%20user:p@ss@nzbget.example", got["NZBGet"].Check)
	assert.Equal("http://qbit.example", got["qBittorrent"].Check)
	assert.Equal("http://rtorrent.example", got["rTorrent"].Check)
	assert.Equal("200,401", got["rTorrent"].Expect)
	assert.Equal("http://sab.example/api?mode=version&apikey=sabkey", got["SABnzbd"].Check)
	assert.Equal("401", got["TransmissionAuth"].Expect)
	assert.Equal("409", got["TransmissionOpen"].Expect)
	assert.Equal("http://tautulli.example/api/v2?cmd=status&apikey=tautkey", got["Tautulli"].Check)
	assert.Equal(services.MinimumCheckInterval, got["Tautulli"].IntervalDur)
	assert.Equal(services.PlexServerName, got[services.PlexServerName].Name)
	assert.Equal("http://plex.example|X-Plex-Token:plextok", got[services.PlexServerName].Check)
	assert.Equal("200", got["Lidarr"].Expect)
	assert.Equal(services.CheckHTTP, got["Lidarr"].Type)
}

func TestAddAppsNZBGetURLVariants(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	svc := services.New(&services.Config{})
	svc.AddApps(&apps.Apps{
		NZBGet: []apps.NZBGet{
			{
				NZBGetConfig: apps.NZBGetConfig{
					ExtraConfig: extraConfig("AlreadyAuthed", 0),
					Config:      nzbget.Config{URL: "https://user:pass@nzbget.example", User: "ignored", Pass: "ignored"},
				},
			},
			{
				NZBGetConfig: apps.NZBGetConfig{
					ExtraConfig: extraConfig("HTTPS", 0),
					Config:      nzbget.Config{URL: "https://nzbget.example", User: "u", Pass: "p"},
				},
			},
		},
	}, nil)

	got := resultsByName(svc.GetResults())
	assert.Equal("https://user:pass@nzbget.example", got["AlreadyAuthed"].Check)
	assert.Equal("https://u:p@nzbget.example", got["HTTPS"].Check)
}

func TestAddAppsUsesAppConfigNotClients(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	tautCfg := tautulli.Config{URL: "http://tautulli.example", APIKey: "tautkey"}
	sabCfg := sabnzbd.Config{URL: "http://sab.example", APIKey: "sabkey"}
	plexCfg := plex.Config{URL: "http://plex.example", Token: "plextok"}

	svc := services.New(&services.Config{})
	svc.AddApps(&apps.Apps{
		Tautulli: apps.Tautulli{
			TautulliConfig: apps.TautulliConfig{
				ExtraConfig: extraConfig("Tautulli", 0),
				Config:      tautCfg,
			},
		},
		SabNZB: []apps.SabNZB{{
			SabNZBConfig: apps.SabNZBConfig{
				ExtraConfig: extraConfig("SABnzbd", 0),
				Config:      &sabCfg,
			},
		}},
		Plex: apps.Plex{
			PlexConfig: apps.PlexConfig{
				Config:      plexCfg,
				ExtraConfig: extraConfig("", 0),
			},
		},
	}, nil)

	got := resultsByName(svc.GetResults())
	assert.Equal("http://tautulli.example/api/v2?cmd=status&apikey=tautkey", got["Tautulli"].Check)
	assert.Equal("http://sab.example/api?mode=version&apikey=sabkey", got["SABnzbd"].Check)
	assert.Equal("http://plex.example|X-Plex-Token:plextok", got[services.PlexServerName].Check)
	assert.Equal(services.StateUnknown, got["Tautulli"].State)
}

func TestAddAppsSkipsDisabledAndInvalid(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	disabled := extraConfig("Disabled", 0)
	noName := extraConfig("", 0)
	negative := extraConfig("Negative", -time.Second)

	svc := services.New(&services.Config{})
	svc.AddApps(&apps.Apps{
		Lidarr: []apps.Lidarr{
			{StarrApp: starrApp("", "http://lidarr.example", "key", 0)},
			{ExtraConfig: extraConfig("NoURL", 0)},
		},
		Qbit: []apps.Qbit{{
			ExtraConfig: negative,
			URL:         "http://qbit.example",
		}},
		Rtorrent: []apps.Rtorrent{{
			ExtraConfig: negative, URL: "http://rtorrent.example",
		}},
		SabNZB: []apps.SabNZB{{
			SabNZBConfig: apps.SabNZBConfig{
				ExtraConfig: negative,
				Config:      &sabnzbd.Config{URL: "http://sab.example", APIKey: "k"},
			},
		}},
		Transmission: []apps.Xmission{{
			URL: "http://x.example", ExtraConfig: negative,
		}},
		Tautulli: apps.Tautulli{
			TautulliConfig: apps.TautulliConfig{
				ExtraConfig: noName,
				Config:      tautulli.Config{URL: "http://tautulli.example", APIKey: "k"},
			},
		},
		Plex: apps.Plex{
			PlexConfig: apps.PlexConfig{
				Config:      plex.Config{URL: "http://plex.example", Token: "tok"},
				ExtraConfig: apps.ExtraConfig{Interval: cnfg.Duration{Duration: -time.Second}},
			},
		},
		Deluge: []apps.Deluge{{ExtraConfig: disabled}},
	}, []snapshot.MySQLConfig{
		{Name: "empty-host", Host: ""},
		{Name: "neg-timeout", Host: "db.example", Timeout: cnfg.Duration{Duration: -time.Second}},
		{Name: "", Host: "db.example"},
		{Name: "unix", Host: "@unix(/tmp/mysql.sock)"},
		{Name: "empty-tcp", Host: "@tcp()"},
	})

	assert.Zero(svc.SvcCount(), "disabled, unnamed, negative-interval, and invalid mysql hosts must be skipped")
}

func TestAddAppsMySQLAlreadyHasPort(t *testing.T) {
	t.Parallel()

	svc := services.New(&services.Config{})
	svc.AddApps(&apps.Apps{}, []snapshot.MySQLConfig{
		{Name: "custom", Host: "db.example:3310", Timeout: cnfg.Duration{Duration: time.Second}},
	})

	got := resultsByName(svc.GetResults())
	require.Contains(t, got, "custom")
	assert.Equal(t, "db.example:3310", got["custom"].Check)
	assert.Equal(t, services.CheckTCP, got["custom"].Type)
}
