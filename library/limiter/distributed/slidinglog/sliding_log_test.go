package slidinglog

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/mock/tools/miniredis"
)

func TestSlidingLog(t *testing.T) {
	ctx := context.Background()

	// test not change
	func() {
		key := "sliding_log_1"
		c := clock.NewMock()
		rc := miniredis.NewClient()
		sl, _ := NewSlidingLog(func() *redis.Client {
			return rc
		}, WithClock(c),
			WithKeyTTL(time.Minute),
			WithLimit(1),
			WithWindow(time.Second*3),
		)

		c.Set(time.Now())
		ok, err := sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)

		// not over 3 second, should limit
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, false, ok)

		// over 3 second, should not limit
		c.Add(time.Second * 4)
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)

		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, false, ok)
	}()

	// test change limit
	func() {
		key := "sliding_log_2"
		c := clock.NewMock()
		rc := miniredis.NewClient()
		sl, _ := NewSlidingLog(func() *redis.Client {
			return rc
		}, WithClock(c),
			WithKeyTTL(time.Minute),
			WithLimit(1),
			WithWindow(time.Second*3),
		)

		c.Set(time.Now())
		ok, err := sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)

		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, false, ok)

		c.Add(time.Second)
		sl.SetLimit(ctx, key, 2)
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)
	}()

	// test change window
	func() {
		key := "sliding_log_3"
		c := clock.NewMock()
		rc := miniredis.NewClient()
		sl, _ := NewSlidingLog(func() *redis.Client {
			return rc
		}, WithClock(c),
			WithKeyTTL(time.Minute),
			WithLimit(1),
			WithWindow(time.Second*3),
		)

		c.Set(time.Now())
		ok, err := sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)
		count, _ := rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)

		c.Add(time.Second * 2)
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, false, ok)
		count, _ = rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)

		sl.SetWindow(ctx, key, time.Second)
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)
		count, _ = rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)
	}()
}
