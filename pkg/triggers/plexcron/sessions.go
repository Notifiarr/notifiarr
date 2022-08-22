package plexcron

import (
	"fmt"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// sendPlexSessions is fired by a timer if Plex Sessions feature has an interval defined.
func (c *cmd) sendPlexSessions(event website.EventType) {
	sessions, err := c.getSessions(time.Minute)
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

// getSessions interacts with the for loop/channels in runSessionHolder().
// The Lock ensures only one request to Plex happens at once.
// Because of the cache two requests may get the same answer.
func (c *cmd) getSessions(allowedAge time.Duration) (*plex.Sessions, error) {
	c.Lock()
	defer c.Unlock()

	item := data.Get("plexCurrentSessions")
	if item != nil && time.Now().Add(-allowedAge).Before(item.Time) && item.Data != nil {
		return item.Data.(*plex.Sessions), nil //nolint:forcetypeassert
	}

	sessions, err := c.Plex.GetSessions()

	switch {
	case err != nil:
		return &plex.Sessions{Name: c.Plex.Name}, fmt.Errorf("plex sessions: %w", err)
	case item != nil && item.Data != nil:
		c.plexSessionTracker(sessions, item.Data.(*plex.Sessions)) //nolint:forcetypeassert
	default:
		c.plexSessionTracker(sessions, nil)
	}

	sessions.Name = c.Plex.Name

	return sessions, nil
}

// plexSessionTracker checks for state changes between the previous session pull
// and the current session pull. if changes are present, a timestmp is added.
func (c *cmd) plexSessionTracker(current, previous *plex.Sessions) {
	now := time.Now()

	// data.Save("plexPreviousSessions", previous)
	data.Save("plexCurrentSessions", current)

	for _, currSess := range current.Sessions {
		// make sure every session has a start time.
		currSess.Player.StateTime.Time = now

		switch {
		case previous == nil:
			continue // this only happens once.
		case c.checkExistingSession(currSess, current, previous):
			continue // existing session.
		case currSess.Player.State == playing && c.ClientInfo.Actions.Plex.TrackSess:
			// We are tracking sessions (no webhooks); send this brand new session to website.
			c.sendSessionPlaying(currSess, current, mediaPlay)
		}
	}
}

func (c *cmd) checkExistingSession(currSess *plex.Session, current, previous *plex.Sessions) bool {
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
		if currSess.Player.State == playing && prevSess.Player.State == paused &&
			// Check if we're tracking sessions. If yes, send this resumed session.
			c.ClientInfo.Actions.Plex.TrackSess {
			c.sendSessionPlaying(currSess, current, mediaResume)
		}

		// we found this current session in previous session list, so go to the next one.
		return true
	}

	return false // session not found in previous list.
}
