package update //nolint:testpackage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
}

func TestGetUnstableNonJSONBody(t *testing.T) {
	t.Parallel()

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
