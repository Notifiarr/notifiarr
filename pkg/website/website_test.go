package website

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
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
