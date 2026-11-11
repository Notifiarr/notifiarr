package website //nolint:testpackage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/stretchr/testify/require"
	"golift.io/cnfg"
)

func TestUnmarshalResponseUnauthorized(t *testing.T) {
	t.Parallel()

	body := io.NopCloser(strings.NewReader(`{"result":"error"}`))
	_, err := unmarshalResponse("https://notifiarr.com/api", http.StatusUnauthorized, body)

	if !errors.Is(err, ErrInvalidAPIKey) {
		t.Fatalf("401 should be ErrInvalidAPIKey, got: %v", err)
	}

	if !errors.Is(err, ErrNon200) {
		t.Fatalf("401 should also wrap ErrNon200, got: %v", err)
	}
}

func TestCompressBzip2SkipsSmallAndDisabled(t *testing.T) {
	t.Parallel()

	srv := &server{config: &Config{}}
	body := []byte("small")

	got, compressed, err := srv.compressBzip2(body)
	if err != nil {
		t.Fatalf("small payload: %v", err)
	}

	if compressed || !bytes.Equal(got.Bytes(), body) {
		t.Fatalf("small payload should be sent unchanged, compressed=%v body=%q", compressed, got.Bytes())
	}

	srv.config.NoCompress = true
	body = bytes.Repeat([]byte("a"), 2000)

	got, compressed, err = srv.compressBzip2(body)
	if err != nil {
		t.Fatalf("disabled compression: %v", err)
	}

	if compressed || !bytes.Equal(got.Bytes(), body) {
		t.Fatalf("disabled compression should be sent unchanged, compressed=%v len=%d", compressed, got.Len())
	}
}

func TestCompressBzip2WritesBzipStream(t *testing.T) {
	t.Parallel()

	srv := &server{config: &Config{}}
	body := bytes.Repeat([]byte(`{"name":"radarr"}`), 80)

	got, compressed, err := srv.compressBzip2(body)
	if err != nil {
		t.Fatalf("compress: %v", err)
	}

	if !compressed || !bytes.HasPrefix(got.Bytes(), []byte("BZh")) {
		t.Fatalf("expected a bzip2 stream, compressed=%v prefix=%q", compressed, got.Bytes()[:min(4, got.Len())])
	}
}

func TestCompressBzip2PanicSavesPayloadAndPostsRaw(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	origLog := mnd.Log
	mnd.Log = pathLog{path: filepath.Join(dir, "debug.log")}
	t.Cleanup(func() { mnd.Log = origLog })

	payload := bytes.Repeat([]byte("payload-bytes-"), 100)
	var posted []byte
	var encoding string

	httpSrv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		encoding = req.Header.Get("Content-Encoding")
		posted, _ = io.ReadAll(req.Body)
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(writer, `{"result":"success"}`)
	}))
	t.Cleanup(httpSrv.Close)

	srv := &server{
		config: &Config{
			Apps:    &apps.Apps{APIKey: "test-key"},
			Timeout: cnfg.Duration{Duration: time.Minute},
		},
		client:    &httpClient{Client: httpSrv.Client()},
		bzipWrite: func(io.Writer, []byte) error { panic("index out of range [2251799813686761]") },
		sendData:  make(chan *Request, 1),
	}

	code, size, body, err := srv.sendJSON(context.Background(), httpSrv.URL, bytes.NewBuffer(payload), false)
	require.NoError(t, err)

	t.Cleanup(func() { body.Close() })
	require.Equal(t, http.StatusOK, code)
	require.Empty(t, encoding)
	require.Equal(t, payload, posted)
	require.Equal(t, len(payload), size)

	matches, err := filepath.Glob(filepath.Join(dir, "bzip2-panic-*.json"))
	require.NoError(t, err)
	require.Len(t, matches, 1)

	saved, err := os.ReadFile(matches[0])
	require.NoError(t, err)
	require.Equal(t, payload, saved)

	// Windows only honors the read-only bit, so the Unix mode is not visible there.
	if runtime.GOOS != "windows" {
		info, err := os.Stat(matches[0])
		require.NoError(t, err)
		require.Equal(t, os.FileMode(mnd.Mode0600), info.Mode().Perm())
	}
}

// pathLog is a website-test logger that points crash dumps at a directory.
type pathLog struct {
	path string
}

func (pathLog) Trace(string, ...any) string          { return "id" }
func (pathLog) Print(string, ...any)                 {}
func (pathLog) Printf(string, string, ...any)        {}
func (pathLog) Error(string, ...any)                 {}
func (pathLog) Errorf(string, string, ...any)        {}
func (pathLog) ErrorfNoShare(string, string, ...any) {}
func (pathLog) Debug(string, ...any)                 {}
func (pathLog) Debugf(string, string, ...any)        {}
func (pathLog) DebugEnabled() bool                   { return false }
func (p pathLog) DebugLogPath() string               { return p.path }
func (pathLog) CapturePanic()                        {}
