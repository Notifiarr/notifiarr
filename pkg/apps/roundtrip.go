package apps

import (
	"io"
	"net/http"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"golift.io/datacounter"
	"golift.io/version"
)

/*
 * The code in this files powers the metrics collection for pretty much every integrated app,
 * but only when debug is disabled.
 */

type fakeCloser struct {
	App     string
	Method  string
	Rcvd    func() uint64
	CloseFn func() error
	io.Reader
}

func (f *fakeCloser) Close() error {
	defer mnd.Apps.Add(f.App+"&&"+f.Method+mnd.BytesReceived, int64(f.Rcvd()))
	return f.CloseFn()
}

// LoggingRoundTripper allows us to use a data counter to log http request data.
type LoggingRoundTripper struct {
	next http.RoundTripper
	app  string
}

// NewMetricsRoundTripper returns a round tripper to log requests counts and response sizes.
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
	req.Header.Set("User-Agent", mnd.Title+"/"+version.Version+"-"+version.Revision)

	if req.Body != nil {
		sent := datacounter.NewReaderCounter(req.Body)
		req.Body = io.NopCloser(sent)

		defer func() {
			mnd.Apps.Add(rt.app+"&&"+req.Method+mnd.BytesSent, int64(sent.Count()))
		}()
	}

	resp, err := rt.next.RoundTrip(req)
	checkResp(rt.app, req.Method, resp, err)

	if resp == nil || resp.Body == nil {
		return resp, err //nolint:wrapcheck
	}

	resp.Body = NewFakeCloser(rt.app, req.Method, resp.Body)

	return resp, err //nolint:wrapcheck
}

func NewFakeCloser(app, method string, body io.ReadCloser) io.ReadCloser {
	rcvd := datacounter.NewReaderCounter(body)

	return &fakeCloser{
		Method:  method,
		App:     app,
		Rcvd:    rcvd.Count, // This gets added...
		CloseFn: body.Close, // when this gets called.
		Reader:  rcvd,
	}
}

func checkResp(app, method string, resp *http.Response, err error) {
	mnd.Apps.Add(app+"&&"+method+mnd.Requests, 1)

	if resp != nil {
		mnd.Apps.Add(app+"&&"+method+" Response: "+resp.Status, 1)
	}

	if err != nil || resp == nil {
		mnd.Apps.Add(app+"&&"+method+" Request Errors", 1)
	}
}

// metricMakerCallback is used as a callback function from the starr/debuglog package.
// This is used when debug is enabled.
// This does not interact with or use any other methods in this file.
func metricMakerCallback(app string) func(string, string, int, int, error) {
	return func(status, method string, sent, rcvd int, err error) {
		mnd.Apps.Add(app+"&&"+method+mnd.BytesReceived, int64(rcvd))
		mnd.Apps.Add(app+"&&"+method+mnd.Requests, 1)

		if method != "GET" || sent > 0 {
			mnd.Apps.Add(app+"&&"+method+mnd.BytesSent, int64(sent))
		}

		if err != nil {
			mnd.Apps.Add(app+"&&"+method+" Request Errors", 1)
		} else {
			mnd.Apps.Add(app+"&&"+method+" Response: "+status, 1)
		}
	}
}
