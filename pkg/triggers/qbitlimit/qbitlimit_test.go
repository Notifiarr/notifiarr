package qbitlimit //nolint:testpackage

import (
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/apps"
	"github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex"
	"github.com/Notifiarr/notifiarr/pkg/triggers/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/cnfg"
	"golift.io/qbit"
)

func TestWanSession(t *testing.T) {
	t.Parallel()

	wanMovie := &plex.Session{Type: movieType, Player: plex.Player{State: playing}}
	wanMovie.Session.Location = "wan"

	wanEpisode := &plex.Session{Type: episodeType, Player: plex.Player{State: paused}}
	wanEpisode.Session.Location = "wan"

	lanMovie := &plex.Session{Type: movieType, Player: plex.Player{State: playing}}
	lanMovie.Session.Location = lanLocation

	buffering := &plex.Session{Type: movieType, Player: plex.Player{State: "buffering"}}
	buffering.Session.Location = "wan"

	track := &plex.Session{Type: "track", Player: plex.Player{State: playing}}
	track.Session.Location = "wan"

	tests := []struct {
		name    string
		session *plex.Session
		want    bool
	}{
		{name: "nil", session: nil, want: false},
		{name: "wan playing movie", session: wanMovie, want: true},
		{name: "wan paused episode", session: wanEpisode, want: true},
		{name: "lan playing movie", session: lanMovie, want: false},
		{name: "wan buffering movie", session: buffering, want: false},
		{name: "wan playing track", session: track, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, wanSession(test.session))
		})
	}
}

func TestHasWANSession(t *testing.T) {
	t.Parallel()

	assert.False(t, hasWANSession(nil))
	assert.False(t, hasWANSession(&plex.Sessions{}))

	session := &plex.Session{Type: movieType, Player: plex.Player{State: playing}}
	session.Session.Location = "wan"
	assert.True(t, hasWANSession(&plex.Sessions{Sessions: []*plex.Session{session}}))
}

func TestPlanTurtle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		current   bool
		desired   bool
		weEnabled bool
		on        bool
		newWe     bool
		skip      bool
	}{
		{
			name:      "desired already on not ours",
			current:   true,
			desired:   true,
			weEnabled: false,
			skip:      true,
			on:        true,
			newWe:     false,
		},
		{name: "desired already on we own", current: true, desired: true, weEnabled: true, skip: true, on: true, newWe: true},
		{name: "desired off so we enable", current: false, desired: true, weEnabled: false, on: true, newWe: true},
		{name: "restore what we enabled", current: true, desired: false, weEnabled: true, on: false, newWe: false},
		{name: "leave users turtle", current: true, desired: false, weEnabled: false, skip: true, on: true, newWe: false},
		{name: "already off", current: false, desired: false, weEnabled: false, skip: true, on: false, newWe: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			on, newWe, skip := planTurtle(test.current, test.desired, test.weEnabled)
			assert.Equal(t, test.on, on)
			assert.Equal(t, test.newWe, newWe)
			assert.Equal(t, test.skip, skip)
		})
	}
}

func TestInstanceDesired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		desired  bool
		selected bool
		owned    bool
		want     bool
		skip     bool
	}{
		{name: "selected want on", desired: true, selected: true, owned: false, want: true},
		{name: "selected want off", desired: false, selected: true, owned: true, want: false},
		{name: "unselected not ours", desired: true, selected: false, owned: false, skip: true},
		{name: "unselected we own so restore", desired: true, selected: false, owned: true, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			want, skip := instanceDesired(test.desired, test.selected, test.owned)
			assert.Equal(t, test.want, want)
			assert.Equal(t, test.skip, skip)
		})
	}
}

func TestCooldown(t *testing.T) {
	t.Parallel()

	assert.Equal(t, defaultCool, cooldown(cnfg.Duration{}))
	assert.Equal(t, time.Minute, cooldown(cnfg.Duration{Duration: time.Minute}))
}

