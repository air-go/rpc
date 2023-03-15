package slidinglog

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-redis/redis/v8"
	"github.com/why444216978/go-util/assert"
)

type SlidingLog interface {
	Allow(ctx context.Context, key string) (bool, error)
	SetLimit(limit int64)
	SetWindow(w time.Duration)
}

type option struct {
	clock clock.Clock
}

type Option func(o *option)

func WithClock(clock clock.Clock) Option {
	return func(o *option) { o.clock = clock }
}

func defaultOption() *option {
	return &option{}
}

type slidingLog struct {
	opts   *option
	mu     sync.RWMutex
	limit  int64
	window time.Duration
	client *redis.Client
}

var _ SlidingLog = (*slidingLog)(nil)

func NewSlidingLog(limit int64, w time.Duration, client *redis.Client, opts ...Option) *slidingLog {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	l := &slidingLog{
		opts:   opt,
		client: client,
		limit:  limit,
		window: w,
	}

	return l
}

func (sl *slidingLog) Allow(ctx context.Context, key string) (ok bool, err error) {
	sl.mu.RLock()
	window := sl.window
	limit := sl.limit
	sl.mu.RUnlock()

	end := sl.now()
	begin := end.Add(-window)

	if err = sl.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(begin.UnixMicro()-1, 10)).Err(); err != nil {
		return
	}

	count, err := sl.client.ZCount(ctx, key, strconv.FormatInt(begin.UnixMicro(), 10), strconv.FormatInt(end.UnixMicro(), 10)).Result()
	if err != nil {
		return
	}

	if count >= limit {
		return
	}

	if err = sl.client.ZAdd(ctx, key, &redis.Z{Score: float64(end.UnixMicro()), Member: end.UnixMicro()}).Err(); err != nil {
		return
	}

	return true, nil
}

func (sl *slidingLog) SetLimit(limit int64) {
	sl.mu.Lock()
	sl.limit = limit
	sl.mu.Unlock()
}

func (sl *slidingLog) SetWindow(w time.Duration) {
	sl.mu.Lock()
	sl.window = w
	sl.mu.Unlock()
}

func (sl *slidingLog) now() time.Time {
	if assert.IsNil(sl.opts.clock) {
		return time.Now()
	}
	return sl.opts.clock.Now()
}
