package apps //nolint:testpackage // handlers and the version probe are unexported.

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
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

func TestSonarrTagLabel(t *testing.T) {
	t.Parallel()

	require.Equal(t, "my-tag", sonarrTagLabel("My Tag"))
	require.Equal(t, "freeleech-75", sonarrTagLabel("Freeleech_75"))
	require.Equal(t, "already-ok", sonarrTagLabel("ALREADY-ok"))
	require.Equal(t, "a--b", sonarrTagLabel("a__b"))
}

func TestSetupSonarrSelectsAPI(t *testing.T) {
	t.Parallel()

	t.Run("v5", func(t *testing.T) {
		t.Parallel()

		var paths []string

		server := sonarrStatusServer(t, &paths, `{"version":"5.0.0.2742"}`)
		got := setupOneSonarr(t, server.URL)
		require.Equal(t, sonarr.APIv5, got.APIVersion())
		require.Equal(t, "5.0.0.2742", got.Version())

		_, err := got.GetQualityProfilesContext(context.Background())
		require.NoError(t, err)
		require.Equal(t, []string{"/api/v3/system/status", "/api/v5/qualityProfile"}, paths)
	})

	t.Run("v4", func(t *testing.T) {
		t.Parallel()

		var paths []string

		server := sonarrStatusServer(t, &paths, `{"version":"4.0.20.3014"}`)
		got := setupOneSonarr(t, server.URL)
		require.Equal(t, sonarr.APIver, got.APIVersion())
		require.Equal(t, "4.0.20.3014", got.Version())

		_, err := got.GetQualityProfilesContext(context.Background())
		require.NoError(t, err)
		require.Equal(t, []string{"/api/v3/system/status", "/api/v3/qualityProfile"}, paths)
	})

	t.Run("probe failure stays on v3", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(`{"message":"nope"}`))
		}))
		t.Cleanup(server.Close)

		got := setupOneSonarr(t, server.URL)
		require.Equal(t, sonarr.APIver, got.APIVersion())
		require.Empty(t, got.Version())
	})
}

func TestSonarrUpdateSeriesKeepsLanguage(t *testing.T) {
	t.Parallel()

	var got string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/api/v3/series/9":
			_, _ = writer.Write([]byte(`{"id":9,"language":{"id":2,"name":"French"},"seasonType":"anime"}`))
		case req.Method == http.MethodPut && req.URL.Path == "/api/v3/series/9":
			raw, _ := io.ReadAll(req.Body)
			got = string(raw)
			_, _ = writer.Write([]byte(`{"id":9}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	req := sonarrRequest(t, server.URL, `{"id":9,"title":"Show"}`, nil)
	code, msg := sonarrUpdateSeries(req)
	require.Equal(t, http.StatusOK, code)
	require.Contains(t, got, `"language":{"id":2,"name":"French"}`)
	require.Contains(t, got, `"seasonType":"anime"`)
	require.Equal(t, "sonarr seems to have worked", msg)
}

func TestSonarrUpdateSeriesKeepsProvidedLanguage(t *testing.T) {
	t.Parallel()

	gets := 0

	var got string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			gets++
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}

		raw, _ := io.ReadAll(req.Body)
		got = string(raw)
		_, _ = writer.Write([]byte(`{"id":9}`))
	}))
	t.Cleanup(server.Close)

	body := `{"id":9,"title":"Show","seasonType":"standard","language":{"id":4,"name":"German"}}`
	req := sonarrRequest(t, server.URL, body, nil)
	code, _ := sonarrUpdateSeries(req)
	require.Equal(t, http.StatusOK, code)
	require.Zero(t, gets)
	require.Contains(t, got, `"seasonType":"standard"`)
	require.Contains(t, got, `"name":"German"`)
}

func TestSonarrLangProfilesV5(t *testing.T) {
	t.Parallel()

	client := sonarr.NewV5(starr.New("k", "http://127.0.0.1:1", time.Second))
	ctx := context.WithValue(context.Background(), starr.Sonarr, Sonarr{Sonarr: client})
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/languageProfiles", nil)

	code, got := sonarrLangProfiles(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, map[int64]string{}, got)
}

func TestSonarrSetTagNormalizesLabel(t *testing.T) {
	t.Parallel()

	var got string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		raw, _ := io.ReadAll(req.Body)
		got = string(raw)
		_, _ = writer.Write([]byte(`{"id":3,"label":"my-tag"}`))
	}))
	t.Cleanup(server.Close)

	req := sonarrRequest(t, server.URL, "", map[string]string{"label": "My Tag"})
	code, id := sonarrSetTag(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, 3, id)
	require.Contains(t, got, `"label":"my-tag"`)
}

func TestSonarrQualityProfileKeepsItemSizes(t *testing.T) {
	t.Parallel()

	var (
		got  string
		path string
	)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		path = req.URL.Path

		raw, _ := io.ReadAll(req.Body)
		got = string(raw)
		_, _ = writer.Write([]byte(`{"id":3}`))
	}))
	t.Cleanup(server.Close)

	body := `{"name":"Any","items":[{"name":"HDTV-720p","allowed":true,"minSize":3.5,"maxSize":44,"preferredSize":0}]}`
	req := sonarrRequest(t, server.URL, body, map[string]string{"profileID": "3"})
	code, msg := sonarrUpdateQualityProfile(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "OK", msg)
	require.Equal(t, "/api/v3/qualityProfile/3", path)
	require.Contains(t, got, `"minSize":3.5`)
	require.Contains(t, got, `"maxSize":44`)
	require.Contains(t, got, `"preferredSize":0`)
}

func sonarrStatusServer(t *testing.T, paths *[]string, status string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		*paths = append(*paths, req.URL.Path)

		switch req.URL.Path {
		case "/api/v3/system/status":
			_, _ = writer.Write([]byte(status))
		case "/api/v3/qualityProfile", "/api/v5/qualityProfile":
			_, _ = writer.Write([]byte(`[]`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	return server
}

func setupOneSonarr(t *testing.T, appURL string) Sonarr {
	t.Helper()

	cfg := &AppsConfig{Sonarr: []StarrConfig{{
		URL:     appURL,
		APIKey:  "k",
		Name:    "TV",
		Timeout: cnfg.Duration{Duration: time.Second},
	}}}
	got, err := cfg.setupSonarr(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)

	return got[0]
}

func sonarrRequest(t *testing.T, appURL, body string, vars map[string]string) *http.Request {
	t.Helper()

	client := sonarr.New(&starr.Config{URL: appURL, APIKey: "k"})
	ctx := context.WithValue(context.Background(), starr.Sonarr, Sonarr{Sonarr: client})
	req := httptest.NewRequestWithContext(ctx, http.MethodPut, "/", strings.NewReader(body))

	if vars != nil {
		req = mux.SetURLVars(req, vars)
	}

	return req
}
