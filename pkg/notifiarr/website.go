package notifiarr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/miolini/datacounter"
)

//nolint:gochecknoglobals
var websiteMap = exp.GetMap("notifiarr.com").Init()

// httpClient is our custom http client to wrap Do and provide retries.
type httpClient struct {
	Retries int
	*log.Logger
	*http.Client
}

// sendPlexMeta is kicked off by the webserver in go routine.
// It's also called by the plex cron (with webhook set to nil).
// This runs after Plex drops off a webhook telling us someone did something.
// This gathers cpu/ram, and waits 10 seconds, then grabs plex sessions.
// It's all POSTed to notifiarr. May be used with a nil Webhook.
func (c *Config) sendPlexMeta(event EventType, hook *plexIncomingWebhook, wait bool) (*Response, error) {
	extra := time.Second
	if wait {
		extra = c.Plex.Delay.Duration
	}

	ctx, cancel := context.WithTimeout(context.Background(), extra+c.Snap.Timeout.Duration)
	defer cancel()

	var (
		payload = &Payload{Load: hook, Plex: &plex.Sessions{Name: c.Plex.Name}}
		wg      sync.WaitGroup
	)

	rep := make(chan error)
	defer close(rep)

	go func() {
		for err := range rep {
			if err != nil {
				c.Errorf("Building Metadata: %v", err)
			}
		}
	}()

	wg.Add(1)

	go func() {
		payload.Snap = c.getMetaSnap(ctx)
		wg.Done() // nolint:wsl
	}()

	if !wait || !c.Plex.NoActivity {
		var err error
		if payload.Plex, err = c.GetSessions(wait); err != nil {
			rep <- fmt.Errorf("getting sessions: %w", err)
		}
	}

	wg.Wait()

	return c.SendData(PlexRoute.Path(event), payload, true)
}

// getMetaSnap grabs some basic system info: cpu, memory, username.
func (c *Config) getMetaSnap(ctx context.Context) *snapshot.Snapshot {
	var (
		snap = &snapshot.Snapshot{}
		wg   sync.WaitGroup
	)

	rep := make(chan error)
	defer close(rep)

	go func() {
		for err := range rep {
			if err != nil { // maybe move this out of this method?
				c.Errorf("Building Metadata: %v", err)
			}
		}
	}()

	wg.Add(3) //nolint: gomnd,wsl
	go func() {
		rep <- snap.GetCPUSample(ctx)
		wg.Done() //nolint:wsl
	}()
	go func() {
		rep <- snap.GetMemoryUsage(ctx)
		wg.Done() //nolint:wsl
	}()
	go func() {
		for _, err := range snap.GetLocalData(ctx) {
			rep <- err
		}
		wg.Done() //nolint:wsl
	}()

	wg.Wait()

	return snap
}

func (c *Config) GetData(url string) (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout.Duration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("X-API-Key", c.Apps.APIKey)

	start := time.Now()

	// body gets closed in unmarshalResponse.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making http request: %w", err)
	}

	if !c.LogConfig.Debug {
		return unmarshalResponse(http.MethodGet, url, resp.StatusCode, io.NopCloser(resp.Body))
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	// copy the body into a buffer we can pass into json.Decode().
	tee := io.TeeReader(resp.Body, &buf) // must read tee first.
	defer c.debughttplog(resp, url, start, "", &buf)

	if err != nil {
		return nil, fmt.Errorf("reading http response body: %w", err)
	}

	return unmarshalResponse(http.MethodGet, url, resp.StatusCode, io.NopCloser(tee))
}

// SendData sends raw data to a notifiarr URL as JSON.
func (c *Config) SendData(uri string, payload interface{}, log bool) (*Response, error) {
	var (
		post []byte
		err  error
	)

	if data, err := json.Marshal(payload); err == nil {
		var torn map[string]interface{}
		if err := json.Unmarshal(data, &torn); err == nil {
			if torn["host"], err = c.GetHostInfoUID(); err != nil {
				c.Errorf("Host Info Unknown: %v", err)
			}

			payload = torn
		}
	}

	if log {
		post, err = json.MarshalIndent(payload, "", " ")
	} else {
		post, err = json.Marshal(payload)
	}

	if err != nil {
		return nil, fmt.Errorf("encoding data to JSON (report this bug please): %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout.Duration)
	defer cancel()

	code, body, err := c.sendJSON(ctx, c.BaseURL+uri, post, log)
	if err != nil {
		return nil, err
	}

	return unmarshalResponse(http.MethodPost, c.BaseURL+uri, code, body)
}

// unmarshalResponse attempts to turn the reply from notifiarr.com into structured data.
func unmarshalResponse(method, url string, code int, body io.ReadCloser) (*Response, error) {
	defer body.Close()

	var r Response

	counter := datacounter.NewReaderCounter(body)
	defer func() {
		websiteMap.Add(method+"BytesRcvd", int64(counter.Count()))
	}()

	err := json.NewDecoder(counter).Decode(&r)

	if code < http.StatusOK || code > http.StatusIMUsed {
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %d %s (unmarshal error: %v)",
				ErrNon200, url, code, http.StatusText(code), err)
		}

		return nil, fmt.Errorf("%w: %s: %d %s, %s: %s",
			ErrNon200, url, code, http.StatusText(code), r.Result, r.Details.Response)
	}

	if err != nil {
		return nil, fmt.Errorf("converting json response: %w", err)
	}

	return &r, nil
}

