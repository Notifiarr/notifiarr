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

	cooler := cooldown.NewTimer(false, 10*time.Millisecond)
	assert.False(cooler.Active("key", time.Second), "this new key shuould not be active")
	assert.True(cooler.Active("key", time.Second), "the key was just used, so it must be active within last second")
	assert.True(cooler.Active("key", 20*time.Millisecond), "this could return false on a slow system")

	time.Sleep(50 * time.Millisecond)

	assert.False(cooler.Active("key", 25*time.Millisecond),
		"it's been more than 25ms since this key was seen, so it should not be active")
	assert.True(cooler.Active("key", 25*time.Millisecond), "we just saw this kety, so it should now be active")
	assert.True(cooler.Running(), "cooler is running tho...")
	cooler.StopTimer()
	assert.False(cooler.Running(), "we just stopped it, so it should be stopped!")
}
