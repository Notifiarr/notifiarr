package plexcron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// sendPlexSessions is fired by a timer if plex monitoring is enabled.
func (c *cmd) sendPlexSessions(event website.EventType) {
	c.collectSessions(event, nil)
}

// collectSessions is kicked off by the webserver in go routine.
// It's also called by the plex cron (with webhook set to nil).
// This runs after Plex drops off a webhook telling us someone did something.
// This gathers cpu/ram, and waits 10 seconds, then grabs plex sessions.
// It's all POSTed to notifiarr. May be used with a nil Webhook.
func (c *cmd) collectSessions(event website.EventType, hook *plex.IncomingWebhook) {
	extra := time.Second
	if hook != nil {
		extra = c.ClientInfo.Actions.Plex.Delay.Duration
	}

	ctx, cancel := context.WithTimeout(context.Background(), extra+c.Snapshot.Timeout.Duration)
	defer cancel()

	var (
		payload = &website.Payload{Load: hook, Plex: &plex.Sessions{Name: c.Plex.Name}}
		wg      sync.WaitGroup
	)

	wg.Add(1)

	go func() {
		payload.Snap = c.getMetaSnap(ctx)
		wg.Done() // nolint:wsl
	}()

	if hook == nil || !c.ClientInfo.Actions.Plex.NoActivity {
		var err error
		if payload.Plex, err = c.getSessions(hook != nil); err != nil {
			c.Errorf("Getting Plex sessions: %v", err)
		}
	}

	wg.Wait()

	msg := ""
	if hook != nil {
		msg = " (and webhook)"
	}

	c.SendData(&website.Request{
		Route:      website.PlexRoute,
		Event:      event,
		Payload:    payload,
		LogMsg:     fmt.Sprintf("Plex Sessions%s", msg),
		LogPayload: true,
	})
}

// getMetaSnap grabs some basic system info: cpu, memory, username.
func (c *cmd) getMetaSnap(ctx context.Context) *snapshot.Snapshot {
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

	wg.Add(1)

	go func() {
		rep <- snap.GetCPUSample(ctx)

		wg.Done()
	}()

	wg.Add(1)

	go func() {
		rep <- snap.GetMemoryUsage(ctx)

		wg.Done()
	}()

	wg.Add(1)

	go func() {
		for _, err := range snap.GetLocalData(ctx) {
			rep <- err
		}

		wg.Done()
	}()

	wg.Wait()

	return snap
}
