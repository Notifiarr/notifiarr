// Notifiarr package provides a standard interface for sending data to notifiarr.com.
// This includes crontabs that run
package notifiarr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	ErrNon200 = fmt.Errorf("return code was not 200")
)

// Notifiarr URLs.
const (
	BaseURL     = "https://notifiarr.com"
	ProdURL     = BaseURL + "/notifier.php"
	TestURL     = BaseURL + "/notifierTest.php"
	DevBaseURL  = "http://dev.notifiarr.com"
	DevURL      = DevBaseURL + "/notifier.php"
	APIKeyRoute = "/api/v1/user/apikey/"
)

// These are used as 'source' values in json payloads sent to the webserver.
const (
	PlexCron = "plexcron"
	SnapCron = "snapcron"
	PlexHook = "plexhook"
	LogLocal = "loglocal"
)

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
	URL          string
	BaseURL      string
	Timeout      time.Duration
	*logs.Logger // log file writer
	stopPlex     chan struct{}
	stopSnap     chan struct{}
	client       *http.Client
}

// Start (and log) snapshot and plex cron jobs if they're configured.
func (c *Config) Start(mode string) {
	switch mode {
	default:
		fallthrough
	case "test", "testing":
		c.URL = TestURL
		c.BaseURL = BaseURL
	case "prod", "production":
		c.URL = ProdURL
		c.BaseURL = BaseURL
	case "dev", "development":
		c.URL = DevURL
		c.BaseURL = DevBaseURL
	}

	go c.startSnapCron()
	go c.startPlexCron()
}

// Stop snapshot and plex cron jobs.
func (c *Config) Stop() {
	if c != nil && c.stopSnap != nil {
		c.stopSnap <- struct{}{}
	}

	if c != nil && c.stopPlex != nil {
		c.stopPlex <- struct{}{}
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

	_, e, err := c.SendData(url, payload)

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

// CheckAPIKey returns an error if the API key is wrong.
func (c *Config) CheckAPIKey() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+APIKeyRoute+c.Apps.APIKey, nil)
	if err != nil {
		return fmt.Errorf("creating http request: %w", err)
	}

	resp, err := c.getClient().Do(req)
	if err != nil {
		return fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	_, _ = io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return ErrNon200
	}

	return nil
}

// SendJSON posts a JSON payload to a URL. Returns the response body or an error.
// The response status code is lost.
func (c *Config) SendJSON(url string, data []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.Apps.APIKey)
	c.Debugf("Sending JSON Payload to %s:\n%s", url, string(data))

	resp, err := c.getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, fmt.Errorf("reading http response: %w, body: %s", err, string(body))
	}

	return body, nil
}

func (c *Config) SendData(url string, payload interface{}) ([]byte, []byte, error) {
	post, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		return nil, nil, fmt.Errorf("encoding data: %w", err)
	}

	reply, err := c.SendJSON(url, post)

	return post, reply, err
}

func (c *Config) getClient() *http.Client {
	if c.client == nil {
		c.client = &http.Client{Timeout: c.Timeout}
	}

	return c.client
}
