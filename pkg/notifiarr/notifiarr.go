// Package notifiarr provides a standard interface for sending data to notifiarr.com.
// Several methods are exported to make POSTing data to notifarr easier. This package
// also handles the incoming Plex webhook as well as the "crontab" timers for plex
// sessions, snapshots, dashboard state, custom format sync for Radarr and release
// profile sync for Sonarr.
// This package's cofiguration is provided by the configfile  package.
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
	ClientRoute = "/api/v1/user/client"
	// CFSyncRoute is the webserver route to send sync requests to.
	CFSyncRoute = "/api/v1/user/trash"
	DashRoute   = "/api/v1/user/dashboard"
	GapsRoute   = "/api/v1/user/gaps"
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
	Type string               `json:"eventType"`
	Plex *plex.Sessions       `json:"plex,omitempty"`
	Snap *snapshot.Snapshot   `json:"snapshot,omitempty"`
	Load *plexIncomingWebhook `json:"payload,omitempty"`
}

// Config is the input data needed to send payloads to notifiarr.
type Config struct {
	Apps         *apps.Apps       // has API key
	Plex         *plex.Server     // plex sessions
	Snap         *snapshot.Config // system snapshot data
	DashDur      time.Duration
	Retries      int
	URL          string
	BaseURL      string
	Timeout      time.Duration
	Trigger      Triggers
	*logs.Logger // log file writer
	extras
}

type extras struct {
	clientInfo *ClientInfo
	client     *httpClient
	radarrCF   map[int]*cfMapIDpayload
	sonarrRP   map[int]*cfMapIDpayload
	plexTimer  *Timer
}

// Triggers allow trigger actions in the timer routine.
type Triggers struct {
	stop   chan struct{}      // Triggered by calling Stop()
	syncCF chan chan struct{} // Sync Radarr CF and Sonarr RP
	gaps   chan string        // Send Radarr Collection Gaps
	stuck  chan string        // Stuck Items
	plex   chan string        // Send Plex Sessions
	state  chan struct{}      // Dashboard State
	snap   chan string        // Snapshot
	sess   chan time.Time     // Return Plex Sessions
	sessr  chan *holder       // Session Return Channel
}

// Start (and log) snapshot and plex cron jobs if they're configured.
func (c *Config) Start(mode string) {
	if c.Trigger.stop != nil {
		panic("notifiarr timers cannot run twice")
	}

	switch strings.ToLower(mode) {
	default:
		fallthrough
	case "prod", "production":
		c.URL = ProdURL
		c.BaseURL = BaseURL
	case "test", "testing":
		c.URL = TestURL
		c.BaseURL = BaseURL
	case "dev", "devel", "development":
		c.URL = DevURL
		c.BaseURL = DevBaseURL
	}

	if c.Retries < 0 {
		c.Retries = 0
	} else if c.Retries == 0 {
		c.Retries = DefaultRetries
	}

	c.extras.radarrCF = make(map[int]*cfMapIDpayload)
	c.extras.sonarrRP = make(map[int]*cfMapIDpayload)
	c.extras.plexTimer = &Timer{}
	c.extras.client = &httpClient{
		Retries: c.Retries,
		Logger:  c.ErrorLog,
		Client:  &http.Client{},
	}
	c.Trigger.syncCF = make(chan chan struct{})
	c.Trigger.stuck = make(chan string)
	c.Trigger.plex = make(chan string)
	c.Trigger.state = make(chan struct{})
	c.Trigger.snap = make(chan string)
	c.Trigger.sess = make(chan time.Time)
	c.Trigger.gaps = make(chan string)

	go c.runSessionHolder()
	c.startTimers()
}

// Stop snapshot and plex cron jobs.
func (c *Config) Stop() {
	if c != nil && c.Trigger.stop != nil {
		c.Trigger.stop <- struct{}{}
		close(c.Trigger.syncCF)
		close(c.Trigger.stuck)
		close(c.Trigger.plex)
		close(c.Trigger.state)
		close(c.Trigger.snap)
		close(c.Trigger.gaps)

		defer close(c.Trigger.sess)
		c.Trigger.sess = nil
	}
}

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
