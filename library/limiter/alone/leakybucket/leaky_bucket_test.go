package leakybucket

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func TestLeaky(t *testing.T) {
	ctx := context.Background()
	key := "test"
	c := clock.NewMock()
	lb := NewLeakyBucket(
		WithClock(c),
		WithRate(time.Second, 1),
		WithVolume(2),
	)

	// test volume, can process 2
	first := time.Now()
	c.Set(first)
	ok, err := lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	// test add 1 perRequest, 1 can be processed
	second := first.Add(time.Second)
	c.Set(second)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	// test add 2 second perRequest, 2 can be processed
	third := second.Add(time.Second * 2)
	c.Set(third)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	// test add 3 second perRequest, but exceed 2 volume, only 2 can be processed
	forth := third.Add(time.Second * 3)
	c.Set(forth)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
}

func TestSetRate(t *testing.T) {
	ctx := context.Background()
	key := "test"
	c := clock.NewMock()
	lb := NewLeakyBucket(
		WithClock(c),
		WithRate(time.Second, 1),
		WithVolume(2),
	)

	// test volume, can process 2
	first := time.Now()
	c.Set(first)
	ok, err := lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	// test add 1 rate, dynamic change rate to 2, 2 can be processed
	lb.SetRate(ctx, key, time.Second, 2)
	second := first.Add(time.Second)
	c.Set(second)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = lb.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
}
