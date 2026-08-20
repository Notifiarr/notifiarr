package services_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/cnfg"
)

func serveHTTP(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return server
}

func checkHTTP(t *testing.T, value, expect string) *services.CheckResult {
	t.Helper()

	return (&services.ServiceConfig{
		Name:    t.Name(),
		Type:    services.CheckHTTP,
		Value:   value,
		Expect:  expect,
		Timeout: cnfg.Duration{Duration: time.Second},
	}).CheckOnly(t.Context())
}

func TestCheckOnlyHTTPOK(t *testing.T) {
	t.Parallel()

	server := serveHTTP(t, http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))

	result := checkHTTP(t, server.URL, "")
	require.Equal(t, services.StateOK, result.State)
	assert.Contains(t, result.Output.String(), "200")
}

func TestCheckOnlyHTTPSecretsAndCodes(t *testing.T) {
	t.Parallel()

	server := serveHTTP(t, http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte("token leaked: supersecret"))
	}))

	unexpected := checkHTTP(t, server.URL+"?token=supersecret", "200")
	require.Equal(t, services.StateCritical, unexpected.State)
	assert.NotContains(t, unexpected.Output.String(), "supersecret")
	assert.Contains(t, unexpected.Output.String(), "********")

	allowed := checkHTTP(t, server.URL, "200,404")
	assert.Equal(t, services.StateOK, allowed.State)
}

func TestCheckOnlyHTTPRedirects(t *testing.T) {
	t.Parallel()

	server := serveHTTP(t, http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		http.Redirect(writer, req, "/ok", http.StatusFound)
	}))

	assert.Equal(t, services.StateCritical, checkHTTP(t, server.URL, "200").State)
	assert.Equal(t, services.StateOK, checkHTTP(t, server.URL, "302").State)
}

func TestCheckOnlyHTTPHeaders(t *testing.T) {
	t.Parallel()

	server := serveHTTP(t, http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.Header.Get("X-Test") != "yes" || req.Host != "custom.example" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}))

	result := checkHTTP(t, server.URL+"|X-Test:yes|Host:custom.example", "")
	assert.Equal(t, services.StateOK, result.State)
}

func TestCheckOnlyHTTPOutputFormatting(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/amp", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte("Tom & Jerry"))
	})
	mux.HandleFunc("/big", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(strings.Repeat("body ", 200)))
	})

	server := serveHTTP(t, mux)

	escaped := checkHTTP(t, server.URL+"/amp", "")
	assert.Equal(t, services.StateCritical, escaped.State)
	assert.Contains(t, escaped.Output.String(), "Tom & Jerry")

	raw, err := json.Marshal(escaped.Output)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `\u0026amp;`)

	truncated := checkHTTP(t, server.URL+"/big", "")
	assert.Equal(t, services.StateCritical, truncated.State)
	assert.LessOrEqual(t, len(truncated.Output.String()), 170)
}

func TestCheckOnlyHTTPErrors(t *testing.T) {
	t.Parallel()

	down := checkHTTP(t, "http://127.0.0.1:1", "")
	assert.Equal(t, services.StateCritical, down.State)
	assert.Contains(t, down.Output.String(), "making request")

	badURL := checkHTTP(t, ":", "")
	assert.Equal(t, services.StateUnknown, badURL.State)
	assert.Contains(t, badURL.Output.String(), "creating request")
}

func TestCheckOnlyHTTPTLS(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	insecure := (&services.ServiceConfig{
		Name:    "tls-skip",
		Type:    services.CheckHTTP,
		Value:   server.URL,
		Timeout: cnfg.Duration{Duration: time.Second},
	}).CheckOnly(t.Context())
	require.Equal(t, services.StateOK, insecure.State)

	strict := (&services.ServiceConfig{
		Name:    "tls-strict",
		Type:    services.CheckHTTP,
		Value:   server.URL,
		Expect:  "200,SSL",
		Timeout: cnfg.Duration{Duration: time.Second},
	}).CheckOnly(t.Context())
	assert.Equal(t, services.StateCritical, strict.State)
	assert.Contains(t, strict.Output.String(), "making request")
}

func TestCheckOnlyTCP(t *testing.T) {
	t.Parallel()

	listenCfg := new(net.ListenConfig)

	listener, err := listenCfg.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}

			_ = conn.Close()
		}
	}()

	connected := (&services.ServiceConfig{
		Name:    "tcp-ok",
		Type:    services.CheckTCP,
		Value:   listener.Addr().String(),
		Timeout: cnfg.Duration{Duration: time.Second},
	}).CheckOnly(t.Context())
	require.Equal(t, services.StateOK, connected.State)
	assert.Contains(t, connected.Output.String(), "connected to port")

	closedPort, err := listenCfg.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := closedPort.Addr().String()
	require.NoError(t, closedPort.Close())

	down := (&services.ServiceConfig{
		Name:    "tcp-down",
		Type:    services.CheckTCP,
		Value:   addr,
		Timeout: cnfg.Duration{Duration: time.Second},
	}).CheckOnly(t.Context())
	assert.Equal(t, services.StateCritical, down.State)
	assert.Contains(t, down.Output.String(), "connection error")
}

