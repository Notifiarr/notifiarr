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

	"github.com/Go-Lift-TV/notifiarr/pkg/apps"
	"github.com/Go-Lift-TV/notifiarr/pkg/logs"
	"github.com/Go-Lift-TV/notifiarr/pkg/plex"
	"github.com/Go-Lift-TV/notifiarr/pkg/snapshot"
)

var ErrNon200 = fmt.Errorf("return code was not 200")

// Notifiarr URLs.
const (
	BaseURL = "https://notifiarr.com"
	ProdURL = BaseURL + "/notifier.php"
	TestURL = BaseURL + "/notifierTest.php"
	DevURL  = "http://dev.notifiarr.com/notifier.php"
)

const (
	PlexCron = "plexcron"
	SnapCron = "snapcron"
	PlexHook = "plexhook"
	LogLocal = "loglocal"
)

// Payload is the outbound payload structure that is sent to Notifiarr.
// No other payload formats are used for data sent to notifiarr.com.
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
	case "prod", "production":
		c.URL = ProdURL
	case "dev", "development":
		c.URL = DevURL
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
func (c *Config) SendMeta(eventType, url string, hook *plex.Webhook, wait time.Duration) ([]byte, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), wait+time.Second*2)
	defer cancel()

	var (
		payload = &Payload{Type: eventType, Snap: &snapshot.Snapshot{}, Load: hook}
		wg      sync.WaitGroup
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
		rep <- payload.Snap.GetCPUSample(ctx, true)
		wg.Done() //nolint:wsl
	}()
	go func() {
		rep <- payload.Snap.GetMemoryUsage(ctx, true)
		wg.Done() //nolint:wsl
	}()
	go func() {
		for _, err := range payload.Snap.GetLocalData(ctx, false) {
			rep <- err
		}
		wg.Done() //nolint:wsl
	}()

	time.Sleep(wait)

	var err error
	if payload.Plex, err = c.Plex.GetSessions(); err != nil {
		rep <- fmt.Errorf("getting sessions: %w", err)
	}

	wg.Wait()

	return c.SendData(url, payload)
}

// CheckURLResponse returns nil if the request returns a 200.
func (c *Config) CheckAPIKey() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, BaseURL+"/api/user/0/apikey/"+c.Apps.APIKey, nil)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.Apps.APIKey)

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

func (c *Config) SendData(url string, payload *Payload) ([]byte, []byte, error) {
	post, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding data: %w", err)
	}

	reply, err := c.SendJSON(url, post)

	return post, reply, err
}

func (c *Config) getClient() *http.Client {
	if c.client == nil {
		c.client = &http.Client{Timeout: time.Minute}
	}

	return c.client
}
