package plexcron

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
)

// This cron tab runs every minute to send a report when a user gets to the end of a movie or tv show.
// This is basically a hack to "watch" Plex for when an active item gets to around 90% complete.
// This usually means the user has finished watching the item and we can send a "done" notice.
// Plex does not send a webhook or identify in any other way when an item is "finished".
func (c *cmd) checkForFinishedItems(ctx context.Context, _ *common.ActionInput) {
	sessions, err := c.getSessions(ctx, time.Second)
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

		// Make sure we didn't already send this session.
		if _, ok := c.sent[session.Session.ID+session.SessionKey]; !ok {
			msg = c.checkSessionDone(ctx, session, pct)
		}

		//nolint:lll
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
func (c *cmd) checkSessionDone(ctx context.Context, session *plex.Session, pct float64) string {
	ci := clientinfo.Get()

	switch cfg := ci.Actions.Plex; {
	case session.Duration == 0:
		return statusIgnoring
	case session.Player.State != playing:
		return statusPaused
	case cfg.MoviesPC > 0 && website.EventType(session.Type) == website.EventMovie:
		if pct < float64(cfg.MoviesPC) {
			return statusWatching
		}

		return c.sendSessionDone(ctx, session)
	case cfg.SeriesPC > 0 && website.EventType(session.Type) == website.EventEpisode:
		if pct < float64(cfg.SeriesPC) {
			return statusWatching
		}

		return c.sendSessionDone(ctx, session)
	default:
		return statusIgnoring
	}
}

// sendSessionDone is the last method to run that sends a finished session to the website.
func (c *cmd) sendSessionDone(ctx context.Context, session *plex.Session) string {
	if err := c.checkPlexAgent(ctx, session); err != nil {
		return statusError + ": " + err.Error()
	}

	c.SendData(&website.Request{
		Route: website.PlexRoute,
		Event: website.EventType(session.Type),
		Payload: &website.Payload{
			Snap: c.getMetaSnap(ctx),
			Plex: &plex.Sessions{Name: c.Plex.Server.Name(), Sessions: []*plex.Session{session}},
		},
		LogMsg:     "Plex Completed Sessions",
		LogPayload: true,
		ErrorsOnly: !c.DebugEnabled(),
	})

	c.sent[session.Session.ID+session.SessionKey] = struct{}{}

	return statusSending
}

// checkPlexAgent checks the plex agent and makes another request to find the section key.
// This is because Plex servers using the Plex Agent do not provide the show Title in the session.
func (c *cmd) checkPlexAgent(ctx context.Context, session *plex.Session) error {
	if !strings.Contains(session.GUID, "plex://") || session.Key == "" {
		return nil
	}

	sections, err := c.Plex.GetPlexSectionKeyWithContext(ctx, session.Key)
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
