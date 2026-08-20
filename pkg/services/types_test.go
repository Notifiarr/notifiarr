package services_test

import (
	"encoding/json"
	"testing"

	"github.com/Notifiarr/notifiarr/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckStateStringAndValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state services.CheckState
		name  string
		value uint
	}{
		{services.StateOK, "OK", 0},
		{services.StateWarning, "Warning", 1},
		{services.StateCritical, "Critical", 2},
		{services.StateUnknown, "Unknown", 3},
		{services.CheckState(99), "Unknown", 99},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.name, test.state.String())
			assert.Equal(t, test.value, test.state.Value())
		})
	}
}

func TestOutputStringAndJSON(t *testing.T) {
	t.Parallel()

	var nilOutput *services.Output

	assert.Empty(t, nilOutput.String())

	plain := &services.Output{}
	require.NoError(t, json.Unmarshal([]byte(`"hello & world"`), plain))
	assert.Equal(t, "hello & world", plain.String())

	raw, err := json.Marshal(plain)
	require.NoError(t, err)
	assert.JSONEq(t, `"hello & world"`, string(raw))

	escaped := &services.Output{}
	require.NoError(t, json.Unmarshal([]byte(`"Tom &amp; Jerry"`), escaped))
	// Unmarshal stores the JSON string as-is; String() unescapes only when esc is set,
	// which CheckOnly sets for HTTP error bodies. Marshal must not unescape.
	raw, err = json.Marshal(escaped)
	require.NoError(t, err)
	assert.JSONEq(t, `"Tom &amp; Jerry"`, string(raw))
}

func TestPackageDefaults(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "http", string(services.CheckHTTP))
	assert.Equal(t, "tcp", string(services.CheckTCP))
	assert.Equal(t, "ping", string(services.CheckPING))
	assert.Equal(t, "icmp", string(services.CheckICMP))
	assert.Equal(t, "process", string(services.CheckPROC))
	assert.Equal(t, "Plex Server", services.PlexServerName)
	assert.Equal(t, 10, int(services.MinimumCheckInterval.Seconds()))
	assert.Equal(t, 1, int(services.MinimumTimeout.Seconds()))
	assert.Equal(t, 10, int(services.DefaultTimeout.Seconds()))
}
