package clientinfo //nolint:testpackage // the check-in builder and version poll are unexported.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/stretchr/testify/require"
	"golift.io/cnfg"
	"golift.io/starr"
	"golift.io/starr/sonarr"
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
func (nopLog) CapturePanic()                        {}

func TestMain(m *testing.M) {
	mnd.Log = nopLog{}
	os.Exit(m.Run())
}

func TestSonarrVersionOnCheckIn(t *testing.T) {
	t.Parallel()

	live := apps.Sonarr{Name: "TV"}
	live.SetVersion("5.0.0.2742")

	down := apps.Sonarr{Name: "Down"}
	movies := apps.Radarr{Name: "Movies"}

	info := (&Config{Apps: &apps.Apps{
		Sonarr: []apps.Sonarr{live, down},
		Radarr: []apps.Radarr{movies},
	}}).getAppConfigs(context.Background(), true)

	raw, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded struct {
		Sonarr []map[string]any `json:"sonarr"`
		Radarr []map[string]any `json:"radarr"`
	}

	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Equal(t, "5.0.0.2742", decoded.Sonarr[0]["version"])
	require.Equal(t, "TV", decoded.Sonarr[0]["name"])
	require.NotContains(t, decoded.Sonarr[1], "version")
	require.NotContains(t, decoded.Radarr[0], "version")
}

func TestGetSonarrVersionRefreshesCheckIn(t *testing.T) {
	t.Parallel()

	var path string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		path = req.URL.Path
		_, _ = writer.Write([]byte(`{"version":"5.0.0.2742"}`))
	}))
	t.Cleanup(server.Close)

	cfg := &Config{Apps: &apps.Apps{Sonarr: []apps.Sonarr{testSonarr(server.URL)}}}
	cfg.Sonarr[0].SetVersion("4.0.20.3014")

	var wait sync.WaitGroup

	stats := make([]*SonarrConTest, 1)
	cfg.getSonarrVersion(context.Background(), &wait, cfg.Sonarr, stats)
	wait.Wait()

	require.Equal(t, "/api/v3/system/status", path)
	require.Equal(t, "5.0.0.2742", cfg.Sonarr[0].Version())
	require.Equal(t, sonarr.APIv5, cfg.Sonarr[0].APIVersion())
	require.True(t, stats[0].Up)
	require.Equal(t, "5.0.0.2742", stats[0].Status.Version)
}

func TestGetSonarrVersionSwitchesBackToV3(t *testing.T) {
	t.Parallel()

	var path string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		path = req.URL.Path
		_, _ = writer.Write([]byte(`{"version":"4.0.20.3014"}`))
	}))
	t.Cleanup(server.Close)

	instance := testSonarr(server.URL)
	instance.Sonarr = instance.WithAPI(sonarr.APIv5)
	instance.SetVersion("5.0.0.2742")

	cfg := &Config{Apps: &apps.Apps{Sonarr: []apps.Sonarr{instance}}}

	var wait sync.WaitGroup

	stats := make([]*SonarrConTest, 1)
	cfg.getSonarrVersion(context.Background(), &wait, cfg.Sonarr, stats)
	wait.Wait()

	require.Equal(t, "/api/v5/system/status", path)
	require.Equal(t, "4.0.20.3014", cfg.Sonarr[0].Version())
	require.Equal(t, sonarr.APIver, cfg.Sonarr[0].APIVersion())
	require.Equal(t, sonarr.APIver, instance.APIVersion())
}

func TestGetSonarrVersionKeepsCachedVersionOnFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(`{"message":"nope"}`))
	}))
	t.Cleanup(server.Close)

	instance := testSonarr(server.URL)
	instance.Sonarr = instance.WithAPI(sonarr.APIv5)
	instance.SetVersion("5.0.0.2742")

	cfg := &Config{Apps: &apps.Apps{Sonarr: []apps.Sonarr{instance}}}

	var wait sync.WaitGroup

	stats := make([]*SonarrConTest, 1)
	cfg.getSonarrVersion(context.Background(), &wait, cfg.Sonarr, stats)
	wait.Wait()

	require.Equal(t, "5.0.0.2742", cfg.Sonarr[0].Version())
	require.Equal(t, sonarr.APIv5, cfg.Sonarr[0].APIVersion())
	require.False(t, stats[0].Up)
}

func testSonarr(appURL string) apps.Sonarr {
	config := starr.New("k", appURL, time.Second)

	return apps.Sonarr{
		Config:  *config,
		Name:    "TV",
		Timeout: cnfg.Duration{Duration: time.Second},
		Sonarr:  sonarr.New(config),
	}
}
