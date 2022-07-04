package exp

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

// LoggingRoundTripper allows us to use a datacounter to log http request data.
type LoggingRoundTripper struct {
	next http.RoundTripper
	app  string
}

// NewLoggingRounNewMetricsRoundTripperdTripper returns a round tripper to log requests counts and response sizes.
func NewMetricsRoundTripper(app string, next http.RoundTripper) *LoggingRoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}

	if app == "" {
		panic("round trip wrapper app may not be empty")
	}

	return &LoggingRoundTripper{
		next: next,
		app:  app,
	}
}

// RoundTrip satisfies the http.RoundTripper interface.
// This is where our logging takes place.
func (rt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.next.RoundTrip(req)
	if err != nil {
		Apps.Add(rt.app+"&&"+req.Method+" Errors", 1)
	}

	Apps.Add(rt.app+"&&"+req.Method+" Requests", 1)

	if resp == nil || resp.Body == nil {
		return resp, err //nolint:wrapcheck
	}

	defer resp.Body.Close()

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)
	size, _ := io.Copy(ioutil.Discard, tee)
	resp.Body = io.NopCloser(&buf)

	Apps.Add(rt.app+"&&Bytes Received", size)

	return resp, err //nolint:wrapcheck
}
