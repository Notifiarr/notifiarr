package plexcron

import (
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
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

// getSessions interacts with the for loop/channels in runSessionHolder().
func (c *cmd) getSessions(wait bool) (*plex.Sessions, error) {
	if wait {
		c.sess <- time.Now().Add(c.ClientInfo.Actions.Plex.Delay.Duration)
	} else {
		c.sess <- time.Now().Add(-c.ClientInfo.Actions.Plex.Delay.Duration)
	}

	s := <-c.sessr

	return s.sessions, s.error
}

// runSessionHolder runs a session holder routine. Call Run() first.
func (c *cmd) runSessionHolder() {
	defer func() {
		defer c.CapturePanic()
		close(c.sessr) // indicate we're done.
	}()

	sessions, err := c.Plex.GetSessions() // err not used until for loop.
	if sessions != nil {
		sessions.Updated.Time = time.Now()
		c.plexSessionTracker(sessions, nil)
	}

	for waitUntil := range c.sess {
		if sessions != nil && err == nil && sessions.Updated.After(waitUntil) {
			c.sessr <- &holder{sessions: sessions}
			continue
		}

		if t := time.Until(waitUntil); t > 0 {
			time.Sleep(t)
		}

		var currSessions *plex.Sessions // so we can update the error.
		if currSessions, err = c.Plex.GetSessions(); currSessions != nil {
			c.plexSessionTracker(currSessions, sessions)
			delSessions(sessions) // memory cleanup.

			currSessions.Updated.Time = time.Now()
			sessions = currSessions
		}

		c.sessr <- &holder{sessions: sessions, error: err}
	}
}

// delSessions sets pointers to nil. Should free up memory.
func delSessions(sess *plex.Sessions) {
	for idx := range sess.Sessions {
		sess.Sessions[idx] = nil
	}
}

// plexSessionTracker checks for state changes between the previous session pull
// and the current session pull. if changes are present, a timestmp is added.
func (c *cmd) plexSessionTracker(current, previous *plex.Sessions) {
	now := time.Now()

	for _, currSess := range current.Sessions {
		// make sure every session has a start time.
		currSess.Player.StateTime.Time = now

		switch {
		case previous == nil:
			continue // this only happens once.
		case c.checkExistingSession(currSess, current, previous):
			continue // existing session.
		case currSess.Player.State == playing && c.ClientInfo != nil && c.ClientInfo.Actions.Plex.TrackSess:
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
			c.ClientInfo != nil && c.ClientInfo.Actions.Plex.TrackSess {
			c.sendSessionPlaying(currSess, current, mediaResume)
		}

		// we found this current session in previous session list, so go to the next one.
		return true
	}

	return false // session not found in previous list.
}
