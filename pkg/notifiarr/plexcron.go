package notifiarr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
)

// Statuses for an item being played on Plex.
const (
	statusIgnoring = "ignoring"
	statusPaused   = "ignoring, paused"
	statusWatching = "watching"
	statusSending  = "sending"
	statusError    = "error"
	statusSent     = "sent"
)

type holder struct {
	sessions *plex.Sessions
	error    error
}

var ErrNoChannel = fmt.Errorf("no channel to send session request")

// SendPlexSessions sends plex sessions in a go routine through a channel.
func (t *Triggers) SendPlexSessions(event EventType) {
	t.exec(event, TrigPlexSessions)
}

// sendPlexSessions is fired by a timer if plex monitoring is enabled.
func (c *Config) sendPlexSessions(event EventType) {
	if !c.Plex.Configured() {
		return
	}

	c.collectSessions(event, nil)
}

// collectSessions is called in a go routine after a plex media.play webhook.
// This reaches back into Plex, asks for sessions and then sends the whole
// payloads (incoming webhook and sessions) over to notifiarr.com.
// SendMeta also collects system snapshot info, so a lot happens here.
func (c *Config) collectSessions(event EventType, hook *plexIncomingWebhook) {
	wait := false
	msg := ""

	if hook != nil {
		wait = true
		msg = " (and webhook)"
	}

	if resp, err := c.sendPlexMeta(event, hook, wait); err != nil {
		c.Errorf("[%s requested] Sending Plex Sessions%s to Notifiarr: %v%v", event, msg, err, resp)
	} else {
		c.Printf("[%s requested] Plex Sessions%s sent to Notifiar.%v", event, msg, resp)
	}
}

// GetSessions returns the plex sessions. This uses a channel so concurrent requests are avoided.
// Passing wait=true makes sure the results are current. Waits up to 10 seconds before requesting.
// Passing wait=false will allow for sessions up to 10 seconds old. This may return faster.
func (c *Config) GetSessions(wait bool) (*plex.Sessions, error) {
	if c.Trigger.sess == nil {
		return nil, ErrNoChannel
	}

	if wait {
		c.Trigger.sess <- time.Now().Add(c.Plex.Delay.Duration)
	} else {
		c.Trigger.sess <- time.Now().Add(-c.Plex.Delay.Duration)
	}

	s := <-c.Trigger.sessr

	return s.sessions, s.error
}

func (c *Config) runSessionHolder() { //nolint:cyclop
	defer c.CapturePanic()

	sessions, err := c.Plex.GetSessions() // err not used until for loop.
	if sessions != nil {
		sessions.Updated.Time = time.Now()
		if len(sessions.Sessions) > 0 {
			c.plexSessionTracker(sessions.Sessions, nil)
		}
	}

	c.Trigger.sessr = make(chan *holder)
	defer close(c.Trigger.sessr)

	for waitUntil := range c.Trigger.sess {
		if sessions != nil && err == nil && sessions.Updated.After(waitUntil) {
			c.Trigger.sessr <- &holder{sessions: sessions}
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

			currSessions.Updated.Time = time.Now()
			sessions = currSessions
		}

		c.Trigger.sessr <- &holder{sessions: sessions, error: err}
	}
}

// plexSessionTracker checks for state changes between the previous session pull
// and the current session pull. if changes are present, a timestmp is added.
func (c *Config) plexSessionTracker(curr, prev []*plex.Session) {
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
	}
}

// This cron tab runs every minute to send a report when a user gets to the end of a movie or tv show.
// This is basically a hack to "watch" Plex for when an active item gets to around 90% complete.
// This usually means the user has finished watching the item and we can send a "done" notice.
// Plex does not send a webhook or identify in any other way when an item is "finished".
func (c *Config) checkPlexFinishedItems(sent map[string]struct{}) {
	sessions, err := c.GetSessions(false)
	if err != nil {
		c.Errorf("[PLEX] Getting Sessions from %s: %v", c.Plex.URL, err)
		return
	} else if len(sessions.Sessions) == 0 {
		c.Debugf("[PLEX] No Sessions Collected from %s", c.Plex.URL)
		return
	}

	for _, session := range sessions.Sessions {
		var (
			_, ok = sent[session.Session.ID+session.SessionKey]
			pct   = session.ViewOffset / session.Duration * 100
			msg   = statusSent
		)

		if !ok { // ok means we already sent a message for this session.
			msg = c.checkSessionDone(session, pct)
			if strings.HasPrefix(msg, statusSending) {
				sent[session.Session.ID+session.SessionKey] = struct{}{}
			}
		}

		// nolint:lll
		// [DEBUG] 2021/04/03 06:05:11 [PLEX] https://plex.domain.com {dsm195u1jurq7w1ejlh6pmr9/34} username => episode: Hard Facts: Vandalism and Vulgarity (playing) 8.1%
		// [DEBUG] 2021/04/03 06:00:39 [PLEX] https://plex.domain.com {dsm195u1jurq7w1ejlh6pmr9/33} username => movie: Come True (playing) 81.3%
		if strings.HasPrefix(msg, statusSending) || strings.HasPrefix(msg, statusError) {
			c.Printf("[PLEX] %s {%s/%s} %s => %s: %s (%s) %.1f%% (%s)",
				c.Plex.URL, session.Session.ID, session.SessionKey, session.User.Title,
				session.Type, session.Title, session.Player.State, pct, msg)
		} else {
			c.Debugf("[PLEX] %s {%s/%s} %s => %s: %s (%s) %.1f%% (%s)",
				c.Plex.URL, session.Session.ID, session.SessionKey, session.User.Title,
				session.Type, session.Title, session.Player.State, pct, msg)
		}
	}
}

func (c *Config) checkSessionDone(session *plex.Session, pct float64) string {
	switch {
	case session.Duration == 0:
		return statusIgnoring
	case session.Player.State != "playing":
		return statusPaused
	case c.Plex.MoviesPC > 0 && EventType(session.Type) == EventMovie:
		if pct < float64(c.Plex.MoviesPC) {
			return statusWatching
		}

		return c.sendSessionDone(session)
	case c.Plex.SeriesPC > 0 && EventType(session.Type) == EventEpisode:
		if pct < float64(c.Plex.SeriesPC) {
			return statusWatching
		}

		return c.sendSessionDone(session)
	default:
		return statusIgnoring
	}
}

func (c *Config) sendSessionDone(session *plex.Session) string {
	if err := c.checkPlexAgent(session); err != nil {
		return statusError + ": " + err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Snap.Timeout.Duration)
	snap := c.getMetaSnap(ctx)
	cancel() //nolint:wsl

	route := PlexRoute.Path(EventType(session.Type))

	_, err := c.SendData(route, &Payload{
		Snap: snap,
		Plex: &plex.Sessions{Name: c.Plex.Name, Sessions: []*plex.Session{session}},
	}, true)
	if err != nil {
		return statusError + ": sending to " + route + ": " + err.Error()
	}

	return statusSending
}

func (c *Config) checkPlexAgent(session *plex.Session) error {
	if !strings.Contains(session.GUID, "plex://") || session.Key == "" {
		return nil
	}

	sections, err := c.Plex.GetPlexSectionKey(session.Key)
	if err != nil {
		return fmt.Errorf("getting plex key %s: %w", session.Key, err)
	}

	for _, section := range sections.Metadata {
		if section.RatingKey == session.RatingKey {
			session.GuID = section.GuID
			return nil
		}
	}

	return nil
}
