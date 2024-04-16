package leakybucket

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/limiter"
	"github.com/air-go/rpc/mock/tools/miniredis"
)

func TestLeakyBucket(t *testing.T) {
	c := clock.NewMock()
	rc := miniredis.NewClient()
	lb, _ := NewLeakyBucket(func() *redis.Client {
		return rc
	}, WithClock(c))

	ctx := context.Background()
	key := "leaky_bucket"

	first := time.Now()
	c.Set(first)
	ok, err := lb.Allow(ctx, key)
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
	ok, err = lb.Allow(ctx, key, limiter.OptionCount(100))
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
	assert.Equal(t, "100", res)
	res, err = rc.HGet(ctx, key, "last_time").Result()
	assert.Nil(t, err)
	assert.Equal(t, strconv.FormatInt(second.Unix(), 10), res)

	third := second.Add(time.Second * 1)
	c.Set(third)
	ok, err = lb.Allow(ctx, key, limiter.OptionCount(3000))
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

	ok, err = lb.Allow(ctx, key)
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

	forth := third.Add(time.Second * 1)
	c.Set(forth)
	lb.SetRate(ctx, key, 10000)
	lb.SetVolume(ctx, key, 3001)
	ok, err = lb.Allow(ctx, key, limiter.OptionCount(500))
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
	assert.Equal(t, "500", res)
	res, err = rc.HGet(ctx, key, "last_time").Result()
	assert.Nil(t, err)
	assert.Equal(t, strconv.FormatInt(forth.Unix(), 10), res)
}
