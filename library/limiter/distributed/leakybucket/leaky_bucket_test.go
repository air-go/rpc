package leakybucket

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/mock/tools/miniredis"
)

func TestLeakyBucket(t *testing.T) {
	c := clock.NewMock()
	rc := miniredis.NewClient()
	lb := NewLeakyBucket(1, 3000, rc, WithClock(c))

	ctx := context.Background()
	key := "leaky_bucket"

	first := time.Now()
	c.Set(first)
	ok, err := lb.Allow(ctx, key, 1)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	res, err := rc.HGet(ctx, key, "volume").Result()
	assert.Nil(t, err)
	assert.Equal(t, "3000", res)
	res, err = rc.HGet(ctx, key, "rate").Result()
	assert.Nil(t, err)
	assert.Equal(t, "1", res)
	res, err = rc.HGet(ctx, key, "count").Result()
	assert.Nil(t, err)
	assert.Equal(t, "1", res)
	res, err = rc.HGet(ctx, key, "last_time").Result()
	assert.Nil(t, err)
	assert.Equal(t, strconv.FormatInt(first.Unix(), 10), res)

	second := first.Add(time.Second * 10)
	c.Set(second)
	ok, err = lb.Allow(ctx, key, 100)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	res, err = rc.HGet(ctx, key, "volume").Result()
	assert.Nil(t, err)
	assert.Equal(t, "3000", res)
	res, err = rc.HGet(ctx, key, "rate").Result()
	assert.Nil(t, err)
	assert.Equal(t, "1", res)
	res, err = rc.HGet(ctx, key, "count").Result()
	assert.Nil(t, err)
	assert.Equal(t, "91", res)
	res, err = rc.HGet(ctx, key, "last_time").Result()
	assert.Nil(t, err)
	assert.Equal(t, strconv.FormatInt(second.Unix(), 10), res)

	third := second.Add(time.Second * 1)
	c.Set(third)
	ok, err = lb.Allow(ctx, key, 3000)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
	res, err = rc.HGet(ctx, key, "volume").Result()
	assert.Nil(t, err)
	assert.Equal(t, "3000", res)
	res, err = rc.HGet(ctx, key, "rate").Result()
	assert.Nil(t, err)
	assert.Equal(t, "1", res)
	res, err = rc.HGet(ctx, key, "count").Result()
	assert.Nil(t, err)
	assert.Equal(t, "3000", res)
	res, err = rc.HGet(ctx, key, "last_time").Result()
	assert.Nil(t, err)
	assert.Equal(t, strconv.FormatInt(third.Unix(), 10), res)

	lb.SetVolume(3001)
	ok, err = lb.Allow(ctx, key, 1)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	res, err = rc.HGet(ctx, key, "volume").Result()
	assert.Nil(t, err)
	assert.Equal(t, "3001", res)
	res, err = rc.HGet(ctx, key, "rate").Result()
	assert.Nil(t, err)
	assert.Equal(t, "1", res)
	res, err = rc.HGet(ctx, key, "count").Result()
	assert.Nil(t, err)
	assert.Equal(t, "3001", res)
	res, err = rc.HGet(ctx, key, "last_time").Result()
	assert.Nil(t, err)
	assert.Equal(t, strconv.FormatInt(third.Unix(), 10), res)

	lb.SetRate(10000)
	forth := third.Add(time.Second * 1)
	c.Set(forth)
	ok, err = lb.Allow(ctx, key, 500)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	res, err = rc.HGet(ctx, key, "volume").Result()
	assert.Nil(t, err)
	assert.Equal(t, "3001", res)
	res, err = rc.HGet(ctx, key, "rate").Result()
	assert.Nil(t, err)
	assert.Equal(t, "10000", res)
	res, err = rc.HGet(ctx, key, "count").Result()
	assert.Nil(t, err)
	assert.Equal(t, "0", res)
	res, err = rc.HGet(ctx, key, "last_time").Result()
	assert.Nil(t, err)
	assert.Equal(t, strconv.FormatInt(forth.Unix(), 10), res)
}
