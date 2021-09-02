package notifiarr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
)

// SendMeta is kicked off by the webserver in go routine.
// It's also called by the plex cron (with webhook set to nil).
// This runs after Plex drops off a webhook telling us someone did something.
// This gathers cpu/ram, and waits 10 seconds, then grabs plex sessions.
// It's all POSTed to notifiarr. May be used with a nil Webhook.
func (c *Config) SendMeta(eventType, url string, hook *plexIncomingWebhook, wait bool) ([]byte, error) {
	extra := time.Second
	if wait {
		extra = plex.WaitTime
	}

	ctx, cancel := context.WithTimeout(context.Background(), extra+c.Snap.Timeout.Duration)
	defer cancel()

	var (
		payload = &Payload{
			Type: eventType,
			Load: hook,
			Plex: &plex.Sessions{
				Name:       c.Plex.Name,
				AccountMap: strings.Split(c.Plex.AccountMap, "|"),
			},
		}
		wg sync.WaitGroup
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
		payload.Snap = c.GetMetaSnap(ctx)
		wg.Done() // nolint:wsl
	}()

	if !wait || !c.Plex.NoActivity {
		var err error
		if payload.Plex, err = c.GetSessions(wait); err != nil {
			rep <- fmt.Errorf("getting sessions: %w", err)
		}
	}

	wg.Wait()

	_, e, err := c.SendData(url, payload, true) //nolint:bodyclose // already closed

	return e, err
}

// GetMetaSnap grabs some basic system info: cpu, memory, username.
func (c *Config) GetMetaSnap(ctx context.Context) *snapshot.Snapshot {
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
		rep <- snap.GetCPUSample(ctx, true)
		wg.Done() //nolint:wsl
	}()
	go func() {
		rep <- snap.GetMemoryUsage(ctx, true)
		wg.Done() //nolint:wsl
	}()
	go func() {
		for _, err := range snap.GetLocalData(ctx, false) {
			rep <- err
		}
		wg.Done() //nolint:wsl
	}()

	wg.Wait()

	return snap
}

// SendJSON posts a JSON payload to a URL. Returns the response body or an error.
func (c *Config) SendJSON(url string, data []byte) (*http.Response, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.Apps.APIKey)

	start := time.Now()

	resp, err := c.client.Do(req)
	if err != nil {
		c.Debugf("Sent JSON Payload to %s in %s:\n%s\nResponse (0): %s",
			url, time.Since(start).Round(time.Microsecond), string(data), err)
		return nil, nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	defer func() {
		headers := ""

		for k, vs := range resp.Header {
			for _, v := range vs {
				headers += k + ": " + v + "\n"
			}
		}

		c.Debugf("Sent JSON Payload to %s in %s:\n%s\nResponse (%s):\n%s\n%s",
			url, time.Since(start).Round(time.Microsecond), string(data), resp.Status, headers, string(body))
	}()

	if err != nil {
		return resp, body, fmt.Errorf("reading http response body: %w", err)
	}

	return resp, body, nil
}

func (c *Config) GetData(url string) (*http.Response, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("X-API-Key", c.Apps.APIKey)

	start := time.Now()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	defer func() {
		headers := ""

		for k, vs := range resp.Header {
			for _, v := range vs {
				headers += k + ": " + v + "\n"
			}
		}

		c.Debugf("Sent GET Request to %s in %s, Response (%s):\n%s\n%s",
			url, time.Since(start).Round(time.Microsecond), resp.Status, headers, string(body))
	}()

	if err != nil {
		return resp, body, fmt.Errorf("reading http response body: %w", err)
	}

	return resp, body, nil
}

// SendData sends raw data to a notifiarr URL as JSON.
func (c *Config) SendData(url string, payload interface{}, pretty bool) (*http.Response, []byte, error) {
	var (
		post []byte
		err  error
	)

	if pretty {
		post, err = json.MarshalIndent(payload, "", " ")
	} else {
		post, err = json.Marshal(payload)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("encoding data to JSON (report this bug please): %w", err)
	}

	return c.SendJSON(url, post)
}

// httpClient is our custom http client to wrap Do and provide retries.
type httpClient struct {
	Retries int
	*log.Logger
	*http.Client
}

// Do performs an http Request with retries and logging!
func (h *httpClient) Do(req *http.Request) (*http.Response, error) {
	deadline, ok := req.Context().Deadline()
	if !ok {
		deadline = time.Now().Add(h.Timeout)
	}

	timeout := time.Until(deadline).Round(time.Millisecond)

	for i := 0; ; i++ {
		resp, err := h.Client.Do(req)
		if err == nil && resp.StatusCode < http.StatusInternalServerError {
			return resp, nil
		} else if err == nil { // resp.StatusCode is 500 or higher, make that en error.
			body, _ := ioutil.ReadAll(resp.Body) // must read the entire body when err == nil
			resp.Body.Close()                    // do not defer, because we're in a loop.
			// shoehorn a non-200 error into the empty http error.
			err = fmt.Errorf("%w: %s: %s", ErrNon200, resp.Status, string(body))
		}

		switch {
		case errors.Is(err, context.DeadlineExceeded):
			if i == 0 {
				return nil, fmt.Errorf("notifiarr.com req timed out after %s: %w", timeout, err)
			}

			return nil, fmt.Errorf("[%d/%d] notifiarr.com reqs timed out after %s, giving up: %w",
				i+1, h.Retries+1, timeout, err)
		case i == h.Retries:
			return nil, fmt.Errorf("[%d/%d] notifiarr.com req failed: %w", i+1, h.Retries+1, err)
		default:
			h.Printf("[%d/%d] Request to Notifiarr.com failed, retrying in %s, error: %v", i+1, h.Retries+1, RetryDelay, err)
			time.Sleep(RetryDelay)
		}
	}
}
