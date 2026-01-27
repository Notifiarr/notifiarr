package plexcron

import (
	"context"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
)

const (
	randomMilliseconds  = 3000
	randomMilliseconds2 = 400
)

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	Plex []apps.Plex
	sent map[string]struct{} // Tracks Finished sessions already sent.
	sync.Mutex
}

const (
	TrigPlexSessions      common.TriggerName = "Gathering and sending Plex Sessions."
	TrigPlexSessionsCheck common.TriggerName = "Checking Plex for completed sessions."
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

// New configures the library.
func New(config *common.Config, plex []apps.Plex) *Action {
	return &Action{
		cmd: &cmd{
			Config: config,
			Plex:   plex,
			sent:   make(map[string]struct{}),
		},
	}
}

// Enabled returns true if at least one Plex server is configured.
func (a *Action) Enabled() bool {
	return len(a.cmd.Plex) > 0
}

// First returns the first Plex instance, or nil if none configured.
func (a *Action) First() *apps.Plex {
	if len(a.cmd.Plex) == 0 {
		return nil
	}

	return &a.cmd.Plex[0]
}

// Send sends plex sessions in a go routine through a channel.
func (a *Action) Send(event website.EventType) {
	a.cmd.Exec(&common.ActionInput{Type: event}, TrigPlexSessions)
}

// Create initializes the library.
func (a *Action) Create() {
	a.cmd.run()
}

func (c *cmd) run() {
	info := clientinfo.Get()
	if len(c.Plex) == 0 || info == nil {
		return
	}

	var dur time.Duration

	cfg := info.Actions.Plex
	if cfg.Interval.Duration > 0 {
		randomTime := time.Duration(c.Config.Rand().Intn(randomMilliseconds)) * time.Millisecond
		dur = cfg.Interval.Duration + randomTime

		for idx := range c.Plex {
			mnd.Log.Printf("==> Plex Sessions Collection Started (%d): %s, interval:%s timeout:%s webhook_cooldown:%v delay:%v",
				idx+1, c.Plex[idx].Server.URL, cfg.Interval, c.Plex[idx].Timeout, cfg.Cooldown, cfg.Delay)
		}
	}

	c.Add(&common.Action{
		Key:  "TrigPlexSessions",
		Name: TrigPlexSessions,
		Fn:   c.sendPlexSessions,
		C:    make(chan *common.ActionInput, 1),
		D:    cnfg.Duration{Duration: dur},
	})

	if cfg.MoviesPC != 0 || cfg.SeriesPC != 0 || cfg.TrackSess {
		for idx := range c.Plex {
			mnd.Log.Printf("==> Plex Sessions Tracker Started (%d): %s, interval:1m timeout:%s movies:%d%% series:%d%% play:%v",
				idx+1, c.Plex[idx].Server.URL, c.Plex[idx].Timeout, cfg.MoviesPC, cfg.SeriesPC, cfg.TrackSess)
		}

		c.Add(&common.Action{
			Key:  "TrigPlexSessionsCheck",
			Name: TrigPlexSessionsCheck,
			Hide: true, // do not log this one.
			Fn:   c.checkForFinishedItems,
			D: cnfg.Duration{Duration: time.Minute +
				time.Duration(c.Config.Rand().Intn(randomMilliseconds2))*time.Millisecond},
		})
	}
}

// SendWebhook is called in a go routine after a plex media.play webhook is received.
func (a *Action) SendWebhook(hook *plex.IncomingWebhook) {
	go a.cmd.sendWebhook(hook)
}

func (c *cmd) sendWebhook(hook *plex.IncomingWebhook) {
	// Find the matching Plex server by name, or use the first one.
	plexServer := c.findPlexServer(hook.Server.Title)
	sessions := &plex.Sessions{Name: plexServer.Name()}
	ci := clientinfo.Get()
	ctx := context.Background()

	// If NoActivity=false, then grab sessions, but wait 'Delay' to make sure they're updated.
	if ci != nil && !ci.Actions.Plex.NoActivity {
		time.Sleep(ci.Actions.Plex.Delay.Duration)
		ctx, cancel := context.WithTimeout(ctx, plexServer.Timeout.Duration)

		var err error
		if sessions, err = c.getSessionsForServer(ctx, plexServer, time.Second); err != nil {
			mnd.Log.Errorf("Getting Plex sessions: %v", err)
		}

		cancel()
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) //nolint:mnd // wait max 5 seconds for system info.
	defer cancel()

	website.SendData(&website.Request{
		Route:      website.PlexRoute,
		Event:      website.EventHook,
		Payload:    &website.Payload{Snap: c.getMetaSnap(ctx), Load: hook, Plex: sessions},
		LogMsg:     "Plex Webhook (and sessions)",
		LogPayload: true,
	})
}

// findPlexServer finds a Plex server by name, or returns the first one if no match.
func (c *cmd) findPlexServer(serverTitle string) *apps.Plex {
	if len(c.Plex) == 0 {
		return nil
	}

	for idx := range c.Plex {
		if c.Plex[idx].Server.Name() == serverTitle {
			return &c.Plex[idx]
		}
	}

	// Return first configured server if no match found.
	return &c.Plex[0]
}

// GetSessions returns the plex sessions up to 1 minute old from all Plex servers.
func (a *Action) GetSessions(ctx context.Context) (*plex.Sessions, error) {
	return a.cmd.getAllSessions(ctx, time.Minute)
}

// GetMetaSnap grabs some basic system info: cpu, memory, username. Gets added to Plex sessions and webhook payloads.
func (a *Action) GetMetaSnap(ctx context.Context) *snapshot.Snapshot {
	return a.cmd.getMetaSnap(ctx)
}

// getMetaSnap grabs some basic system info: cpu, memory, username. Gets added to Plex sessions and webhook payloads.
func (c *cmd) getMetaSnap(ctx context.Context) *snapshot.Snapshot {
	var (
		snap = &snapshot.Snapshot{}
		wait sync.WaitGroup
	)

	wait.Add(1)

	go func() {
		defer wait.Done()

		_ = snap.GetCPUSample(ctx)
	}()

	wait.Add(1)

	go func() {
		defer wait.Done()

		_ = snap.GetMemoryUsage(ctx)
	}()

	wait.Add(1)

	go func() {
		defer wait.Done()

		_ = snap.GetLocalData(ctx)
	}()

	wait.Wait()

	return snap
}