func TestIdleState(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cool := 10 * time.Second

	since, restore := idleState(true, now.Add(-time.Hour), now, cool)
	assert.True(t, since.IsZero())
	assert.False(t, restore)

	since, restore = idleState(false, time.Time{}, now, cool)
	assert.Equal(t, now, since)
	assert.False(t, restore)

	since, restore = idleState(false, now, now.Add(9*time.Second), cool)
	assert.Equal(t, now, since)
	assert.False(t, restore)

	since, restore = idleState(false, now, now.Add(cool), cool)
	assert.Equal(t, now, since)
	assert.True(t, restore)
}

func TestEnqueueCoalesces(t *testing.T) {
	t.Parallel()

	cmd := &cmd{accepting: true, kick: make(chan *common.ActionInput, 1)}
	require.True(t, cmd.enqueue(&common.ActionInput{ReqID: "poll"}))
	require.True(t, cmd.enqueue(&common.ActionInput{ReqID: "poll-2"}))
	require.True(t, cmd.enqueue(&common.ActionInput{ReqID: "mode", Args: []string{"disable"}}))

	got := <-cmd.kick
	assert.Equal(t, []string{"disable"}, got.Args)

	cmd.accepting = false
	assert.False(t, cmd.enqueue(&common.ActionInput{Args: []string{"enable"}}))
}

func TestSeedOwnedURL(t *testing.T) { //nolint:paralleltest // mutates the package ownership map
	ownedMu.Lock()
	ownedByURL = map[string]bool{}
	remoteDeadline = time.Time{}
	ownedMu.Unlock()

	t.Cleanup(func() {
		ownedMu.Lock()
		ownedByURL = map[string]bool{}
		remoteDeadline = time.Time{}
		ownedMu.Unlock()
	})

	setOwnedURL("http://qbit.example", true)
	setOwnedURL("", true)

	// Keyed so ExtraConfig stays zero; modernize wants an unkeyed embed that does not compile.
	app := apps.Qbit{QbitConfig: apps.QbitConfig{Config: qbit.Config{URL: "http://qbit.example"}}} //nolint:modernize
	cmd := &cmd{
		weEnabled: map[int]bool{},
		Config:    &common.Config{Apps: &apps.Apps{Qbit: []apps.Qbit{app}}},
	}
	cmd.seedOwned()
	assert.True(t, cmd.weEnabled[1])

	setOwnedURL("http://qbit.example", false)
	cmd.weEnabled = map[int]bool{}
	cmd.seedOwned()
	assert.False(t, cmd.weEnabled[1])

	deadline := time.Now().Add(time.Hour)
	setRemoteDeadline(deadline)
	cmd.remoteUntil = time.Time{}
	cmd.seedOwned()
	assert.Equal(t, deadline, cmd.remoteUntil)

	setRemoteDeadline(time.Now().Add(-time.Minute))
	cmd.remoteUntil = time.Time{}
	cmd.seedOwned()
	assert.True(t, cmd.remoteUntil.IsZero())
}

func TestFreshSessions(t *testing.T) {
	t.Parallel()

	now := time.Now()
	session := &plex.Session{Type: movieType, Player: plex.Player{State: playing}}
	session.Session.Location = "wan"
	list := &plex.Sessions{Sessions: []*plex.Session{session}}

	assert.True(t, hasWANSession(freshSessions(now, list, now)))
	assert.Nil(t, freshSessions(now.Add(-sessionCacheAge), list, now))
	assert.Nil(t, freshSessions(time.Time{}, list, now))
	assert.Nil(t, freshSessions(now, "nope", now))
	assert.Nil(t, freshSessions(now, nil, now))
}

func TestDropMissing(t *testing.T) {
	t.Parallel()

	owned := map[int]bool{1: true, 2: true, 3: false}
	dropMissing(owned, []bool{true})
	assert.True(t, owned[1])
	assert.False(t, owned[2])
	assert.False(t, owned[3])

	dropMissing(owned, []bool{false})
	assert.False(t, owned[1])
}
