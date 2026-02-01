//nolint:godot
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/logs"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
)

// PlexHandler handles an incoming webhook from Plex.
//
//	@Summary		Accept Plex Media Server Webhook
//	@Description	Accepts a Plex webhook; when conditions are satisfied sends a notification to the website,
//	@Description	and may include snapshot data and/or fetched session data. Does not require X-API-Key header.
//	@Tags			Plex
//	@Accept			json
//	@Produce		text/plain
//	@Param			token	query		string					true	"Plex Token or Client API Key"
//	@Param			POST	body		plex.IncomingWebhook	true	"webhook payload"
//	@Success		202		{string}	string					"accepted"
//	@Success		208		{string}	string					"ignored"
//	@Failure		400		{string}	string					"bad input"
//	@Failure		404		{string}	string					"bad token or api key"
//	@Router			/plex [post]
func (c *Client) PlexHandler(w http.ResponseWriter, r *http.Request) { //nolint:cyclop,varnamelen,funlen
	mnd.Apps.Add("Plex&&Incoming Webhooks", 1)

	start := time.Now()

	r.Body = apps.NewFakeCloser("Plex", "Webhook", r.Body)
	defer r.Body.Close()

	if err := r.ParseMultipartForm(mnd.Megabyte); err != nil {
		logs.Log.Errorf("Parsing Multipart Form (plex): %v", err)
		mnd.Apps.Add("Plex&&Webhook Errors", 1)
		http.Error(w, "form parse error", http.StatusBadRequest)

		return
	}

	payload := r.Form.Get("payload")
	logs.Log.Debugf("Plex Webhook Payload: %s", payload)
	r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))

	hook := plex.IncomingWebhook{ReqID: mnd.GetID(r.Context())}

	switch err := json.Unmarshal([]byte(payload), &hook); {
	case err != nil:
		mnd.Apps.Add("Plex&&Webhook Errors", 1)
		http.Error(w, "payload error", http.StatusBadRequest)
		logs.Log.Errorf("Unmarshalling Plex payload: %v", err)
	case strings.EqualFold(hook.Event, "admin.database.backup"):
		fallthrough
	case strings.EqualFold(hook.Event, "device.new"):
		fallthrough
	case strings.EqualFold(hook.Event, "media.rate"):
		fallthrough
	case strings.EqualFold(hook.Event, "library.new"):
		fallthrough
	case strings.EqualFold(hook.Event, "admin.database.corrupt"):
		logs.Log.Printf("{trace:%s} Plex Incoming Webhook: %s, %s '%s' ~> %s (relaying to Notifiarr)",
			mnd.GetID(r.Context()), hook.Server.Title, hook.Account.Title, hook.Event, hook.Metadata.Title)
		website.SendData(&website.Request{
			ReqID:      mnd.GetID(r.Context()),
			Route:      website.PlexRoute,
			Event:      website.EventHook,
			LogPayload: true,
			LogMsg:     fmt.Sprintf("Plex Webhook: %s '%s' ~> %s", hook.Account.Title, hook.Event, hook.Metadata.Title),
			Payload: &website.Payload{
				Snap: c.triggers.PlexCron.GetMetaSnap(r.Context()),
				Load: &hook,
				Plex: &plex.Sessions{Name: c.apps.Plex.Name()},
			},
		})
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		http.Error(w, "process", http.StatusAccepted)
	case strings.EqualFold(hook.Event, "media.resume") && c.plexTimer.Active(hook.Metadata.Key+"resume", c.plexCooldown()):
		logs.Log.Printf("{trace:%s} Plex Incoming Webhook Ignored (cooldown): %s, %s '%s' ~> %s",
			mnd.GetID(r.Context()), hook.Server.Title, hook.Account.Title, hook.Event, hook.Metadata.Title)
		http.Error(w, "ignored, cooldown", http.StatusAlreadyReported)
	case strings.EqualFold(hook.Event, "media.play"), strings.EqualFold(hook.Event, "playback.started"):
		if c.plexTimer.Active(hook.Metadata.Key+"play", c.plexCooldown()) {
			logs.Log.Printf("{trace:%s} Plex Incoming Webhook Ignored (cooldown): %s, %s '%s' ~> %s",
				mnd.GetID(r.Context()), hook.Server.Title, hook.Account.Title, hook.Event, hook.Metadata.Title)
			http.Error(w, "ignored, cooldown", http.StatusAlreadyReported)

			return
		}

		fallthrough
	case strings.EqualFold(hook.Event, "media.scrobble"):
		fallthrough
	case strings.EqualFold(hook.Event, "media.resume"):
		c.triggers.PlexCron.SendWebhook(&hook) //nolint:contextcheck,nolintlint
		logs.Log.Printf("{trace:%s} Plex Incoming Webhook: %s, %s '%s' ~> %s (collecting sessions)",
			mnd.GetID(r.Context()), hook.Server.Title, hook.Account.Title, hook.Event, hook.Metadata.Title)
		r.Header.Set("X-Request-Time", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))
		http.Error(w, "processing", http.StatusAccepted)
	default:
		http.Error(w, "ignored, unsupported", http.StatusAlreadyReported)
		logs.Log.Printf("{trace:%s} Plex Incoming Webhook Ignored (unsupported): %s, %s '%s' ~> %s",
			mnd.GetID(r.Context()), hook.Server.Title, hook.Account.Title, hook.Event, hook.Metadata.Title)
	}
}

func (c *Client) plexCooldown() time.Duration {
	if ci := clientinfo.Get(); ci != nil {
		return ci.Actions.Plex.Cooldown.Duration
	}

	return time.Minute
}
