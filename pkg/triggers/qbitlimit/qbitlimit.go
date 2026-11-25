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
	pollInterval     = 15 * time.Second
	remoteTTL        = 4 * time.Hour
	defaultCool      = 30 * time.Second
	reconcileTimeout = time.Minute
	sessionCacheAge  = time.Minute // same window as plexcron.GetSessions
	playing          = "playing"
	paused           = "paused"
	movieType        = "movie"
	episodeType      = "episode"
	lanLocation      = "lan"
)

// SendStatus is the result of queueing a speed-limit reconcile.
type SendStatus uint8

const (
	// SendDisabled means the feature or qBittorrent is off.
	SendDisabled SendStatus = iota
	// SendQueued means a worker will reconcile this request.
	SendQueued
	// SendStopping means the worker is shut down and did not accept the request.
	SendStopping
)

// Action contains the exported methods for this package.
type Action struct {
	cmd *cmd
}

type cmd struct {
	*common.Config
	plex        *plexcron.Action
	mu          sync.Mutex
	kickMu      sync.Mutex
	weEnabled   map[int]bool
	leftAlone   map[int]bool
	remoteUntil time.Time
	idleSince   time.Time
	kick        chan *common.ActionInput
	cancel      context.CancelFunc
	done        chan struct{}
	accepting   bool
}

// ownedByURL survives a full configuration reload, which builds a new Action and
// would otherwise treat turtles this process enabled as user-owned.
var (
	ownedMu    sync.Mutex
	ownedByURL = map[string]bool{}
)

// New configures the library.
func New(config *common.Config, plexCron *plexcron.Action) *Action {
	return &Action{cmd: &cmd{
		Config:    config,
		plex:      plexCron,
		weEnabled: make(map[int]bool),
		leftAlone: make(map[int]bool),
		kick:      make(chan *common.ActionInput, 1),
		accepting: true,
	}}
}

// Run starts the reconcile worker. The shared timer loop only kicks this worker.
func (a *Action) Run(ctx context.Context) {
	if a == nil || a.cmd == nil {
		return
	}

	a.cmd.startWorker(ctx)
}

// Stop cancels the reconcile worker and waits for it to leave qBittorrent alone.
func (a *Action) Stop() {
	if a == nil || a.cmd == nil {
		return
	}

	a.cmd.stopWorker()
}

// Create initializes the library.
func (a *Action) Create() {
	reqID := mnd.ReqID()
	a.cmd.create(reqID)
}

func (c *cmd) create(reqID string) {
	c.kickMu.Lock()
	c.accepting = true
	c.kickMu.Unlock()

	var dur time.Duration

	info := clientinfo.Get()
	enabled := info != nil && info.Actions.QbitThrottle.Enabled && c.hasQbit()

	c.mu.Lock()
	c.seedOwned()
	c.dropGone()
	owned := c.hasOwned()
	c.mu.Unlock()

	switch {
	case enabled:
		dur = pollInterval
		mnd.Log.Printf(reqID, "==> qBittorrent Speed Limit Timer Enabled, interval:%s cooldown:%s plex:%v",
			dur, cooldown(info.Actions.QbitThrottle.Cooldown), info.Actions.QbitThrottle.Plex)
	case owned && c.hasQbit():
		dur = pollInterval
		mnd.Log.Printf(reqID, "==> qBittorrent Speed Limit Timer restoring owned turtle mode, interval:%s", dur)
	}

	c.Add(&common.Action{
		Key:  "TrigQbitSpeed",
		Name: TrigQbitSpeed,
		Hide: true,   // 15s poll would spam Event Triggered.
		Fn:   c.poke, // must return immediately; reconcile runs on the worker.
		C:    make(chan *common.ActionInput, 1),
		D:    cnfg.Duration{Duration: dur},
	})

	if !enabled && owned {
		_ = c.enqueue(&common.ActionInput{Type: website.EventUser, ReqID: reqID})
	}
}

func (c *cmd) hasQbit() bool {
	for idx := range c.Apps.Qbit {
		if c.Apps.Qbit[idx].Enabled() {
			return true
		}
	}

	return false
}

func (c *cmd) hasOwned() bool {
	for _, owned := range c.weEnabled {
		if owned {
			return true
		}
	}

	return false
}

// dropGone forgets ownership of qBit instances that were removed or disabled.
func (c *cmd) dropGone() {
	if c.Apps == nil {
		return
	}

	live := make([]bool, len(c.Apps.Qbit))
	for idx := range c.Apps.Qbit {
		live[idx] = c.Apps.Qbit[idx].Enabled()
	}

	dropMissing(c.weEnabled, live)
}

// dropMissing removes ownership of instances that are gone or disabled.
func dropMissing(weEnabled map[int]bool, enabled []bool) {
	for instance, owned := range weEnabled {
		if !owned {
			continue
		}

		idx := instance - 1
		if idx < 0 || idx >= len(enabled) || !enabled[idx] {
			delete(weEnabled, instance)
		}
	}
}

