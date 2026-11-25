package common //nolint:testpackage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTryExecDuringClose(t *testing.T) {
	t.Parallel()

	const name TriggerName = "Reconciling qBittorrent alternative speed limits."

	cfg := &Config{stop: &Action{C: make(chan *ActionInput)}}
	ch := make(chan *ActionInput, 1)
	cfg.Add(&Action{Name: name, C: ch})

	var wait sync.WaitGroup

	wait.Go(func() {
		for range 50 {
			_ = cfg.TryExec(&ActionInput{ReqID: "reload"}, name)
		}
	})

	cfg.closeTriggers(cfg.list)
	wait.Wait()

	require.NotPanics(t, func() {
		assert.False(t, cfg.TryExec(&ActionInput{ReqID: "after"}, name))
	})
}
