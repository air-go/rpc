package tokenbucket

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func TestTokenBucket(t *testing.T) {
	ctx := context.Background()

	c := clock.NewMock()
	now := time.Now()
	c.Set(now)

	l := NewTokenBucket(
		WithClock(c),
		WithLimit(1),
		WithBurst(2),
	)

	key := "test"
	ok, err := l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	now = now.Add(time.Second)
	c.Set(now)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	l.SetLimit(ctx, key, 2)
	now = now.Add(time.Second)
	c.Set(now)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	l.SetBurst(ctx, key, 3)
	now = now.Add(time.Second * 3)
	c.Set(now)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	ok, err = l.Allow(ctx, key)
	assert.Equal(t, true, ok)
	ok, err = l.Allow(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
}
