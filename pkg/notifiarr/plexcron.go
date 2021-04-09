package notifiarr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
)

// This cron tab runs every 10-60 minutes to send a report of who's currently watching.
func (c *Config) startPlexCron() {
	if c.Plex == nil || c.Plex.Interval.Duration == 0 || c.Plex.URL == "" || c.Plex.Token == "" {
		return
	}

	// Add a little splay to the timers to not hit plex at the same time too often.
	timer1 := time.NewTicker(c.Plex.Interval.Duration + 139*time.Millisecond)
	timer2 := time.NewTicker(time.Minute + 179*time.Millisecond)
	c.stopPlex = make(chan struct{})

	defer func() {
		timer1.Stop()
		close(c.stopPlex)
		c.stopPlex = nil
	}()

	c.Printf("==> Plex Sessions Collection Started, URL: %s, interval: %v, timeout: %v, webhook cooldown: %v",
		c.Plex.URL, c.Plex.Interval, c.Plex.Timeout, c.Plex.Cooldown)

	if c.Plex.MoviesPC != 0 || c.Plex.SeriesPC != 0 {
		defer timer2.Stop()
		c.Printf("==> Plex Completed Items Started, URL: %s, interval: 1m, timeout: %v movies: %d%%, series: %d%%",
			c.Plex.URL, c.Plex.Timeout, c.Plex.MoviesPC, c.Plex.SeriesPC)
	} else {
		timer2.Stop() // nothing to check, so turn off this timer.
	}

	c.plexCron(timer1, timer2)
}

// Do not call this directly. Called from above, only.
func (c *Config) plexCron(timer1, timer2 *time.Ticker) {
	sent := make(map[string]struct{})

	for {
		select {
		case <-timer1.C:
			if body, err := c.SendMeta(PlexCron, c.URL, nil, 0); err != nil {
				c.Errorf("Sending Plex Session to %s: %v: %v", c.URL, err, string(body))
			} else if fields := strings.Split(string(body), `"`); len(fields) > 3 { //nolint:gomnd
				c.Printf("Plex Sessions sent to %s, sending again in %s, reply: %s", c.URL, c.Plex.Interval, fields[3])
			} else {
				c.Printf("Plex Sessions sent to %s, sending again in %s, reply: %s", c.URL, c.Plex.Interval, string(body))
			}
		case <-timer2.C:
			c.checkForFinishedItems(sent)
		case <-c.stopPlex:
			return
		}
	}
}

// Statuses for an item being played on Plex.
const (
	statusIgnoring = "ignoring"
	statusWatching = "watching"
	statusSending  = "sending"
	statusError    = "error"
	statusSent     = "sent"
)

const (
	movie   = "movie"
	episode = "episode"
)

// This cron tab runs every minute to send a report when a user gets to the end of a movie or tv show.
// This is basically a hack to "watch" Plex for when an active item gets to around 90% complete.
// This usually means the user has finished watching the item and we can send a "done" notice.
// Plex does not send a webhook or identify in any other way when an item is "finished".
func (c *Config) checkForFinishedItems(sent map[string]struct{}) {
	sessions, err := c.Plex.GetSessions()
	if err != nil {
		c.Errorf("[PLEX] Getting Sessions from %s: %v", c.Plex.URL, err)
		return
	} else if len(sessions) == 0 {
		// c.Debugf("[PLEX] No Sessions Collected from %s", c.Plex.URL)
		return
	}

	for _, s := range sessions {
		var (
			_, ok = sent[s.Session.ID+s.SessionKey]
			pct   = s.ViewOffset / s.Duration * 100
			msg   = statusSent
		)

		if !ok { // ok means we already sent a message for this session.
			msg = c.checkSessionDone(s, pct)
			if strings.HasPrefix(msg, statusSending) {
				sent[s.Session.ID+s.SessionKey] = struct{}{}
			}
		}

		// nolint:lll
		// [DEBUG] 2021/04/03 06:05:11 [PLEX] https://plex.domain.com {dsm195u1jurq7w1ejlh6pmr9/34} username => episode: Hard Facts: Vandalism and Vulgarity (playing) 8.1%
		// [DEBUG] 2021/04/03 06:00:39 [PLEX] https://plex.domain.com {dsm195u1jurq7w1ejlh6pmr9/33} username => movie: Come True (playing) 81.3%
		if strings.HasPrefix(msg, statusSending) || strings.HasPrefix(msg, statusError) {
			c.Printf("[PLEX] %s {%s/%s} %s => %s: %s (%s) %.1f%% (%s)",
				c.Plex.URL, s.Session.ID, s.SessionKey, s.User.Title,
				s.Type, s.Title, s.Player.State, pct, msg)
		} else {
			c.Debugf("[PLEX] %s {%s/%s} %s => %s: %s (%s) %.1f%% (%s)",
				c.Plex.URL, s.Session.ID, s.SessionKey, s.User.Title,
				s.Type, s.Title, s.Player.State, pct, msg)
		}
	}
}

func (c *Config) checkSessionDone(s *plex.Session, pct float64) string {
	switch {
	case c.Plex.MoviesPC > 0 && strings.EqualFold(s.Type, movie):
		if pct < float64(c.Plex.MoviesPC) {
			return statusWatching
		}

		return c.sendSessionDone(s)
	case c.Plex.SeriesPC > 0 && strings.EqualFold(s.Type, episode):
		if pct < float64(c.Plex.SeriesPC) {
			return statusWatching
		}

		return c.sendSessionDone(s)
	default:
		return statusIgnoring
	}
}

func (c *Config) sendSessionDone(s *plex.Session) string {
	if err := c.checkPlexAgent(s); err != nil {
		return statusError + ": " + err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Snap.Timeout.Duration)
	snap := c.GetMetaSnap(ctx)
	cancel() //nolint:wsl

	_, _, err := c.SendData(c.URL, &Payload{
		Type: "plex_session_complete_" + s.Type,
		Snap: snap,
		Plex: &plex.Sessions{
			Name:       c.Plex.Name,
			Sessions:   []*plex.Session{s},
			AccountMap: strings.Split(c.Plex.AccountMap, "|"),
		},
	})
	if err != nil {
		return statusError + ": sending to " + c.URL + ": " + err.Error()
	}

	return statusSending + " to " + c.URL
}

func (c *Config) checkPlexAgent(s *plex.Session) error {
	if !strings.Contains(s.GUID, "plex://") || s.Key == "" {
		return nil
	}

	sections, err := c.Plex.GetPlexSectionKey(s.Key)
	if err != nil {
		return fmt.Errorf("getting plex key %s: %w", s.Key, err)
	}

	for _, section := range sections.Metadata {
		if section.RatingKey == s.RatingKey {
			s.GuID = section.GuID
			return nil
		}
	}

	return nil
}
