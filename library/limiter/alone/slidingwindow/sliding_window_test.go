package slidingwindow

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func TestAllow(t *testing.T) {
	c := clock.NewMock()
	now := time.Now()
	c.Set(now)

	sw := NewSlidingWindow(1, time.Second, WithClock(c))
	assert.Equal(t, true, sw.Allow())
	assert.Equal(t, false, sw.Allow())

	c.Set(now.Add(time.Second))
	assert.Equal(t, true, sw.Allow())
	assert.Equal(t, false, sw.Allow())
}

func TestSetLimit(t *testing.T) {
	c := clock.NewMock()
	now := time.Now()
	c.Set(now)

	sw := NewSlidingWindow(1, time.Second, WithClock(c))
	assert.Equal(t, true, sw.Allow())
	assert.Equal(t, false, sw.Allow())

	c.Set(now.Add(time.Second))
	sw.SetLimit(2)
	assert.Equal(t, true, sw.Allow())
	assert.Equal(t, true, sw.Allow())
	assert.Equal(t, false, sw.Allow())
}

func TestSetWindow(t *testing.T) {
	c := clock.NewMock()
	now := time.Now()
	c.Set(now)

	sw := NewSlidingWindow(1, time.Second, WithClock(c))
	assert.Equal(t, true, sw.Allow())
	assert.Equal(t, false, sw.Allow())

	sw.SetWindow(time.Millisecond * 500)
	c.Set(now.Add(time.Millisecond * 499))
	assert.Equal(t, false, sw.Allow())
	c.Set(now.Add(time.Millisecond * 500))
	assert.Equal(t, true, sw.Allow())
	assert.Equal(t, false, sw.Allow())
}
