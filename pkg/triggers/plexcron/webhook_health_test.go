package plexcron

import (
	"testing"
	"time"
)

func TestShouldAlertForStaleWebhook(t *testing.T) {
	now := time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC)

	cfg := webhookHealthConfig{
		Enabled:       true,
		StaleAfter:    20 * time.Minute,
		StartupGrace:  10 * time.Minute,
		AlertCooldown: 6 * time.Hour,
	}

	tests := []struct {
		name           string
		activeSessions int
		state          webhookHealthState
		want           bool
	}{
		{
			name:           "alerts when sessions active and webhook stale",
			activeSessions: 2,
			state: webhookHealthState{
				StartAt:       now.Add(-2 * time.Hour),
				LastWebhookAt: now.Add(-25 * time.Minute),
			},
			want: true,
		},
		{
			name:           "does not alert when no sessions",
			activeSessions: 0,
			state: webhookHealthState{
				StartAt:       now.Add(-2 * time.Hour),
				LastWebhookAt: now.Add(-25 * time.Minute),
			},
			want: false,
		},
		{
			name:           "does not alert when webhook is recent",
			activeSessions: 1,
			state: webhookHealthState{
				StartAt:       now.Add(-2 * time.Hour),
				LastWebhookAt: now.Add(-5 * time.Minute),
			},
			want: false,
		},
		{
			name:           "does not alert during startup grace",
			activeSessions: 3,
			state: webhookHealthState{
				StartAt:       now.Add(-5 * time.Minute),
				LastWebhookAt: time.Time{},
			},
			want: false,
		},
		{
			name:           "does not alert during cooldown",
			activeSessions: 3,
			state: webhookHealthState{
				StartAt:       now.Add(-2 * time.Hour),
				LastWebhookAt: now.Add(-25 * time.Minute),
				LastAlertAt:   now.Add(-2 * time.Hour),
			},
			want: false,
		},
		{
			name:           "alerts if no webhook was ever seen and grace has passed",
			activeSessions: 1,
			state: webhookHealthState{
				StartAt:       now.Add(-2 * time.Hour),
				LastWebhookAt: time.Time{},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldAlertForStaleWebhook(now, tc.activeSessions, cfg, tc.state)
			if got != tc.want {
				t.Fatalf("shouldAlertForStaleWebhook() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDefaultWebhookHealthConfig(t *testing.T) {
	cfg := defaultWebhookHealthConfig()

	if !cfg.Enabled {
		t.Fatal("expected default webhook health config to be enabled")
	}

	if cfg.StaleAfter <= 0 || cfg.StartupGrace <= 0 || cfg.AlertCooldown <= 0 {
		t.Fatalf("expected positive durations, got stale=%s grace=%s cooldown=%s",
			cfg.StaleAfter, cfg.StartupGrace, cfg.AlertCooldown)
	}
}
