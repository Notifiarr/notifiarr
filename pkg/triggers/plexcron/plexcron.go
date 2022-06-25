package plexcron

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
)

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	Plex    *plex.Server
	sess    chan time.Time      // Return Plex Sessions
	sessr   chan *holder        // Session Return Channel
	sent    map[string]struct{} // Tracks Finished sessions already sent.
	psMutex sync.RWMutex        // Locks plex session thread.
}

const TrigPlexSessions common.TriggerName = "Gathering and sending Plex Sessions."

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

// New configures the library.
func New(config *common.Config, plex *plex.Server) *Action {
	return &Action{
		cmd: &cmd{
			Config: config,
			Plex:   plex,
			sess:   make(chan time.Time, 1),
			sessr:  make(chan *holder),
			sent:   make(map[string]struct{}),
		},
	}
}

// SendPlexSessions sends plex sessions in a go routine through a channel.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(event, TrigPlexSessions)
}

// Run initializes the library.
func (a *Action) Run() {
	a.cmd.run()
}

func (c *cmd) run() {
	if !c.Plex.Configured() || c.ClientInfo == nil {
		return
	}

	go c.runSessionHolder()

	var ticker *time.Ticker

	cfg := c.ClientInfo.Actions.Plex
	if cfg.Interval.Duration > 0 {
		// Add a little splay to the timers to not hit plex at the same time too often.
		ticker = time.NewTicker(cfg.Interval.Duration + 139*time.Millisecond)
		c.Printf("==> Plex Sessions Collection Started, URL: %s, interval:%s timeout:%s webhook_cooldown:%v delay:%v",
			c.Plex.URL, cfg.Interval, c.Plex.Timeout, cfg.Cooldown, cfg.Delay)
	}

	c.Add(&common.Action{
		Name: TrigPlexSessions,
		Fn:   c.sendPlexSessions,
		C:    make(chan website.EventType, 1),
		T:    ticker,
	})

	if cfg.MoviesPC != 0 || cfg.SeriesPC != 0 || cfg.TrackSess {
		c.Printf("==> Plex Sessions Tracker Started, URL: %s, interval:1m timeout:%s movies:%d%% series:%d%% play:%v",
			c.Plex.URL, c.Plex.Timeout, cfg.MoviesPC, cfg.SeriesPC, cfg.TrackSess)
		c.Add(&common.Action{
			Name: "Checking Plex for completed sessions.",
			Hide: true, // do not log this one.
			Fn:   c.checkPlexFinishedItems,
			T:    time.NewTicker(time.Minute + 179*time.Millisecond),
		})
	}
}

// sendPlexSessions is fired by a timer if plex monitoring is enabled.
func (c *cmd) sendPlexSessions(event website.EventType) {
	if !c.Plex.Configured() {
		return
	}

	c.collectSessions(event, nil)
}

// CollectSessions is called in a go routine after a plex media.play webhook.
// This reaches back into Plex, asks for sessions and then sends the whole
// payloads (incoming webhook and sessions) over to notifiarr.com.
// SendMeta also collects system snapshot info, so a lot happens here.
func (a *Action) CollectSessions(event website.EventType, hook *plex.IncomingWebhook) {
	a.cmd.collectSessions(event, hook)
}

