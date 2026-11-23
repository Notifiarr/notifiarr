package clientinfo //nolint:testpackage // getAppConfigs is unexported.

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/stretchr/testify/require"
)

type nopLog struct{}

func (nopLog) Trace(string, ...any) string          { return "id" }
func (nopLog) Print(string, ...any)                 {}
func (nopLog) Printf(string, string, ...any)        {}
func (nopLog) Error(string, ...any)                 {}
func (nopLog) Errorf(string, string, ...any)        {}
func (nopLog) ErrorfNoShare(string, string, ...any) {}
func (nopLog) Debug(string, ...any)                 {}
func (nopLog) Debugf(string, string, ...any)        {}
func (nopLog) DebugEnabled() bool                   { return false }
func (nopLog) DebugLogPath() string                 { return "" }
func (nopLog) CapturePanic()                        {}

func TestMain(m *testing.M) {
	mnd.Log = nopLog{}
	os.Exit(m.Run())
}

func TestGetAppConfigsIncludesDownloadersAndMedia(t *testing.T) {
	t.Parallel()

	info := testApps().getAppConfigs(context.Background(), true)
	raw, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	require.Equal(t, []any{
		map[string]any{"instance": float64(1), "name": "Seed"},
		map[string]any{"instance": float64(2), "name": "Ratio"},
	}, decoded["deluge"])
	require.Equal(t, []any{
		map[string]any{"instance": float64(1), "name": "Torrents"},
	}, decoded["qbit"])
	require.Equal(t, []any{
		map[string]any{"instance": float64(1), "name": "Watch"},
	}, decoded["rtorrent"])
	require.Equal(t, []any{
		map[string]any{"instance": float64(1), "name": "Send"},
	}, decoded["transmission"])
	require.Equal(t, []any{
		map[string]any{"instance": float64(1), "name": "News"},
	}, decoded["sabnzbd"])
	require.Equal(t, []any{
		map[string]any{"instance": float64(1), "name": "Usenet"},
	}, decoded["nzbget"])
	require.Equal(t, map[string]any{"instance": float64(1), "name": "Movies"}, decoded["plex"])
	require.Equal(t, map[string]any{"instance": float64(1), "name": "Stats"}, decoded["tautulli"])
	require.Equal(t, []any{
		map[string]any{"instance": float64(1), "name": "Films"},
	}, decoded["radarr"])
	require.Nil(t, decoded["lidarr"])
}

func TestGetAppConfigsKeepsTautulliNameWithUsers(t *testing.T) {
	t.Parallel()

	users := &tautulli.Users{}
	users.Response.Data = []tautulli.User{{
		UserID:       7,
		Username:     "ada",
		FriendlyName: "Ada Lovelace",
	}}
	data.Save("tautulliUsers", users)

	info := testApps().getAppConfigs(context.Background(), false)
	require.Equal(t, 1, info.Tautulli.Instance)
	require.Equal(t, "Stats", info.Tautulli.Name)
	require.Equal(t, "Ada Lovelace", info.Tautulli.Users["7"])
	require.Equal(t, "Ada Lovelace", info.Tautulli.Users["ada"])
}

func TestGetAppConfigsOmitsDisabledPlex(t *testing.T) {
	t.Parallel()

	info := (&Config{Apps: &apps.Apps{
		Deluge: []apps.Deluge{{Name: "Seed"}},
	}}).getAppConfigs(context.Background(), true)

	raw, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.NotContains(t, decoded, "plex")
	require.Nil(t, decoded["tautulli"])
	require.Equal(t, []any{
		map[string]any{"instance": float64(1), "name": "Seed"},
	}, decoded["deluge"])
}

func testApps() *Config {
	return &Config{Apps: &apps.Apps{
		Deluge: []apps.Deluge{
			{Name: "Seed"},
			{Name: "Ratio"},
		},
		Qbit:         []apps.Qbit{{Name: "Torrents"}},
		Rtorrent:     []apps.Rtorrent{{Name: "Watch"}},
		Transmission: []apps.Xmission{{Name: "Send"}},
		SabNZB:       []apps.SabNZB{{Name: "News"}},
		NZBGet:       []apps.NZBGet{{Name: "Usenet"}},
		Radarr:       []apps.Radarr{{Name: "Films"}},
		Plex: apps.Plex{PlexConfig: apps.PlexConfig{
			ExtraConfig: apps.ExtraConfig{Name: "Movies"},
			Config:      plex.Config{URL: "http://plex.local", Token: "token"},
		}},
		Tautulli: apps.Tautulli{TautulliConfig: apps.TautulliConfig{
			ExtraConfig: apps.ExtraConfig{Name: "Stats"},
			Config:      tautulli.Config{URL: "http://tautulli.local", APIKey: "key"},
		}},
	}}
}
