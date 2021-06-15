// Package notifiarr provides a standard interface for sending data to notifiarr.com.
// Several methods are exported to make POSTing data to notifarr easier. This package
// also handles the "crontab" timers for plex sessions, snapshots, custom format sync
// for Radarr and release profile sync for Sonarr. This package's cofiguration is
// provided by the configfile package.
package notifiarr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
)

// Errors returned by this library.
var (
	ErrNon200          = fmt.Errorf("return code was not 200")
	ErrInvalidResponse = fmt.Errorf("invalid response")
)

// Notifiarr URLs.
const (
	BaseURL     = "https://notifiarr.com"
	ProdURL     = BaseURL + "/notifier.php"
	TestURL     = BaseURL + "/notifierTest.php"
	DevBaseURL  = "http://dev.notifiarr.com"
	DevURL      = DevBaseURL + "/notifier.php"
	APIKeyRoute = "/api/v1/user/apikey"
	// CFSyncRoute is the webserver route to send sync requests to.
	CFSyncRoute = "/api/v1/user/trash"
)

// These are used as 'source' values in json payloads sent to the webserver.
const (
	PlexCron = "plexcron"
	SnapCron = "snapcron"
	PlexHook = "plexhook"
	LogLocal = "loglocal"
)

const (
	// DefaultRetries is the number of times to attempt a request to notifiarr.com.
	// 4 means 5 total tries: 1 try + 4 retries.
	DefaultRetries = 4
	// RetryDelay is how long to Sleep between retries.
	RetryDelay = 222 * time.Millisecond
)

// success is a ssuccessful tatus message from notifiarr.com.
const success = "success"

// Payload is the outbound payload structure that is sent to Notifiarr for Plex and system snapshot data.
type Payload struct {
	Type string             `json:"eventType"`
	Plex *plex.Sessions     `json:"plex,omitempty"`
	Snap *snapshot.Snapshot `json:"snapshot,omitempty"`
	Load *plex.Webhook      `json:"payload,omitempty"`
}

// Config is the input data needed to send payloads to notifiarr.
type Config struct {
	Apps         *apps.Apps       // has API key
	Plex         *plex.Server     // plex sessions
	Snap         *snapshot.Config // system snapshot data
	Retries      int
	URL          string
	BaseURL      string
	Timeout      time.Duration
	*logs.Logger // log file writer
	stopTimers   chan struct{}
	client       *httpClient
	radarrCFs    map[int]*cfMapIDpayload
	sonarrCFs    map[int]*cfMapIDpayload
}

// Start (and log) snapshot and plex cron jobs if they're configured.
func (c *Config) Start(mode string) {
	switch mode {
	default:
		fallthrough
	case "prod", "production":
		c.URL = ProdURL
		c.BaseURL = BaseURL
	case "test", "testing":
		c.URL = TestURL
		c.BaseURL = BaseURL
	case "dev", "development":
		c.URL = DevURL
		c.BaseURL = DevBaseURL
	}

	if c.Retries < 0 {
		c.Retries = 0
	} else if c.Retries == 0 {
		c.Retries = DefaultRetries
	}

	c.radarrCFs = make(map[int]*cfMapIDpayload)
	c.sonarrCFs = make(map[int]*cfMapIDpayload)

	c.startTimers()
}

// Stop snapshot and plex cron jobs.
func (c *Config) Stop() {
	if c != nil && c.stopTimers != nil {
		c.stopTimers <- struct{}{}
	}
}

// SendMeta is kicked off by the webserver in go routine.
// It's also called by the plex cron (with webhook set to nil).
// This runs after Plex drops off a webhook telling us someone did something.
// This gathers cpu/ram, and waits 10 seconds, then grabs plex sessions.
// It's all POSTed to notifiarr. May be used with a nil Webhook.
func (c *Config) SendMeta(eventType, url string, hook *plex.Webhook, wait time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), wait+c.Snap.Timeout.Duration)
	defer cancel()

	var (
		payload = &Payload{Type: eventType, Load: hook}
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
		payload.Snap = c.GetMetaSnap(ctx)
		wg.Done() // nolint:wsl
	}()

	time.Sleep(wait)

	var err error
	if payload.Plex, err = c.Plex.GetXMLSessions(); err != nil {
		rep <- fmt.Errorf("getting sessions: %w", err)
	}

	wg.Wait()

	_, _, e, err := c.SendData(url, payload) //nolint:bodyclose // already closed

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

// CheckAPIKey returns an error if the API key is wrong. Returns a message otherwise.
func (c *Config) CheckAPIKey() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+APIKeyRoute, nil)
	if err != nil {
		return "", fmt.Errorf("creating http request: %w", err)
	}

	c.Debugf("=> Checking API Key @ %s", req.URL)
	req.Header.Set("X-API-Key", c.Apps.APIKey)

	resp, err := c.getClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var v struct {
		Message string `json:"message"`
	}

	if err = json.Unmarshal(body, &v); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return v.Message, ErrNon200
	}

	return v.Message, nil
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
	c.Debugf("Sending JSON Payload to %s:\n%s", url, string(data))

	resp, err := c.getClient().Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, fmt.Errorf("reading http response: %w, body: %s", err, string(body))
	}

	return resp, body, nil
}

// SendData sends raw data to a notifiarr URL as JSON.
func (c *Config) SendData(url string, payload interface{}) (*http.Response, []byte, []byte, error) {
	post, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("encoding data: %w", err)
	}

	resp, reply, err := c.SendJSON(url, post)

	return resp, post, reply, err
}

// httpClient is our custom http client to wrap Do and provide retries.
type httpClient struct {
	Retries int
	*log.Logger
	*http.Client
}

// getClient returns an http client for notifiarr.com. Creates one if it doesn't exist yet.
func (c *Config) getClient() *httpClient {
	if c.client == nil {
		c.client = &httpClient{
			Retries: c.Retries,
			Logger:  c.ErrorLog,
			Client:  &http.Client{Timeout: c.Timeout},
		}
	}

	return c.client
}

// Do performs an http Request with retries and logging!
func (h *httpClient) Do(req *http.Request) (*http.Response, error) {
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
		case i == h.Retries:
			return nil, fmt.Errorf("[%d/%d] notifiarr.com req failed: %w", i+1, h.Retries+1, err)
		default:
			h.Printf("[%d/%d] Request to Notifiarr.com failed, retrying in %d, error: %v", i+1, h.Retries+1, RetryDelay, err)
			time.Sleep(RetryDelay)
		}
	}
}
