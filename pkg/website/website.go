package website

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/private"
	"golift.io/datacounter"
	"golift.io/version"
)

// site can be used to call the website APIs.
var site *server

// httpClient is our custom http client to wrap Do and provide retries.
type httpClient struct {
	Retries int
	Client  *http.Client
}

// Do performs an http Request with retries and logging!
func (h *httpClient) Do(req *http.Request) (*http.Response, error) { //nolint:cyclop
	reqID := mnd.Log.Trace(mnd.GetID(req.Context()), "start: httpClient.Do", req.URL.String())
	defer mnd.Log.Trace(reqID, "end: httpClient.Do", req.URL.String())

	req.Header.Set("User-Agent", fmt.Sprintf("%s v%s-%s %s", mnd.Title, version.Version, version.Revision, version.Branch))

	deadline, ok := req.Context().Deadline()
	if !ok {
		deadline = time.Now().Add(h.Client.Timeout)
	}

	timeout := time.Until(deadline).Round(time.Millisecond)

	for retry := 0; ; retry++ {
		mnd.Website.Add(req.Method+mnd.Requests, 1)

		resp, err := h.Client.Do(req)
		if err == nil {
			for i, c := range resp.Cookies() {
				mnd.Log.ErrorfNoShare(reqID, "Unexpected cookie [%v/%v] returned from website: %s",
					reqID, i+1, len(resp.Cookies()), c.String())
			}

			if resp.StatusCode < http.StatusInternalServerError &&
				(resp.StatusCode != http.StatusBadRequest || resp.Header.Get("Content-Type") != "text/html") {
				mnd.Website.Add(req.Method+mnd.BytesSent, resp.Request.ContentLength)
				return resp, nil
			}

			// resp.StatusCode is 500 or higher, make that en error.
			// or resp.StatusCode is 400 and content-type is text/html (cloudflare error).
			size, _ := io.Copy(io.Discard, resp.Body) // must read the entire body when err == nil
			resp.Body.Close()                         // do not defer, because we're in a loop.
			mnd.Website.Add(req.Method+" Retries", 1)
			mnd.Website.Add(req.Method+mnd.BytesSent, resp.Request.ContentLength)
			mnd.Website.Add(req.Method+mnd.BytesReceived, size)
			// shoehorn a non-200 error into the empty http error.
			err = fmt.Errorf("%w: %s: %d bytes, %s", ErrNon200, req.URL, size, resp.Status)
		}

		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			if retry == 0 {
				return resp, fmt.Errorf("website req timed out after %s: %s: %w", timeout, req.URL, err)
			}

			return resp, fmt.Errorf("[%d/%d] website req timed out after %s, giving up: %w",
				retry+1, h.Retries+1, timeout, err)
		case retry == h.Retries:
			return resp, fmt.Errorf("[%d/%d] website req failed: %w", retry+1, h.Retries+1, err)
		default:
			mnd.Log.ErrorfNoShare(reqID, "[%d/%d] website req failed, retrying in %s, error: %v",
				retry+1, h.Retries+1, RetryDelay, err)
			time.Sleep(RetryDelay)
		}
	}
}

// sendAndLogRequest sends a request to the website and logs the result.
func (s *server) sendAndLogRequest(ctx context.Context, data *Request) {
	mnd.Log.Trace(data.ReqID, "start: sendAndLogRequest", data.Route)
	defer mnd.Log.Trace(data.ReqID, "end: sendAndLogRequest", data.Route)

	switch resp, elapsed, err := s.sendRequest(ctx, data); {
	case data.LogMsg == "", errors.Is(err, ErrInvalidAPIKey):
		return
	case err != nil:
		mnd.Log.ErrorfNoShare(data.ReqID, "[%s requested] Sending (%v, buf=%d/%d): %s: %v%s",
			data.Event, elapsed, len(s.sendData), cap(s.sendData), data.LogMsg, err, resp)
	case !data.ErrorsOnly:
		mnd.Log.Printf(data.ReqID, "[%s requested] Sent %s (%v, buf=%d/%d): %s%s",
			data.Event, mnd.FormatBytes(resp.sent), elapsed, len(s.sendData), cap(s.sendData), data.LogMsg, resp)
	}
}

// sendRequest sends a request to the website and returns the result.
func (s *server) sendRequest(ctx context.Context, data *Request) (*Response, time.Duration, error) {
	if len(s.config.Apps.APIKey) != APIKeyLength {
		err := fmt.Errorf("%w: length must be %d characters", ErrInvalidAPIKey, APIKeyLength)
		if data.respChan != nil {
			data.respChan <- &chResponse{
				Response: nil,
				Elapsed:  0,
				Error:    err,
			}
		}

		return nil, 0, err
	}

	var uri string

	if len(data.Params) > 0 {
		uri = data.Route.Path(data.Event, data.Params...)
	} else {
		uri = data.Route.Path(data.Event)
	}

	var (
		resp  *Response
		err   error
		start = time.Now()
	)

	if data.UploadFile != nil {
		resp, err = s.sendFile(ctx, uri, data.UploadFile)
	} else {
		resp, err = s.sendPayload(ctx, uri, data.Payload, data.LogPayload)
	}

	elapsed := time.Since(start).Round(time.Millisecond)

	if data.respChan != nil {
		data.respChan <- &chResponse{
			Response: resp,
			Elapsed:  elapsed,
			Error:    err,
		}
	}

	return resp, elapsed, err
}

