package slidinglog

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
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
		sl := NewSlidingLog(1, time.Second*3, rc, WithClock(c))

		c.Set(time.Now())
		ok, err := sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)
		count, _ := rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)

		// not over 3 second, should limit
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, false, ok)
		count, _ = rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)

		// over 3 second, should not limit
		c.Add(time.Second * 4)
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)
		count, _ = rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, false, ok)
		count, _ = rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)
	}()

	// test change limit
	func() {
		key := "sliding_log_2"
		c := clock.NewMock()
		rc := miniredis.NewClient()
		sl := NewSlidingLog(1, time.Second*3, rc, WithClock(c))

		c.Set(time.Now())
		ok, err := sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)
		count, _ := rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)

		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, false, ok)
		count, _ = rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)

		c.Add(time.Second)
		sl.SetLimit(2)
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)
		count, _ = rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(2), count)
	}()

	// test change window
	func() {
		key := "sliding_log_3"
		c := clock.NewMock()
		rc := miniredis.NewClient()
		sl := NewSlidingLog(1, time.Second*3, rc, WithClock(c))

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

		sl.SetWindow(1)
		ok, err = sl.Allow(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, true, ok)
		count, _ = rc.ZCount(ctx, key, "0", strconv.FormatInt(c.Now().UnixMicro(), 10)).Result()
		assert.Equal(t, int64(1), count)
	}()
}