func (c *cmd) ready() bool {
	info := clientinfo.Get()

	return info != nil && info.Actions.QbitThrottle.Enabled && c.hasQbit()
}

// Send queues a reconcile. Optional args: "enable" or "disable" (website remote playback).
func (a *Action) Send(input *common.ActionInput) SendStatus {
	if a == nil || a.cmd == nil || !a.cmd.ready() {
		return SendDisabled
	}

	if !a.cmd.enqueue(input) {
		return SendStopping
	}

	return SendQueued
}

// Kick queues a non-blocking reconcile so Plex session updates cannot deadlock.
// Must not take c.mu: GetSessions may call Kick while reconcile already holds it.
func (a *Action) Kick() {
	if a == nil || a.cmd == nil || !a.cmd.ready() {
		return
	}

	_ = a.cmd.enqueue(&common.ActionInput{Type: website.EventHook, ReqID: mnd.ReqID()})
}

// poke is the timer-loop callback. It must not call qBittorrent; one stalled server
// would block every other trigger and shutdown.
func (c *cmd) poke(_ context.Context, input *common.ActionInput) {
	_ = c.enqueue(input)
}

func (c *cmd) startWorker(parent context.Context) {
	c.kickMu.Lock()
	defer c.kickMu.Unlock()

	if c.cancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(parent)
	c.cancel = cancel
	c.accepting = true
	c.done = make(chan struct{})

	go c.loop(ctx, c.done)
}

func (c *cmd) stopWorker() {
	c.kickMu.Lock()
	cancel := c.cancel
	done := c.done
	c.cancel = nil
	c.accepting = false
	c.kickMu.Unlock()

	if cancel != nil {
		cancel()
	}

	if done != nil {
		<-done
	}
}

func (c *cmd) loop(ctx context.Context, done chan struct{}) {
	defer close(done)

	for {
		select {
		case <-ctx.Done():
			return
		case input := <-c.kick:
			if input == nil {
				continue
			}

			c.reconcile(ctx, input)
		}
	}
}

// enqueue coalesces empty polls and keeps the latest enable/disable.
// It returns false when the worker is stopping, and it does not send on a closed channel.
func (c *cmd) enqueue(input *common.ActionInput) bool {
	if input == nil {
		return false
	}

	c.kickMu.Lock()
	defer c.kickMu.Unlock()

	if !c.accepting {
		return false
	}

	select {
	case c.kick <- input:
		return true
	default:
	}

	if len(input.Args) == 0 {
		return true
	}

	select {
	case <-c.kick:
	default:
	}

	select {
	case c.kick <- input:
		return true
	default:
		return false
	}
}

func (c *cmd) reconcile(ctx context.Context, input *common.ActionInput) {
	c.mu.Lock()
	defer c.mu.Unlock()

	info := clientinfo.Get()
	if info == nil {
		return
	}

	c.dropGone()

	if !c.hasQbit() {
		return
	}

	cfg := info.Actions.QbitThrottle
	if !cfg.Enabled {
		if c.hasOwned() {
			c.apply(ctx, input, cfg, false)
		}

		return
	}

	now := time.Now()
	c.applyRemoteArg(input, now)

	desired := c.desired(ctx, cfg, now)
	since, restore := idleState(desired, c.idleSince, now, cooldown(cfg.Cooldown))
	c.idleSince = since

	if desired {
		c.apply(ctx, input, cfg, true)

		return
	}

	// Cooldown starts when playback goes idle, not from the last active poll.
	if !restore || !c.hasOwned() {
		return
	}

	c.apply(ctx, input, cfg, false)
}

func (c *cmd) applyRemoteArg(input *common.ActionInput, now time.Time) {
	if len(input.Args) == 0 {
		return
	}

	switch input.Args[0] {
	case "enable":
		c.remoteUntil = now.Add(remoteTTL)
		mnd.Log.Printf(input.ReqID, "[%s requested] Remote playback active; qBittorrent alternative speed limits wanted",
			input.Type)
	case "disable":
		c.remoteUntil = time.Time{}
		mnd.Log.Printf(input.ReqID,
			"[%s requested] Remote playback ended; qBittorrent alternative speed limits will restore after cooldown",
			input.Type)
	}
}

func (c *cmd) desired(ctx context.Context, cfg clientinfo.QbitThrottleConfig, now time.Time) bool {
	if !c.remoteUntil.IsZero() && now.Before(c.remoteUntil) {
		return true
	}

	if !cfg.Plex || c.plex == nil || !c.Apps.Plex.Enabled() {
		return false
	}

	return hasWANSession(c.plexSessions(ctx, mnd.GetID(ctx)))
}

