package services_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/cnfg"
)

func TestNewAddGetResults(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	require := require.New(t)

	svc := services.New(&services.Config{})
	assert.Zero(svc.SvcCount())
	assert.Empty(svc.GetResults())

	err := svc.Add([]services.ServiceConfig{
		{
			Name:     "web",
			Type:     services.CheckHTTP,
			Value:    "http://example.com/health",
			Tags:     map[string]any{"env": "test"},
			Interval: cnfg.Duration{Duration: time.Second},
			Timeout:  cnfg.Duration{Duration: 50 * time.Millisecond},
		},
		{
			Name:  "db",
			Type:  services.CheckTCP,
			Value: "127.0.0.1:5432",
		},
	})
	require.NoError(err)
	assert.Equal(2, svc.SvcCount())

	got := resultsByName(svc.GetResults())
	require.Contains(got, "web")
	require.Contains(got, "db")
	assert.Equal("http://example.com/health", got["web"].Check)
	assert.Equal("200", got["web"].Expect)
	assert.Equal(services.CheckHTTP, got["web"].Type)
	assert.Equal(services.MinimumCheckInterval, got["web"].IntervalDur)
	assert.Equal("test", got["web"].Metadata["env"])
	assert.Equal("127.0.0.1:5432", got["db"].Check)
	assert.Equal(services.CheckTCP, got["db"].Type)
	assert.Equal(services.StateUnknown, got["web"].State)
}

func TestAddRejectsInvalidAndKeepsPrior(t *testing.T) {
	t.Parallel()

	svc := services.New(&services.Config{})
	err := svc.Add([]services.ServiceConfig{
		{Name: "good", Type: services.CheckTCP, Value: "127.0.0.1:80"},
		{Name: "bad", Type: services.CheckTCP, Value: "no-port"},
	})
	require.ErrorIs(t, err, services.ErrBadTCP)
	assert.Equal(t, 1, svc.SvcCount(), "the valid service added before the error is kept")
}

func TestAddOverwritesSameName(t *testing.T) {
	t.Parallel()

	svc := services.New(&services.Config{})
	require.NoError(t, svc.Add([]services.ServiceConfig{
		{Name: "svc", Type: services.CheckTCP, Value: "127.0.0.1:80"},
		{Name: "svc", Type: services.CheckTCP, Value: "127.0.0.1:443"},
	}))

	got := resultsByName(svc.GetResults())
	require.Len(t, got, 1)
	assert.Equal(t, "127.0.0.1:443", got["svc"].Check)
}

func TestStoppedCheckerGuards(t *testing.T) {
	t.Parallel()

	svc := services.New(&services.Config{})
	assert.False(t, svc.Running())

	require.NoError(t, svc.Add([]services.ServiceConfig{
		{Name: "web", Type: services.CheckHTTP, Value: "http://example.com"},
	}))

	err := svc.RunCheck(t.Context(), website.EventAPI, "web")
	require.ErrorIs(t, err, services.ErrSvcsStopped)

	err = svc.RunCheck(t.Context(), website.EventAPI, "missing")
	require.ErrorIs(t, err, services.ErrSvcsStopped)

	svc.RunChecks(&common.ActionInput{Type: website.EventAPI, ReqID: "test"})
	svc.Pause()
	svc.Resume()
	svc.Stop()
}

func TestDisabledCheckerLifecycle(t *testing.T) {
	t.Parallel()

	svc := services.New(&services.Config{Disabled: true})
	svc.Start(t.Context(), "Theater")
	t.Cleanup(svc.Stop)

	assert.False(t, svc.Running(), "disabled checkers start paused")

	svc.Resume()
	assert.True(t, svc.Running())
	svc.Pause()
	assert.False(t, svc.Running())

	require.NoError(t, svc.Add([]services.ServiceConfig{
		{
			Name:  "listed",
			Type:  services.CheckTCP,
			Value: "127.0.0.1:9",
			Tags:  map[string]any{"role": "db"},
		},
	}))

	listReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/services/list", nil)
	listReq = mux.SetURLVars(listReq, map[string]string{"action": "list"})
	code, body := svc.APIHandler(listReq)
	assert.Equal(t, http.StatusOK, code)

	results, ok := body.([]*services.CheckResult)
	require.True(t, ok)
	got := resultsByName(results)
	require.Contains(t, got, "listed")
	assert.Equal(t, "127.0.0.1:9", got["listed"].Check)

	badReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/services/nope", nil)
	badReq = mux.SetURLVars(badReq, map[string]string{"action": "nope"})
	code, body = svc.APIHandler(badReq)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "unknown service action: nope", body)

	svc.RunChecks(&common.ActionInput{Type: "log", ReqID: "test"})

	raw, err := json.Marshal(got["listed"])
	require.NoError(t, err)
	assert.NotEmpty(t, raw)
}

