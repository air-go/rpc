package leakybucket

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func TestLeaky(t *testing.T) {
	c := clock.NewMock()
	rl := NewLeakyBucket(1, 2, WithPer(time.Second), WithClock(c))

	// test volume, can process 2
	first := time.Now()
	c.Set(first)
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, false, rl.Allow())

	// test add 1 perRequest, 1 can be processed
	second := first.Add(time.Second)
	c.Set(second)
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, false, rl.Allow())

	// test add 2 second perRequest, 2 can be processed
	third := second.Add(time.Second * 2)
	c.Set(third)
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, false, rl.Allow())

	// test add 3 second perRequest, but exceed 2 volume, only 2 can be processed
	forth := third.Add(time.Second * 3)
	c.Set(forth)
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, false, rl.Allow())
}

func TestSetLimit(t *testing.T) {
	c := clock.NewMock()
	rl := NewLeakyBucket(1, 2, WithPer(time.Second), WithClock(c))

	// test volume, can process 2
	first := time.Now()
	c.Set(first)
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, false, rl.Allow())

	// test add 1 rate, dynamic change rate to 2, 2 can be processed
	rl.SetRate(2)
	second := first.Add(time.Second)
	c.Set(second)
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, false, rl.Allow())
}

func TestSetBurst(t *testing.T) {
	c := clock.NewMock()
	rl := NewLeakyBucket(1, 2, WithPer(time.Second), WithClock(c))

	// test volume, can process 2
	first := time.Now()
	c.Set(first)
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, false, rl.Allow())

	// test add 1 second perRequest, dynamic change burst to 3, 3 can be processed
	rl.SetRate(3)
	second := first.Add(time.Second)
	c.Set(second)
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, true, rl.Allow())
	assert.Equal(t, false, rl.Allow())
}
