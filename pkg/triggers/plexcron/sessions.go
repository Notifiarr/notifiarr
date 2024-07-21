package plexcron

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
)

// sendPlexSessions is fired by a timer if Plex Sessions feature has an interval defined.
func (c *cmd) sendPlexSessions(ctx context.Context, input *common.ActionInput) {
	sessions, err := c.getSessions(ctx, time.Minute)
	if err != nil {
		c.Errorf("Getting Plex sessions: %v", err)
	}

	c.SendData(&website.Request{
		Route:      website.PlexRoute,
		Event:      input.Type,
		Payload:    &website.Payload{Snap: c.getMetaSnap(ctx), Plex: sessions},
		LogMsg:     "Plex Sessions",
		LogPayload: true,
	})
}

// getSessions interacts with the for loop/channels in runSessionHolder().
// The Lock ensures only one request to Plex happens at once.
// Because of the cache two requests may get the same answer.
func (c *cmd) getSessions(ctx context.Context, allowedAge time.Duration) (*plex.Sessions, error) {
	c.Lock()
	defer c.Unlock()

	item := data.Get("plexCurrentSessions")
	if item != nil && time.Now().Add(-allowedAge).Before(item.Time) && item.Data != nil {
		return item.Data.(*plex.Sessions), nil //nolint:forcetypeassert
	}

	deadline, _ := ctx.Deadline()
	start := time.Now()
	timeout := deadline.Sub(start)
	sessions, err := c.Plex.GetSessionsWithContext(ctx)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return &plex.Sessions{Name: c.Plex.Server.Name()}, fmt.Errorf("plex sessions timed out after %s: %w", timeout, err)
	case errors.Is(err, context.Canceled):
		return &plex.Sessions{Name: c.Plex.Server.Name()},
			fmt.Errorf("plex sessions cancelled after %s: %w", time.Since(start), err)
	case err != nil:
		return &plex.Sessions{Name: c.Plex.Server.Name()}, fmt.Errorf("plex sessions: %w", err)
	case item != nil && item.Data != nil:
		c.plexSessionTracker(ctx, sessions, item.Data.(*plex.Sessions)) //nolint:forcetypeassert
	default:
		c.plexSessionTracker(ctx, sessions, nil)
	}

	sessions.Name = c.Plex.Server.Name()

	return sessions, nil
}

// plexSessionTracker checks for state changes between the previous session pull
// and the current session pull. if changes are present, a timestmp is added.
func (c *cmd) plexSessionTracker(ctx context.Context, current, previous *plex.Sessions) {
	now := time.Now()
	info := clientinfo.Get()

	// data.Save("plexPreviousSessions", previous)
	data.Save("plexCurrentSessions", current)

	for _, currSess := range current.Sessions {
		// make sure every session has a start time.
		currSess.Player.StateTime.Time = now

		switch {
		case previous == nil:
			continue // this only happens once.
		case c.checkExistingSession(ctx, currSess, current, previous):
			continue // existing session.
		case currSess.Player.State == playing && info.Actions.Plex.TrackSess:
			// We are tracking sessions (no webhooks); send this brand new session to website.
			c.sendSessionPlaying(ctx, currSess, current, mediaPlay)
		}
	}
}

func (c *cmd) checkExistingSession(ctx context.Context, currSess *plex.Session, current, previous *plex.Sessions) bool {
	// now check if a current session matches a previous session
	for _, prevSess := range previous.Sessions {
		if currSess.Session.ID != prevSess.Session.ID {
			continue
		}

		// we have a match, check for state change.
		if currSess.Player.State == prevSess.Player.State {
			// since the state is the same, copy the previous start time.
			currSess.Player.StateTime.Time = prevSess.Player.StateTime.Time
		} else
		// Check for a session that was paused and is now playing (resumed).
		if ci := clientinfo.Get(); currSess.Player.State == playing &&
			prevSess.Player.State == paused && ci.Actions.Plex.TrackSess {
			// Check if we're tracking sessions. If yes, send this resumed session.
			c.sendSessionPlaying(ctx, currSess, current, mediaResume)
		}

		// we found this current session in previous session list, so go to the next one.
		return true
	}

	return false // session not found in previous list.
}
