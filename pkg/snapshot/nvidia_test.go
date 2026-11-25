package snapshot_test

import (
	"testing"

	"github.com/Notifiarr/notifiarr/pkg/snapshot"
	"github.com/stretchr/testify/assert"
)

func TestNvidiaConfigHasID(t *testing.T) {
	t.Parallel()

	var unset *snapshot.NvidiaConfig
	assert.True(t, unset.HasID("00000000:01:00.0"), "nil config should not filter bus IDs")

	empty := &snapshot.NvidiaConfig{}
	assert.True(t, empty.HasID("00000000:01:00.0"), "missing busIDs should not filter")

	blanks := &snapshot.NvidiaConfig{BusIDs: []string{"", ""}}
	assert.True(t, blanks.HasID("00000000:01:00.0"), "empty bus ID entries should not filter")

	filtered := &snapshot.NvidiaConfig{BusIDs: []string{"00000000:01:00.0"}}
	assert.True(t, filtered.HasID("00000000:01:00.0"))
	assert.False(t, filtered.HasID("00000000:04:00.0"))
}
