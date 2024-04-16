package slidingwindow

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func TestAllow(t *testing.T) {
	ctx := context.Background()
	key := "test"
	c := clock.NewMock()
	now := time.Now()
	c.Set(now)

	sw := NewSlidingWindow(
		WithClock(c),
		WithLimit(1),
		WithWindow(time.Second),
	)

	ok, err := sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	c.Set(now.Add(time.Second))
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
}

func TestSetLimit(t *testing.T) {
	ctx := context.Background()
	key := "test"
	c := clock.NewMock()
	now := time.Now()
	c.Set(now)

	sw := NewSlidingWindow(
		WithClock(c),
		WithLimit(1),
		WithWindow(time.Second),
	)

	ok, err := sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	c.Set(now.Add(time.Second))
	sw.SetLimit(ctx, key, 2)
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
}

func TestSetWindow(t *testing.T) {
	ctx := context.Background()
	key := "test"
	c := clock.NewMock()
	now := time.Now()
	c.Set(now)

	sw := NewSlidingWindow(
		WithClock(c),
		WithLimit(1),
		WithWindow(time.Second),
	)

	ok, err := sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	sw.SetWindow(ctx, key, time.Millisecond*500)
	c.Set(now.Add(time.Millisecond * 499))
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
	c.Set(now.Add(time.Millisecond * 500))
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = sw.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
}