func (c *cmd) plexSessions(ctx context.Context, reqID string) *plex.Sessions {
	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	sessions, err := c.plex.GetSessions(ctx)
	if err == nil {
		return sessions
	}

	mnd.Log.Errorf(reqID, "Getting Plex sessions for qBittorrent speed limit: %v", err)

	// item.Time is the save time. Get refreshes last access, which would keep a dead
	// WAN session looking fresh for as long as Plex stays down.
	if item := data.Get("plexCurrentSessions"); item != nil {
		if fresh := freshSessions(item.Time, item.Data, time.Now()); fresh != nil {
			return fresh
		}
	}

	return sessions
}

// freshSessions returns a cached Plex session list only while it is still inside sessionCacheAge.
func freshSessions(saved time.Time, item any, now time.Time) *plex.Sessions {
	if item == nil || saved.IsZero() || now.Sub(saved) >= sessionCacheAge {
		return nil
	}

	sessions, _ := item.(*plex.Sessions)

	return sessions
}

func (c *cmd) apply(ctx context.Context, input *common.ActionInput, cfg clientinfo.QbitThrottleConfig, desired bool) {
	for idx := range c.Apps.Qbit {
		c.syncInstance(ctx, input, cfg, desired, idx)
	}
}

// syncInstance gives this qBit its own deadline so a stalled server cannot consume the
// budget of every instance after it. Each poll starts at index zero.
func (c *cmd) syncInstance(
	ctx context.Context,
	input *common.ActionInput,
	cfg clientinfo.QbitThrottleConfig,
	desired bool,
	idx int,
) {
	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	instance := idx + 1
	app := &c.Apps.Qbit[idx]
	owned := c.weEnabled[instance]
	// Empty Instances means every qBit instance; the website has no picker in v1.
	selected := len(cfg.Instances) == 0 || cfg.Instances.Has(instance)
	want, skipInst := instanceDesired(desired, selected, owned)

	if !app.Enabled() || skipInst {
		return
	}

	current, err := app.SpeedLimitsModeContext(ctx)
	if err != nil {
		mnd.Log.Errorf(input.ReqID, "[%s requested] qBittorrent speed limits mode (%d:%s) failed: %v",
			input.Type, instance, app.URL, err)

		return
	}

	turtle, weOwn, skip := planTurtle(current, want, owned)
	if skip {
		if current && !owned {
			c.logUnowned(input.ReqID, instance, app.URL)
		}

		return
	}

	if err := app.SetSpeedLimitsModeContext(ctx, turtle); err != nil {
		mnd.Log.Errorf(input.ReqID, "[%s requested] Setting qBittorrent alternative speed limits (%d:%s) to %v failed: %v",
			input.Type, instance, app.URL, turtle, err)

		return
	}

	c.weEnabled[instance] = weOwn
	setOwnedURL(app.URL, weOwn)
	delete(c.leftAlone, instance)

	mnd.Log.Printf(input.ReqID, "[%s requested] qBittorrent alternative speed limits (%d:%s) => %v",
		input.Type, instance, app.URL, turtle)
}

// idleState starts the cooldown at the transition to idle.
// The restore flag is false until a full cooldown has elapsed since that transition.
func idleState(desired bool, idleSince, now time.Time, cool time.Duration) (time.Time, bool) {
	if desired {
		return time.Time{}, false
	}

	if idleSince.IsZero() {
		idleSince = now
	}

	return idleSince, now.Sub(idleSince) >= cool
}

func (c *cmd) seedOwned() {
	if c.Apps == nil {
		return
	}

	ownedMu.Lock()
	defer ownedMu.Unlock()

	for idx := range c.Apps.Qbit {
		if ownedByURL[c.Apps.Qbit[idx].URL] {
			c.weEnabled[idx+1] = true
		}
	}
}

func setOwnedURL(url string, owned bool) {
	if url == "" {
		return
	}

	ownedMu.Lock()
	defer ownedMu.Unlock()

	if owned {
		ownedByURL[url] = true

		return
	}

	delete(ownedByURL, url)
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

	// Paused WAN sessions still count: the viewer is still in the stream.
	return session.Player.State == playing || session.Player.State == paused
}

func (c *cmd) logUnowned(reqID string, instance int, url string) {
	if c.leftAlone[instance] {
		return
	}

	c.leftAlone[instance] = true
	mnd.Log.Printf(reqID,
		"qBittorrent alternative speed limits already on (%d:%s); leaving them (not enabled by this client)",
		instance, url)
}

// instanceDesired is the per-instance want/skip before planTurtle.
// Unselected instances we do not own are ignored; owned-but-unselected instances restore.
func instanceDesired(desired, selected, owned bool) (bool, bool) {
	if selected {
		return desired, false
	}

	return false, !owned
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

var (
	_ common.Create = (*Action)(nil)
	_ common.Run    = (*Action)(nil)
)
