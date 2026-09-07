package update //nolint:testpackage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/stretchr/testify/require"
)

type captureLog struct {
	mnd.Logger
	msg string
}

func (clog *captureLog) Errorf(_ string, msg string, args ...any) {
	clog.msg = fmt.Sprintf(msg, args...)
}

func withCaptureLog(t *testing.T) *captureLog {
	t.Helper()

	orig := mnd.Log
	clog := &captureLog{}
	mnd.Log = clog
	t.Cleanup(func() { mnd.Log = orig })

	return clog
}

func TestGetUnstableDecodesJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if !strings.HasSuffix(req.URL.Path, ".txt") {
			http.Error(writer, "missing txt suffix", http.StatusNotFound)
			return
		}

		if req.URL.Query().Get("stamp") == "" {
			http.Error(writer, "missing stamp", http.StatusBadRequest)
			return
		}

		writer.Header().Set("Content-Type", "text/plain")
		writer.Header().Set("Last-Modified", "Mon, 07 Sep 2026 00:24:33 GMT")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"version":  "0.9.8",
			"revision": 3416,
			"size":     14075099,
		})
	}))
	t.Cleanup(srv.Close)

	got, err := GetUnstable(context.Background(), srv.URL+"/notifiarr.amd64.exe.zip")
	require.NoError(t, err)
	require.Equal(t, "0.9.8", got.Ver)
	require.Equal(t, 3416, got.Rev)
	require.Equal(t, int64(14075099), got.Size)

	wantTime, err := time.Parse(http.TimeFormat, "Mon, 07 Sep 2026 00:24:33 GMT")
	require.NoError(t, err)
	require.True(t, got.Time.Equal(wantTime))
}

func TestGetUnstableNonJSONBody(t *testing.T) { //nolint:paralleltest
	clog := withCaptureLog(t)

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("error code: 522"))
	}))
	t.Cleanup(srv.Close)

	_, err := GetUnstable(context.Background(), srv.URL+"/notifiarr.amd64.exe.zip")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decoding")
	require.Contains(t, err.Error(), "invalid character 'e'")
	require.Contains(t, clog.msg, `status 200`)
	require.Contains(t, clog.msg, `type "text/plain"`)
	require.Contains(t, clog.msg, `body "error code: 522"`)
}

func TestGetUnstableNonOKStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte("error code: 522"))
	}))
	t.Cleanup(srv.Close)

	_, err := GetUnstable(context.Background(), srv.URL+"/notifiarr.amd64.exe.zip")
	require.ErrorIs(t, err, ErrBadStatus)
	require.Contains(t, err.Error(), "502")
}

func TestGetUnstableBodyTooLarge(t *testing.T) { //nolint:paralleltest
	clog := withCaptureLog(t)

	const uniqueTail = "UNIQUE_TAIL"

	body := strings.Repeat("A", maxLogBody) + uniqueTail

	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	_, err := GetUnstable(context.Background(), srv.URL+"/notifiarr.amd64.exe.zip")
	require.ErrorIs(t, err, ErrBodyTooLarge)
	require.Contains(t, clog.msg, `status 200`)
	require.Contains(t, clog.msg, `type "text/plain"`)
	require.Contains(t, clog.msg, strings.Repeat("A", maxLogBody))
	require.NotContains(t, clog.msg, uniqueTail)
}