func TestStopCancelsInFlightHTTP(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		close(started)
		<-req.Context().Done()
	}))
	t.Cleanup(server.Close)

	svc := services.New(&services.Config{Disabled: true})
	require.NoError(t, svc.Add([]services.ServiceConfig{{
		Name:     "slow",
		Type:     services.CheckHTTP,
		Value:    server.URL,
		Timeout:  cnfg.Duration{Duration: 10 * time.Second},
		Interval: cnfg.Duration{Duration: time.Minute},
	}}))

	svc.Start(t.Context(), "")
	require.NoError(t, svc.RunCheck(t.Context(), website.EventAPI, "slow"))

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("http check did not start")
	}

	begin := time.Now()
	svc.Stop()
	assert.Less(t, time.Since(begin), 2*time.Second, "Stop should cancel the in-flight HTTP check")

	got := resultsByName(svc.GetResults())
	require.Contains(t, got, "slow")
	assert.Equal(t, services.StateUnknown, got["slow"].State, "canceled checks must not persist Critical")

	svc.Stop() // no-op Stop must not hang or log as a fresh shutdown
}

func waitFinished(t *testing.T, done <-chan struct{}) {
	t.Helper()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for service checker lifecycle")
	}
}

// waitUntilCheckerLive blocks until Start has published checker channels.
func waitUntilCheckerLive(t *testing.T, svc *services.Services, name string) {
	t.Helper()

	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		if !errors.Is(svc.RunCheck(t.Context(), website.EventAPI, name), services.ErrSvcsStopped) {
			return
		}

		select {
		case <-timer.C:
			t.Fatal("timed out waiting for Start to enter checker lifecycle")
		case <-ticker.C:
		}
	}
}

// runUntilStopped runs work until Stop returns, after at least one call has finished.
func runUntilStopped(t *testing.T, stop func(), work func()) {
	t.Helper()

	var stopping, signaled atomic.Bool

	ready := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			work()

			if signaled.CompareAndSwap(false, true) {
				close(ready)
			}

			if stopping.Load() {
				return
			}
		}
	}()

	select {
	case <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for producer to become active")
	}

	stop()
	stopping.Store(true)
	waitFinished(t, done)
}

func TestRunningAndStopDoNotDeadlock(t *testing.T) {
	t.Parallel()

	svc := services.New(&services.Config{Disabled: true})
	svc.Start(t.Context(), "")

	runUntilStopped(t, svc.Stop, func() {
		_ = svc.Running()
	})
}

func TestRunCheckAndStopDoNotDeadlock(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	svc := services.New(&services.Config{Disabled: true})
	require.NoError(t, svc.Add([]services.ServiceConfig{{
		Name:     "web",
		Type:     services.CheckHTTP,
		Value:    server.URL,
		Timeout:  cnfg.Duration{Duration: time.Second},
		Interval: cnfg.Duration{Duration: time.Minute},
	}}))
	svc.Start(t.Context(), "")

	runUntilStopped(t, svc.Stop, func() {
		_ = svc.RunCheck(t.Context(), website.EventAPI, "web")
		svc.RunChecks(&common.ActionInput{Type: "log", ReqID: "race"})
	})
}

func TestAddAfterEmptyStartCanRunCheck(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		close(started)
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	svc := services.New(&services.Config{Disabled: true})
	svc.Start(t.Context(), "")
	t.Cleanup(svc.Stop)

	require.NoError(t, svc.Add([]services.ServiceConfig{{
		Name:     "late",
		Type:     services.CheckHTTP,
		Value:    server.URL,
		Timeout:  cnfg.Duration{Duration: time.Second},
		Interval: cnfg.Duration{Duration: time.Minute},
	}}))
	require.NoError(t, svc.RunCheck(t.Context(), website.EventAPI, "late"))

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("check queued after empty Start never ran; worker pool was empty")
	}
}

func TestConcurrentStartAndStop(t *testing.T) {
	t.Parallel()

	for range 10 {
		svc := services.New(&services.Config{Disabled: true})
		require.NoError(t, svc.Add([]services.ServiceConfig{{
			Name:    "web",
			Type:    services.CheckTCP,
			Value:   "127.0.0.1:9",
			Timeout: cnfg.Duration{Duration: 50 * time.Millisecond},
		}}))

		done := make(chan struct{})
		go func() {
			defer close(done)
			svc.Start(t.Context(), "")
		}()

		waitUntilCheckerLive(t, svc, "web")
		svc.Stop()
		waitFinished(t, done)
		svc.Stop()
	}
}