func (c *cmd) collectSessions(event website.EventType, hook *plex.IncomingWebhook) {
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
func (a *Action) GetSessions(wait bool) (*plex.Sessions, error) {
	return a.cmd.getSessions(wait)
}

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
func (c *cmd) runSessionHolder() { //nolint:cyclop
	defer c.CapturePanic()

	sessions, err := c.Plex.GetSessions() // err not used until for loop.
	if sessions != nil {
		sessions.Updated.Time = time.Now()
		if len(sessions.Sessions) > 0 {
			c.plexSessionTracker(sessions.Sessions, nil)
		}
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

	close(c.sessr) // indicate we're done.
}

// delSessions sets pointers to nil. Should free up memory.
func delSessions(sess *plex.Sessions) {
	for idx := range sess.Sessions {
		sess.Sessions[idx] = nil
	}

	sess = nil
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

// sendSessionNew is used when the end user does not have or use Plex webhooks.
// They can enable the plex session tracker to send notifications for new sessions.
func (c *cmd) sendSessionNew(session *plex.Session) {
	if err := c.checkPlexAgent(session); err != nil {
		c.Errorf("Failed Plex Request: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Snapshot.Timeout.Duration)
	snap := c.getMetaSnap(ctx)
	cancel() //nolint:wsl

	c.SendData(&website.Request{
		Route: website.PlexRoute,
		Event: website.EventType(session.Type),
		Payload: &website.Payload{
			Snap: snap,
			Plex: &plex.Sessions{Name: c.Plex.Name, Sessions: []*plex.Session{session}},
		},
		LogMsg: fmt.Sprintf("Plex New Session on %s {%s/%s} %s => %s: %s (%s) (sending)",
			c.Plex.URL, session.Session.ID, session.SessionKey, session.User.Title,
			session.Type, session.Title, session.Player.State),
		LogPayload: true,
	})

}

// This cron tab runs every minute to send a report when a user gets to the end of a movie or tv show.
// This is basically a hack to "watch" Plex for when an active item gets to around 90% complete.
// This usually means the user has finished watching the item and we can send a "done" notice.
// Plex does not send a webhook or identify in any other way when an item is "finished".
func (c *cmd) checkPlexFinishedItems(website.EventType) {
	sessions, err := c.getSessions(false)
	if err != nil {
		c.Errorf("[PLEX] Getting Sessions from %s: %v", c.Plex.URL, err)
		return
	} else if len(sessions.Sessions) == 0 {
		c.Debugf("[PLEX] No Sessions Collected from %s", c.Plex.URL)
		return
	}

	for _, session := range sessions.Sessions {
		var (
			pct = session.ViewOffset / session.Duration * 100
			msg = statusSent
		)

		if _, ok := c.sent[session.Session.ID+session.SessionKey]; !ok {
			msg = c.checkSessionDone(session, pct)
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

// checkSessionDone checks a session's data to see if it is considered finished.
func (c *cmd) checkSessionDone(session *plex.Session, pct float64) string {
	cfg := c.ClientInfo.Actions.Plex
	switch {
	case session.Duration == 0:
		return statusIgnoring
	case session.Player.State != "playing":
		return statusPaused
	case cfg.MoviesPC > 0 && website.EventType(session.Type) == website.EventMovie:
		if pct < float64(cfg.MoviesPC) {
			return statusWatching
		}

		return c.sendSessionDone(session)
	case cfg.SeriesPC > 0 && website.EventType(session.Type) == website.EventEpisode:
		if pct < float64(cfg.SeriesPC) {
			return statusWatching
		}

		return c.sendSessionDone(session)
	default:
		return statusIgnoring
	}
}

// sendSessionDone is the last method to run that sends a finished session to the website.
func (c *cmd) sendSessionDone(session *plex.Session) string {
	if err := c.checkPlexAgent(session); err != nil {
		return statusError + ": " + err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Snapshot.Timeout.Duration)
	snap := c.getMetaSnap(ctx)
	cancel() //nolint:wsl

	c.SendData(&website.Request{
		Route: website.PlexRoute,
		Event: website.EventType(session.Type),
		Payload: &website.Payload{
			Snap: snap,
			Plex: &plex.Sessions{Name: c.Plex.Name, Sessions: []*plex.Session{session}},
		},
		LogMsg:     "Plex Completed Sessions",
		LogPayload: true,
		ErrorsOnly: true,
	})

	c.sent[session.Session.ID+session.SessionKey] = struct{}{}

	return statusSending
}

// checkPlexAgent checks the plex agent and makes another request to find the section key.
// This is because Plex servers using the Plex Agent do not provide the show Title in the session.
func (c *cmd) checkPlexAgent(session *plex.Session) error {
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

// Stop the Plex session holder.
func (a *Action) Stop() {
	a.cmd.stop()
}

func (c *cmd) stop() {
	c.psMutex.Lock()
	defer c.psMutex.Unlock()

	if c.sess == nil {
		return
	}

	close(c.sess)
	<-c.sessr // wait for session holder to return
	c.sessr = nil
	c.sess = nil
}
