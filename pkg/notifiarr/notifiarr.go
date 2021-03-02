// Notifiarr package provides a standard interface for sending data to notifiarr.com.
// This includes crontabs that run
package notifiarr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/Go-Lift-TV/discordnotifier-client/pkg/apps"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/logs"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/plex"
	"github.com/Go-Lift-TV/discordnotifier-client/pkg/snapshot"
)

// Notifiarr URLs.
const (
	URL     = "https://discordnotifier.com/notifier.php"
	TestURL = "https://discordnotifier.com/notifierTest.php"
)

// Payload is the outbound payload structure that is sent to Notifiarr.
// No other payload formats are used for data sent to notifiarr.com.
type Payload struct {
	Plex *plex.Sessions     `json:"plex,omitempty"`
	Snap *snapshot.Snapshot `json:"snapshot,omitempty"`
	Load *plex.Webhook      `json:"payload,omitempty"`
}

// Config is the input data needed to send payloads to notifiarr.
type Config struct {
	Apps         *apps.Apps       `json:"-"` // has API key
	Plex         *plex.Server     `json:"-"` // plex sessions
	Snap         *snapshot.Config `json:"-"` // system snapshot data
	*logs.Logger `json:"-"`       // log file writer
	stopPlex     chan struct{}
	stopSnap     chan struct{}
	client       *http.Client
}

// Start (and log) snapshot and plex cron jobs if they're configured.
func (c *Config) Start() {
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
func (c *Config) SendMeta(hook *plex.Webhook, wait time.Duration) (b []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	var (
		snap    = &snapshot.Snapshot{}
		payload = &Payload{Snap: snap, Load: hook}
		wg      sync.WaitGroup
	)

	wg.Add(3) //nolint: gomnd,wsl
	go func() {
		_ = snap.GetCPUSample(ctx, true)
		wg.Done() //nolint:wsl
	}()
	go func() {
		_ = snap.GetMemoryUsage(ctx, true)
		wg.Done() //nolint:wsl
	}()
	go func() {
		_ = snap.GetLocalData(ctx, false)
		wg.Done() //nolint:wsl
	}()

	time.Sleep(wait)

	if payload.Plex, err = c.Plex.GetSessions(); err != nil {
		return nil, fmt.Errorf("getting sessions: %w", err)
	}

	wg.Wait()

	return c.SendData(TestURL, payload)
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

func (c *Config) SendData(url string, payload *Payload) ([]byte, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding data: %w", err)
	}

	return c.SendJSON(url, b)
}

func (c *Config) getClient() *http.Client {
	if c.client == nil {
		c.client = &http.Client{Timeout: time.Minute}
	}

	return c.client
}