func TestCheckOnlyProcess(t *testing.T) {
	t.Parallel()

	missing := (&services.ServiceConfig{
		Name:    "missing",
		Type:    services.CheckPROC,
		Value:   "this-process-definitely-does-not-exist-xyz",
		Timeout: cnfg.Duration{Duration: time.Second},
	}).CheckOnly(t.Context())
	assert.Equal(t, services.StateCritical, missing.State)

	notRunning := (&services.ServiceConfig{
		Name:    "not-running",
		Type:    services.CheckPROC,
		Value:   "this-process-definitely-does-not-exist-xyz",
		Expect:  "running",
		Timeout: cnfg.Duration{Duration: time.Second},
	}).CheckOnly(t.Context())
	assert.Equal(t, services.StateOK, notRunning.State, "Expect=running means the process should be absent")

	procs, err := services.GetAllProcesses(t.Context())
	require.NoError(t, err)

	var needle string

	for _, proc := range procs {
		if proc.CmdLine != "" {
			needle = proc.CmdLine
			const maxNeedle = 12
			if len(needle) > maxNeedle {
				needle = needle[:maxNeedle]
			}

			break
		}
	}

	if needle == "" {
		t.Skip("no process command lines available on this host")
	}

	found := (&services.ServiceConfig{
		Name:    "found",
		Type:    services.CheckPROC,
		Value:   needle,
		Expect:  "count:1",
		Timeout: cnfg.Duration{Duration: 2 * time.Second},
	}).CheckOnly(t.Context())
	assert.Equal(t, services.StateOK, found.State)
	assert.Contains(t, found.Output.String(), "found")
}

func TestCheckOnlyValidateError(t *testing.T) {
	t.Parallel()

	result := (&services.ServiceConfig{
		Type:  services.CheckHTTP,
		Value: "http://example.com",
		Tags:  map[string]any{"site": "test"},
	}).CheckOnly(context.Background())

	assert.Equal(t, services.StateCritical, result.State)
	assert.Contains(t, result.Output.String(), services.ErrNoName.Error())
	assert.Equal(t, "test", result.Metadata["site"])
}

func TestRemoveSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		appURL  string
		message string
		want    string
	}{
		{
			name:    "apikey",
			appURL:  "http://example.com/api?apikey=secret-key",
			message: "failed secret-key lookup",
			want:    "failed ******** lookup",
		},
		{
			name:    "token",
			appURL:  "http://example.com/?token=tok123",
			message: "bad tok123",
			want:    "bad ********",
		},
		{
			name:    "password and pass",
			appURL:  "http://example.com/?password=hunter2&pass=abc",
			message: "hunter2 abc",
			want:    "******** ********",
		},
		{
			name:    "ignores headers after pipe",
			appURL:  "http://example.com/?apikey=keepme|X-Api-Key:other",
			message: "keepme should go",
			want:    "******** should go",
		},
		{
			name:    "invalid url left alone",
			appURL:  "://not a url",
			message: "unchanged",
			want:    "unchanged",
		},
		{
			name:    "no secrets",
			appURL:  "http://example.com/health",
			message: "all good",
			want:    "all good",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, services.RemoveSecrets(test.appURL, test.message))
		})
	}
}

func TestServiceDue(t *testing.T) {
	t.Parallel()

	now := time.Now()
	svc := &services.Service{
		ServiceConfig: &services.ServiceConfig{
			Interval: cnfg.Duration{Duration: time.Minute},
		},
	}

	assert.True(t, svc.Due(now), "never-checked services with an interval are due")

	svc.LastCheck = now.Add(-30 * time.Second)
	assert.False(t, svc.Due(now))

	svc.LastCheck = now.Add(-2 * time.Minute)
	assert.True(t, svc.Due(now))

	svc.Interval.Duration = 0
	assert.False(t, svc.Due(now), "zero interval is never due")
}

func TestCheckOnlyPingCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	cfg := &services.ServiceConfig{
		Name:    "ping",
		Type:    services.CheckPING,
		Value:   "127.0.0.1",
		Expect:  "1:1:200",
		Timeout: cnfg.Duration{Duration: 5 * time.Second},
	}

	done := make(chan *services.CheckResult, 1)

	go func() {
		done <- cfg.CheckOnly(ctx)
	}()

	select {
	case result := <-done:
		require.NotNil(t, result)
	case <-time.After(2 * time.Second):
		t.Fatal("ping check did not honor canceled context")
	}
}

func TestGetAllProcesses(t *testing.T) {
	t.Parallel()

	procs, err := services.GetAllProcesses(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, procs)

	for _, proc := range procs {
		assert.NotZero(t, proc.PID)
	}
}
