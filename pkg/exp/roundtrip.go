package exp

import (
	"io"
	"io/ioutil"
	"net/http"

	"golift.io/datacounter"
)

/* The code in thie files powers the metrics collection for prety much every integrated app. */

type fakeCloser struct {
	App     string
	Method  string
	Rcvd    func() uint64
	CloseFn func() error
	io.Reader
}

func (f *fakeCloser) Close() error {
	defer Apps.Add(f.App+"&&"+f.Method+" Bytes Received", int64(f.Rcvd()))
	return f.CloseFn()
}

// LoggingRoundTripper allows us to use a datacounter to log http request data.
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
	if req.Body != nil {
		sent := datacounter.NewReaderCounter(req.Body)
		req.Body = ioutil.NopCloser(sent)

		defer Apps.Add(rt.app+"&&"+req.Method+" Bytes Sent", int64(sent.Count()))
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
	Apps.Add(app+"&&"+method+" Requests", 1)

	if resp != nil {
		Apps.Add(app+"&&"+method+" Response: "+resp.Status, 1)
	}

	if err != nil || resp == nil {
		Apps.Add(app+"&&"+method+" Request Errors", 1)
	}
}
