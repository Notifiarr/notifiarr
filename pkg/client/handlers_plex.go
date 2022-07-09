package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/exp"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// Timer is used to set a cooldown time.
type Timer struct {
	lock  sync.Mutex
	start time.Time
}

// Active returns true if a timer is active, otherwise it becomes active.
func (t *Timer) Active(d time.Duration) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	if time.Since(t.start) < d {
		return true
	}

	t.start = time.Now()

	return false
}

// PlexHandler handles an incoming webhook from Plex.
func (c *Client) PlexHandler(w http.ResponseWriter, r *http.Request) { //nolint:cyclop,varnamelen,funlen
	exp.Apps.Add("Plex&&Incoming Webhooks", 1)

	start := time.Now()
	r.Body = exp.NewFakeCloser("Plex", "Webhook", r.Body)

	if err := r.ParseMultipartForm(mnd.Megabyte); err != nil {
		c.Errorf("Parsing Multipart Form (plex): %v", err)
		exp.Apps.Add("Plex&&Webhook Errors", 1)
		http.Error(w, "form parse error", http.StatusBadRequest)

		return
	}

	payload := r.Form.Get("payload")
	c.Debugf("Plex Webhook Payload: %s", payload)
	r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))

	var v plex.IncomingWebhook

	switch err := json.Unmarshal([]byte(payload), &v); {
	case err != nil:
		exp.Apps.Add("Plex&&Webhook Errors", 1)
		http.Error(w, "payload error", http.StatusBadRequest)
		c.Errorf("Unmarshalling Plex payload: %v", err)
	case strings.EqualFold(v.Event, "admin.database.backup"):
		fallthrough
	case strings.EqualFold(v.Event, "device.new"):
		fallthrough
	case strings.EqualFold(v.Event, "media.rate"):
		fallthrough
	case strings.EqualFold(v.Event, "library.new"):
		fallthrough
	case strings.EqualFold(v.Event, "admin.database.corrupt"):
		c.Printf("Plex Incoming Webhook: %s, %s '%s' ~> %s (relaying to Notifiarr)",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		c.website.SendData(&website.Request{
			Route:      website.PlexRoute,
			Event:      website.EventHook,
			LogPayload: true,
			LogMsg:     fmt.Sprintf("Plex Webhhok: %s '%s' ~> %s", v.Account.Title, v.Event, v.Metadata.Title),
			Payload:    &website.Payload{Load: &v, Plex: &plex.Sessions{Name: c.Config.Plex.Name}},
		})
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		http.Error(w, "process", http.StatusAccepted)
	case strings.EqualFold(v.Event, "media.resume") && c.plexTimer.Active(c.plexCooldown()):
		c.Printf("Plex Incoming Webhook Ignored (cooldown): %s, %s '%s' ~> %s",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		http.Error(w, "ignored, cooldown", http.StatusAlreadyReported)
	case strings.EqualFold(v.Event, "media.play"):
		fallthrough
	case strings.EqualFold(v.Event, "media.resume"):
		go c.triggers.PlexCron.SendWebhook(&v)
		c.Printf("Plex Incoming Webhook: %s, %s '%s' ~> %s (collecting sessions)",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		http.Error(w, "processing", http.StatusAccepted)
	default:
		http.Error(w, "ignored, unsupported", http.StatusAlreadyReported)
		c.Printf("Plex Incoming Webhook Ignored (unsupported): %s, %s '%s' ~> %s",
			v.Server.Title, v.Account.Title, v.Event, v.Metadata.Title)
	}
}

func (c *Client) plexCooldown() time.Duration {
	if ci, _ := c.website.GetClientInfo(); ci != nil {
		return ci.Actions.Plex.Cooldown.Duration
	}

	return time.Minute
}
