package plexcron

import (
	"context"
	"fmt"
	"sync"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

type holder struct {
	sessions *plex.Sessions
	error    error
}

// sendPlexSessions is fired by a timer if Plex Sessions feature has an interval defined.
func (c *cmd) sendPlexSessions(event website.EventType) {
	sessions, err := c.getSessions(false)
	if err != nil {
		c.Errorf("Getting Plex sessions: %v", err)
	}

	c.SendData(&website.Request{
		Route:      website.PlexRoute,
		Event:      event,
		Payload:    &website.Payload{Snap: c.getMetaSnap(), Plex: sessions},
		LogMsg:     "Plex Sessions",
		LogPayload: true,
	})
}

// sendSessionNew is used when the end user does not have or use Plex webhooks.
// They can enable the plex session tracker to send notifications for new sessions.
func (c *cmd) sendSessionNew(session *plex.Session) {
	if err := c.checkPlexAgent(session); err != nil {
		c.Errorf("Failed Plex Request: %v", err)
		return
	}

	c.SendData(&website.Request{
		Route: website.PlexRoute,
		Event: website.EventType(session.Type),
		Payload: &website.Payload{
			Snap: c.getMetaSnap(),
			Plex: &plex.Sessions{Name: c.Plex.Name, Sessions: []*plex.Session{session}},
		},
		LogMsg: fmt.Sprintf("Plex New Session on %s {%s/%s} %s => %s: %s (%s)",
			c.Plex.URL, session.Session.ID, session.SessionKey, session.User.Title,
			session.Type, session.Title, session.Player.State),
		LogPayload: true,
	})
}

// getMetaSnap grabs some basic system info: cpu, memory, username. Gets added to Plex sessions and webhook payloads.
func (c *cmd) getMetaSnap() *snapshot.Snapshot {
	ctx, cancel := context.WithTimeout(context.Background(), c.Snapshot.Timeout.Duration)
	defer cancel()

	var (
		snap = &snapshot.Snapshot{}
		wg   sync.WaitGroup
	)

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = snap.GetCPUSample(ctx)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = snap.GetMemoryUsage(ctx)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = snap.GetLocalData(ctx)
	}()

	wg.Wait()

	return snap
}
