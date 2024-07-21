package plexcron

import (
	"context"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
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
	Plex *apps.PlexConfig
	sent map[string]struct{} // Tracks Finished sessions already sent.
	sync.Mutex
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
func New(config *common.Config, plex *apps.PlexConfig) *Action {
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
	a.cmd.Exec(&common.ActionInput{Type: event}, TrigPlexSessions)
}

// Run initializes the library.
func (a *Action) Create() {
	a.cmd.run()
}

func (c *cmd) run() {
	info := clientinfo.Get()
	if !c.Plex.Enabled() || info == nil {
		return
	}

	var dur time.Duration

	cfg := info.Actions.Plex
	if cfg.Interval.Duration > 0 {
		randomTime := time.Duration(c.Config.Rand().Intn(randomMilliseconds)) * time.Millisecond
		dur = cfg.Interval.Duration + randomTime
		c.Printf("==> Plex Sessions Collection Started, URL: %s, interval:%s timeout:%s webhook_cooldown:%v delay:%v",
			c.Plex.URL, cfg.Interval, c.Plex.Timeout, cfg.Cooldown, cfg.Delay)
	}

	c.Add(&common.Action{
		Name: TrigPlexSessions,
		Fn:   c.sendPlexSessions,
		C:    make(chan *common.ActionInput, 1),
		D:    cnfg.Duration{Duration: dur},
	})

	if cfg.MoviesPC != 0 || cfg.SeriesPC != 0 || cfg.TrackSess {
		c.Printf("==> Plex Sessions Tracker Started, URL: %s, interval:1m timeout:%s movies:%d%% series:%d%% play:%v",
			c.Plex.URL, c.Plex.Timeout, cfg.MoviesPC, cfg.SeriesPC, cfg.TrackSess)

		c.Add(&common.Action{
			Name: "Checking Plex for completed sessions.",
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
	sessions := &plex.Sessions{Name: c.Plex.Server.Name()}
	ci := clientinfo.Get()
	ctx := context.Background()

	// If NoActivity=false, then grab sessions, but wait 'Delay' to make sure they're updated.
	if ci != nil && !ci.Actions.Plex.NoActivity {
		time.Sleep(ci.Actions.Plex.Delay.Duration)
		ctx, cancel := context.WithTimeout(ctx, c.Plex.Timeout.Duration)

		var err error
		if sessions, err = c.getSessions(ctx, time.Second); err != nil {
			c.Errorf("Getting Plex sessions: %v", err)
		}

		cancel()
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) //nolint:mnd // wait max 5 seconds for system info.
	defer cancel()

	c.SendData(&website.Request{
		Route:      website.PlexRoute,
		Event:      website.EventHook,
		Payload:    &website.Payload{Snap: c.getMetaSnap(ctx), Load: hook, Plex: sessions},
		LogMsg:     "Plex Webhook (and sessions)",
		LogPayload: true,
	})
}

// GetSessions returns the plex sessions up to 1 minute old. This uses a channel so concurrent requests are avoided.
func (a *Action) GetSessions(ctx context.Context) (*plex.Sessions, error) {
	return a.cmd.getSessions(ctx, time.Minute)
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
