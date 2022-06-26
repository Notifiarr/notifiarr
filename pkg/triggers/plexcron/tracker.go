package plexcron

import (
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
)

// getSessions interacts with the for loop/channels in runSessionHolder().
func (c *cmd) getSessions(wait bool) (*plex.Sessions, error) {
	c.psMutex.RLock()
	defer c.psMutex.RUnlock()

	if c.sess == nil {
		return nil, common.ErrNoChannel
	}

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
		c.plexSessionTracker(sessions.Sessions, nil)
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
			if sessions == nil || len(sessions.Sessions) < 1 {
				c.plexSessionTracker(currSessions.Sessions, nil)
			} else {
				c.plexSessionTracker(currSessions.Sessions, sessions.Sessions)
			}

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
func (c *cmd) plexSessionTracker(curr, prev []*plex.Session) {
CURRENT:
	for _, currSess := range curr {
		// make sure every current session has a start time.
		currSess.Player.StateTime.Time = time.Now()
		// now check if a current session matches a previous session
		for _, prevSess := range prev {
			if currSess.Session.ID == prevSess.Session.ID {
				// we have a match, check for state change.
				if currSess.Player.State == prevSess.Player.State {
					// since the state is the same, copy the previous start time.
					currSess.Player.StateTime.Time = prevSess.Player.StateTime.Time
				}
				// we found this current session in previous session list, so go to the next one.
				continue CURRENT
			}
		}

		// We found a brand new session. We did not check the session state (yet).
		if c.ClientInfo != nil && c.ClientInfo.Actions.Plex.TrackSess && len(prev) > 0 {
			// We are tracking sessions (no webhooks); send this new session to website.
			c.sendSessionNew(currSess)
		}
	}
}
