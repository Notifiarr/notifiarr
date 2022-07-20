package plexcron

import (
	"context"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
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

// New configures the library.
func New(config *common.Config, plex *plex.Server) *Action {
	return &Action{
		cmd: &cmd{
			Config: config,
			Plex:   plex,
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

	c.sess = make(chan time.Time, 1)
	c.sessr = make(chan *holder)

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
			Fn:   c.checkForFinishedItems,
			T:    time.NewTicker(time.Minute + 179*time.Millisecond),
		})
	}
}

// SendWebhook is called in a go routine after a plex media.play webhook is received.
func (a *Action) SendWebhook(hook *plex.IncomingWebhook) {
	a.cmd.sendWebhook(hook)
}

func (c *cmd) sendWebhook(hook *plex.IncomingWebhook) {
	sessions := &plex.Sessions{Name: c.Plex.Name}
	// If NoActivity=false, then grab sessions.
	if c.ClientInfo != nil && !c.ClientInfo.Actions.Plex.NoActivity {
		var err error
		if sessions, err = c.getSessions(true); err != nil {
			c.Errorf("Getting Plex sessions: %v", err)
		}
	}

	c.SendData(&website.Request{
		Route:      website.PlexRoute,
		Event:      website.EventHook,
		Payload:    &website.Payload{Snap: c.getMetaSnap(), Load: hook, Plex: sessions},
		LogMsg:     "Plex Webhook (and sessions)",
		LogPayload: true,
	})
}

// GetSessions returns the plex sessions. This uses a channel so concurrent requests are avoided.
// Passing wait=true makes sure the results are current. Waits up to 10 seconds before requesting.
// Passing wait=false will allow for sessions up to 10 seconds old. This may return faster.
func (a *Action) GetSessions(wait bool) (*plex.Sessions, error) {
	return a.cmd.getSessions(wait)
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