// sendJSON posts a JSON payload to a URL. Returns the response body or an error.
func (c *Config) sendJSON(ctx context.Context, url string, data []byte, log bool) (int, io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return 0, nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.Apps.APIKey)

	start := time.Now()

	resp, err := c.client.Do(req)
	if err != nil {
		c.debughttplog(nil, url, start, string(data), nil)
		return 0, nil, fmt.Errorf("making http request: %w", err)
	}

	if !c.LogConfig.Debug { // no debug, just return the body.
		return resp.StatusCode, resp.Body, nil
	}

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	defer resp.Body.Close() // close this since we return a fake one after logging.

	if log {
		defer c.debughttplog(resp, url, start, string(data), tee)
	} else {
		defer c.debughttplog(resp, url, start, "<data not logged>", tee)
	}

	return resp.StatusCode, io.NopCloser(&buf), nil
}

// Do performs an http Request with retries and logging!
func (h *httpClient) Do(req *http.Request) (*http.Response, error) {
	deadline, ok := req.Context().Deadline()
	if !ok {
		deadline = time.Now().Add(h.Timeout)
	}

	timeout := time.Until(deadline).Round(time.Millisecond)

	for retry := 0; ; retry++ {
		resp, err := h.Client.Do(req)
		if err == nil {
			for i, c := range resp.Cookies() {
				h.Printf("Unexpected cookie [%v/%v] returned from notifiarr.com: %s", i+1, len(resp.Cookies()), c.String())
			}

			if resp.StatusCode < http.StatusInternalServerError {
				websiteMap.Add(req.Method+"s", 1)
				websiteMap.Add(req.Method+"BytesSent", resp.Request.ContentLength)

				return resp, nil
			}

			// resp.StatusCode is 500 or higher, make that en error.
			size, _ := io.Copy(io.Discard, resp.Body) // must read the entire body when err == nil
			resp.Body.Close()                         // do not defer, because we're in a loop.
			websiteMap.Add(req.Method+"Retries", 1)
			websiteMap.Add(req.Method+"s", 1)
			websiteMap.Add(req.Method+"BytesSent", resp.Request.ContentLength)
			websiteMap.Add(req.Method+"BytesRcvd", size)
			// shoehorn a non-200 error into the empty http error.
			err = fmt.Errorf("%w: %s: %d bytes, %s", ErrNon200, req.URL, size, resp.Status)
		}

		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			if retry == 0 {
				return resp, fmt.Errorf("notifiarr req timed out after %s: %s: %w", timeout, req.URL, err)
			}

			return resp, fmt.Errorf("[%d/%d] Notifiarr req timed out after %s, giving up: %w",
				retry+1, h.Retries+1, timeout, err)
		case retry == h.Retries:
			return resp, fmt.Errorf("[%d/%d] Notifiarr req failed: %w", retry+1, h.Retries+1, err)
		default:
			h.Printf("[%d/%d] Notifiarr req failed, retrying in %s, error: %v", retry+1, h.Retries+1, RetryDelay, err)
			time.Sleep(RetryDelay)
		}
	}
}

func (c *Config) debughttplog(resp *http.Response, url string, start time.Time, data string, body io.Reader) {
	headers := ""
	status := "0"

	if resp != nil {
		status = resp.Status

		for k, vs := range resp.Header {
			for _, v := range vs {
				headers += k + ": " + v + "\n"
			}
		}
	}

	if c.MaxBody > 0 && len(data) > c.MaxBody {
		data = fmt.Sprintf("%s <data truncated, max: %d>", data[:c.MaxBody], c.MaxBody)
	}

	if data == "" {
		c.Debugf("Sent GET Request to %s in %s, Response (%s):\n%s\n%s",
			url, time.Since(start).Round(time.Microsecond), status, headers, readBodyForLog(body, int64(c.MaxBody)))
	} else {
		c.Debugf("Sent JSON Payload to %s in %s:\n%s\nResponse (%s):\n%s\n%s",
			url, time.Since(start).Round(time.Microsecond), data, status, headers, readBodyForLog(body, int64(c.MaxBody)))
	}
}