// sendPayload sends a JSON payload to the website and returns the result.
func (s *server) sendPayload(ctx context.Context, uri string, payload any, log bool) (*Response, error) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: sendPayload", uri)
	defer mnd.Log.Trace(reqID, "end: sendPayload", uri)

	data, err := json.Marshal(payload)
	if err == nil {
		var torn map[string]any
		if err := json.Unmarshal(data, &torn); err == nil {
			if torn["host"], err = GetHostInfo(ctx); err != nil {
				mnd.Log.Errorf(mnd.GetID(ctx), "Host Info Unknown: %v", err)
			}

			torn["private"] = private.Info()
			payload = torn
		}
	}

	var post []byte

	if log {
		post, err = json.MarshalIndent(payload, "", " ")
	} else {
		post, err = json.Marshal(payload)
	}

	if err != nil {
		return nil, fmt.Errorf("encoding data to JSON (report this bug please): %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, s.getTimeout())
	defer cancel()

	code, body, err := s.sendJSON(ctx, BaseURL+uri, post, log)
	if err != nil {
		return nil, err
	}

	resp, err := unmarshalResponse(BaseURL+uri, code, body)
	if resp != nil {
		resp.sent = len(post)
	}

	return resp, err
}

// sendJSON posts a JSON payload to a URL. Returns the response body or an error.
func (s *server) sendJSON(ctx context.Context, url string, data []byte, log bool) (int, io.ReadCloser, error) {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: sendJSON", url)
	defer mnd.Log.Trace(reqID, "end: sendJSON", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return 0, nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("Content-Type", mnd.ContentTypeJSON)
	req.Header.Set("X-Api-Key", s.config.Apps.APIKey)

	start := time.Now()

	resp, err := s.client.Do(req)
	if err != nil {
		s.debughttplog(reqID, nil, url, start, string(data), nil)
		return 0, nil, fmt.Errorf("making http request: %w", err)
	}

	if !mnd.Log.DebugEnabled() { // no debug, just return the body.
		return resp.StatusCode, resp.Body, nil
	}

	return resp.StatusCode, s.debugLogResponseBody(reqID, start, resp, url, data, log), nil
}

func (s *server) getTimeout() time.Duration {
	timeout := s.config.Timeout.Duration
	if timeout > MaxTimeout {
		timeout = MaxTimeout
	} else if timeout < MinTimeout {
		timeout = MinTimeout
	}

	const multiplier = 0.92

	// As the channel fills up, we reduce the timeout proportionally to avoid long waits.
	channelUtilization := float64(len(s.sendData)) / float64(cap(s.sendData))
	// Minimum timeout is 8% of original, maximum is 100% of original
	timeoutMultiplier := 1.0 - (channelUtilization * multiplier)

	return time.Duration(float64(timeout) * timeoutMultiplier)
}

// unmarshalResponse attempts to turn the reply from the website into structured data.
func unmarshalResponse(url string, code int, body io.ReadCloser) (*Response, error) {
	var (
		buf     bytes.Buffer
		resp    Response
		counter = datacounter.NewReaderCounter(body)
	)

	defer func() {
		resp.size = int64(counter.Count())
		mnd.Website.Add("POST"+mnd.BytesReceived, resp.size)
		body.Close()
	}()

	err := json.NewDecoder(io.TeeReader(counter, &buf)).Decode(&resp)
	if code < http.StatusOK || code > http.StatusIMUsed {
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %d %s (unmarshal error: %v), body: %s",
				ErrNon200, url, code, http.StatusText(code), err, buf.String()) //nolint:errorlint
		}

		return &resp, fmt.Errorf("%w: %s: %d %s", ErrNon200, url, code, http.StatusText(code))
	}

	if err != nil {
		return nil, fmt.Errorf("converting json response: %w, body: %s", err, buf.String())
	}

	return &resp, nil
}

// TestApiKey tests if the API key is valid.
func TestApiKey(ctx context.Context, apiKey string) error {
	reqID := mnd.Log.Trace(mnd.GetID(ctx), "start: TestApiKey")
	defer mnd.Log.Trace(reqID, "end: TestApiKey")

	if site == nil {
		return ErrNoChannel // this will never happen, but we'll be safe.
	}

	path := BaseURL + ValidateRoute.Path(EventUser)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return fmt.Errorf("creating notifiarr.com request: %w", err)
	}

	req.Header.Set("X-Api-Key", apiKey)

	resp, err := site.client.Do(req)
	if err != nil {
		return fmt.Errorf("making notifiarr.com request: %w", err)
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrInvalidAPIKey, resp.Status)
	}

	return nil
}
