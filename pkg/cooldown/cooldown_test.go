package cooldown_test

import (
	"testing"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/cooldown"
	"github.com/stretchr/testify/assert"
)

func TestActive(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	cooler := cooldown.NewTimer(false, time.Minute)
	assert.False(cooler.Active("key", time.Second))
	assert.True(cooler.Active("key", time.Second))
	assert.True(cooler.Active("key", time.Second))
	time.Sleep(250 * time.Millisecond)
	assert.False(cooler.Active("key", 200*time.Millisecond))
	assert.True(cooler.Active("key", 200*time.Millisecond))
	assert.True(cooler.Running())
	cooler.StopTimer()
	assert.False(cooler.Running())
}
