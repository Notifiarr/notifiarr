package qbitlimit

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/Notifiarr/notifiarr/pkg/triggers/data"
	"github.com/Notifiarr/notifiarr/pkg/triggers/plexcron"
	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/Notifiarr/notifiarr/pkg/website/clientinfo"
	"golift.io/cnfg"
)

const TrigQbitSpeed common.TriggerName = "Reconciling qBittorrent alternative speed limits."

const (
	pollInterval = 15 * time.Second
	jellyfinTTL  = 4 * time.Hour
	defaultCool  = 30 * time.Second
	playing      = "playing"
	paused       = "paused"
	movieType    = "movie"
	episodeType  = "episode"
	lanLocation  = "lan"
)

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	plex          *plexcron.Action
	mu            sync.Mutex
	weEnabled     map[int]bool
	jellyfinUntil time.Time
	lastDesired   time.Time
}

// New configures the library.
func New(config *common.Config, plexCron *plexcron.Action) *Action {
	return &Action{cmd: &cmd{
		Config:    config,
		plex:      plexCron,
		weEnabled: make(map[int]bool),
	}}
}

// Create initializes the library.
func (a *Action) Create() {
	reqID := mnd.ReqID()
	a.cmd.create(reqID)
}

func (c *cmd) create(reqID string) {
	var dur time.Duration

	info := clientinfo.Get()
	if info != nil && info.Actions.QbitThrottle.Enabled && c.hasQbit() {
		dur = pollInterval
		mnd.Log.Printf(reqID, "==> qBittorrent Speed Limit Timer Enabled, interval:%s cooldown:%s plex:%v jellyfin:%v",
			dur, cooldown(info.Actions.QbitThrottle.Cooldown),
			info.Actions.QbitThrottle.Plex, info.Actions.QbitThrottle.Jellyfin)
	}

	c.Add(&common.Action{
		Key:  "TrigQbitSpeed",
		Name: TrigQbitSpeed,
		Fn:   c.reconcile,
		C:    make(chan *common.ActionInput, 1),
		D:    cnfg.Duration{Duration: dur},
	})
}

func (c *cmd) hasQbit() bool {
	for idx := range c.Apps.Qbit {
		if c.Apps.Qbit[idx].Enabled() {
			return true
		}
	}

	return false
}

// Send queues a reconcile. Optional args: "enable" or "disable" (Jellyfin).
func (a *Action) Send(input *common.ActionInput) bool {
	return a.cmd.Exec(input, TrigQbitSpeed)
}

// Kick queues a non-blocking reconcile so Plex session updates cannot deadlock.
func (a *Action) Kick() {
	if a == nil || a.cmd == nil {
		return
	}

	trig := a.cmd.Get(TrigQbitSpeed)
	if trig == nil || trig.C == nil {
		return
	}

	select {
	case trig.C <- &common.ActionInput{Type: website.EventHook, ReqID: mnd.ReqID()}:
	default:
	}
}

func (c *cmd) reconcile(ctx context.Context, input *common.ActionInput) {
	c.mu.Lock()
	defer c.mu.Unlock()

	info := clientinfo.Get()
	if info == nil || !info.Actions.QbitThrottle.Enabled || !c.hasQbit() {
		return
	}

	cfg := info.Actions.QbitThrottle
	now := time.Now()
	c.applyJellyfinArg(cfg, input.Args, now)

	desired := c.desired(ctx, cfg, now)
	if desired {
		c.lastDesired = now
		c.apply(ctx, input, cfg, true)

		return
	}

	if c.lastDesired.IsZero() || now.Sub(c.lastDesired) < cooldown(cfg.Cooldown) {
		return
	}

	c.apply(ctx, input, cfg, false)
}

func (c *cmd) applyJellyfinArg(cfg clientinfo.QbitThrottleConfig, args []string, now time.Time) {
	if len(args) == 0 || !cfg.Jellyfin {
		return
	}

	switch args[0] {
	case "enable":
		c.jellyfinUntil = now.Add(jellyfinTTL)
	case "disable":
		c.jellyfinUntil = time.Time{}
	}
}

func (c *cmd) desired(ctx context.Context, cfg clientinfo.QbitThrottleConfig, now time.Time) bool {
	if cfg.Jellyfin && !c.jellyfinUntil.IsZero() && now.Before(c.jellyfinUntil) {
		return true
	}

	if !cfg.Plex || c.plex == nil || !c.Apps.Plex.Enabled() {
		return false
	}

	return hasWANSession(c.plexSessions(ctx, mnd.GetID(ctx)))
}

func (c *cmd) plexSessions(ctx context.Context, reqID string) *plex.Sessions {
	sessions, err := c.plex.GetSessions(ctx)
	if err == nil {
		return sessions
	}

	mnd.Log.Errorf(reqID, "Getting Plex sessions for qBittorrent speed limit: %v", err)

	if item := data.Get("plexCurrentSessions"); item != nil && item.Data != nil {
		cached, _ := item.Data.(*plex.Sessions)

		return cached
	}

	return sessions
}

func (c *cmd) apply(ctx context.Context, input *common.ActionInput, cfg clientinfo.QbitThrottleConfig, desired bool) {
	for idx := range c.Apps.Qbit {
		instance := idx + 1
		app := &c.Apps.Qbit[idx]

		if !app.Enabled() || (len(cfg.Instances) > 0 && !cfg.Instances.Has(instance)) {
			continue
		}

		current, err := app.SpeedLimitsModeContext(ctx)
		if err != nil {
			mnd.Log.Errorf(input.ReqID, "[%s requested] qBittorrent speed limits mode (%d:%s) failed: %v",
				input.Type, instance, app.URL, err)
			continue
		}

		turtle, weOwn, skip := planTurtle(current, desired, c.weEnabled[instance])
		c.weEnabled[instance] = weOwn

		if skip {
			continue
		}

		if err := app.SetSpeedLimitsModeContext(ctx, turtle); err != nil {
			mnd.Log.Errorf(input.ReqID, "[%s requested] Setting qBittorrent alternative speed limits (%d:%s) to %v failed: %v",
				input.Type, instance, app.URL, turtle, err)
			continue
		}

		mnd.Log.Printf(input.ReqID, "[%s requested] qBittorrent alternative speed limits (%d:%s) => %v",
			input.Type, instance, app.URL, turtle)
	}
}

func cooldown(dur cnfg.Duration) time.Duration {
	if dur.Duration <= 0 {
		return defaultCool
	}

	return dur.Duration
}

func hasWANSession(sessions *plex.Sessions) bool {
	if sessions == nil {
		return false
	}

	return slices.ContainsFunc(sessions.Sessions, wanSession)
}

func wanSession(session *plex.Session) bool {
	if session == nil {
		return false
	}

	if session.Type != movieType && session.Type != episodeType {
		return false
	}

	if session.Session.Location == lanLocation {
		return false
	}

	return session.Player.State == playing || session.Player.State == paused
}

// planTurtle decides whether to change turtle mode and whether we own the enable.
// If turtle was already on when we wanted it, we do not steal it on restore.
func planTurtle(current, desired, weEnabled bool) (bool, bool, bool) {
	if desired {
		if current {
			return current, weEnabled, true
		}

		return true, true, false
	}

	if !weEnabled {
		return current, false, true
	}

	return false, false, false
}