// SetValue sets a value stored in the website database.
func (c *Config) SetValue(key string, value []byte) error {
	return c.SetValues(map[string][]byte{key: value})
}

// SetValueContext sets a value stored in the website database.
func (c *Config) SetValueContext(ctx context.Context, key string, value []byte) error {
	return c.SetValuesContext(ctx, map[string][]byte{key: value})
}

// SetValues sets values stored in the website database.
func (c *Config) SetValues(values map[string][]byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout.Duration)
	defer cancel()

	return c.SetValuesContext(ctx, values)
}

// SetValuesContext sets values stored in the website database.
func (c *Config) SetValuesContext(ctx context.Context, values map[string][]byte) error {
	for key, val := range values {
		if val != nil { // ignore nil byte slices.
			values[key] = []byte(base64.StdEncoding.EncodeToString(val))
		}
	}

	data, err := json.Marshal(map[string]interface{}{"fields": values})
	if err != nil {
		return fmt.Errorf("converting values to json: %w", err)
	}

	code, body, err := c.sendJSON(ctx, c.BaseURL+ClientRoute.Path("setStates"), data, true)
	if err != nil {
		return fmt.Errorf("inalid response (%d): %w", code, err)
	}

	_, err = unmarshalResponse(http.MethodPost, c.BaseURL+ClientRoute.Path("getStates"), code, body)

	return err
}

// DelValue deletes a value stored in the website database.
func (c *Config) DelValue(keys ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout.Duration)
	defer cancel()

	return c.DelValueContext(ctx, keys...)
}

// DelValueContext deletes a value stored in the website database.
func (c *Config) DelValueContext(ctx context.Context, keys ...string) error {
	values := make(map[string]interface{})
	for _, key := range keys {
		values[key] = nil
	}

	data, err := json.Marshal(map[string]interface{}{"fields": values})
	if err != nil {
		return fmt.Errorf("converting values to json: %w", err)
	}

	code, body, err := c.sendJSON(ctx, c.BaseURL+ClientRoute.Path("setStates"), data, true)
	if err != nil {
		return fmt.Errorf("inalid response (%d): %w", code, err)
	}

	_, err = unmarshalResponse(http.MethodPost, c.BaseURL+ClientRoute.Path("setStates"), code, body)
	if err != nil {
		return err
	}

	return nil
}

// GetValue gets a value stored in the website database.
func (c *Config) GetValue(keys ...string) (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout.Duration)
	defer cancel()

	return c.GetValueContext(ctx, keys...)
}

// GetValueContext gets a value stored in the website database.
func (c *Config) GetValueContext(ctx context.Context, keys ...string) (map[string][]byte, error) {
	data, err := json.Marshal(map[string][]string{"fields": keys})
	if err != nil {
		return nil, fmt.Errorf("converting keys to json: %w", err)
	}

	code, body, err := c.sendJSON(ctx, c.BaseURL+ClientRoute.Path("getStates"), data, true)
	if err != nil {
		return nil, fmt.Errorf("inalid response (%d): %w", code, err)
	}

	resp, err := unmarshalResponse(http.MethodPost, c.BaseURL+ClientRoute.Path("getStates"), code, body)
	if err != nil {
		return nil, err
	}

	var output struct {
		LastUpdated time.Time         `json:"lastUpdated"`
		Fields      map[string][]byte `json:"fields"`
	}

	if err := json.Unmarshal(resp.Details.Response, &output); err != nil {
		return nil, fmt.Errorf("converting response values to json: %w", err)
	}

	for key, val := range output.Fields {
		b, err := base64.StdEncoding.DecodeString(string(val))
		if err != nil {
			return nil, fmt.Errorf("invalid base64 encoded data: %w", err)
		}

		output.Fields[key] = b
	}

	return output.Fields, nil
}

// readBodyForLog truncates the response body, or not, for the debug log. errors are ignored.
func readBodyForLog(body io.Reader, max int64) string {
	if body == nil {
		return ""
	}

	if max > 0 {
		limitReader := io.LimitReader(body, max)
		bodyBytes, _ := ioutil.ReadAll(limitReader)
		remaining, _ := io.Copy(io.Discard, body) // finish reading to the end.

		if remaining > 0 {
			return fmt.Sprintf("%s <body truncated, max: %d>", string(bodyBytes), max)
		}

		return string(bodyBytes)
	}

	bodyBytes, _ := ioutil.ReadAll(body)

	return string(bodyBytes)
}
