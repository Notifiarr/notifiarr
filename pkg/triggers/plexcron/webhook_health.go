package plexcron

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
)

const (
	defaultWebhookStaleAfter   = 20 * time.Minute
	defaultWebhookAlertCD      = 6 * time.Hour
	defaultWebhookStartupGrace = 10 * time.Minute
)

type webhookHealthConfig struct {
	Enabled       bool
	StaleAfter    time.Duration
	AlertCooldown time.Duration
	StartupGrace  time.Duration
}

type webhookHealthState struct {
	StartAt       time.Time
	LastWebhookAt time.Time
	LastAlertAt   time.Time
}

func defaultWebhookHealthConfig() webhookHealthConfig {
	return webhookHealthConfig{
		Enabled:       true,
		StaleAfter:    defaultWebhookStaleAfter,
		AlertCooldown: defaultWebhookAlertCD,
		StartupGrace:  defaultWebhookStartupGrace,
	}
}

func (c *cmd) currentWebhookHealthConfig() webhookHealthConfig {
	cfg := defaultWebhookHealthConfig()
	ci := clientinfo.Get()
	if ci == nil {
		return cfg
	}

	siteCfg := ci.Actions.Plex
	if siteCfg.WebhookHealthEnabled != nil {
		cfg.Enabled = *siteCfg.WebhookHealthEnabled
	}

	if siteCfg.WebhookStaleAfter.Duration > 0 {
		cfg.StaleAfter = siteCfg.WebhookStaleAfter.Duration
	}

	if siteCfg.WebhookAlertCooldown.Duration > 0 {
		cfg.AlertCooldown = siteCfg.WebhookAlertCooldown.Duration
	}

	if siteCfg.WebhookStartupGrace.Duration > 0 {
		cfg.StartupGrace = siteCfg.WebhookStartupGrace.Duration
	}

	return cfg
}

func (c *cmd) currentWebhookHealthState() webhookHealthState {
	c.healthLock.Lock()
	defer c.healthLock.Unlock()

	return webhookHealthState{
		StartAt:       c.startAt,
		LastWebhookAt: c.lastWebhookAt,
		LastAlertAt:   c.lastAlertAt,
	}
}

func (c *cmd) recordWebhookAt(t time.Time) {
	c.healthLock.Lock()
	defer c.healthLock.Unlock()

	c.lastWebhookAt = t
}

func (c *cmd) setWebhookAlertAt(t time.Time) {
	c.healthLock.Lock()
	defer c.healthLock.Unlock()

	c.lastAlertAt = t
}

func shouldAlertForStaleWebhook(
	now time.Time,
	activeSessions int,
	cfg webhookHealthConfig,
	state webhookHealthState,
) bool {
	switch {
	case !cfg.Enabled:
		return false
	case activeSessions < 1:
		return false
	case cfg.StartupGrace > 0 && now.Sub(state.StartAt) < cfg.StartupGrace:
		return false
	case !state.LastAlertAt.IsZero() && cfg.AlertCooldown > 0 && now.Sub(state.LastAlertAt) < cfg.AlertCooldown:
		return false
	case state.LastWebhookAt.IsZero():
		return true
	default:
		return now.Sub(state.LastWebhookAt) > cfg.StaleAfter
	}
}

func (c *cmd) evaluateWebhookHealth(ctx context.Context, sessions *plex.Sessions) {
	if sessions == nil {
		return
	}

	now := time.Now()
	state := c.currentWebhookHealthState()
	cfg := c.currentWebhookHealthConfig()
	if !shouldAlertForStaleWebhook(now, len(sessions.Sessions), cfg, state) {
		return
	}

	c.setWebhookAlertAt(now)

	hook := &plex.IncomingWebhook{ReqID: mnd.GetID(ctx), Event: "notifiarr.webhook.health"}
	hook.Server.Title = c.Plex.Name()
	hook.Metadata.Type = "healthcheck"
	hook.Metadata.Title = "Plex webhook appears inactive"
	hook.Metadata.Summary = fmt.Sprintf(
		"Detected %d active Plex sessions but no accepted webhook for %s.",
		len(sessions.Sessions), now.Sub(state.LastWebhookAt).Round(time.Second))
	if state.LastWebhookAt.IsZero() {
		hook.Metadata.Summary = fmt.Sprintf(
			"Detected %d active Plex sessions but no accepted webhook since startup.",
			len(sessions.Sessions))
	}

	website.SendData(&website.Request{
		ReqID: mnd.GetID(ctx),
		Route: website.PlexRoute,
		Event: website.EventHook,
		Payload: &website.Payload{
			Snap: c.getMetaSnap(ctx),
			Plex: sessions,
			Load: hook,
		},
		LogMsg: fmt.Sprintf(
			"Plex Webhook Health Alert: active sessions=%d, lastWebhook=%s",
			len(sessions.Sessions), state.LastWebhookAt.Format(time.RFC3339)),
		LogPayload: true,
	})
}
