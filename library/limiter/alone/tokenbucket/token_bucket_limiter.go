package tokenbucket

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/why444216978/go-util/assert"
	"golang.org/x/time/rate"

	"github.com/air-go/rpc/library/limiter"
)

type options struct {
	limit        rate.Limit
	defaultBurst int
	clock        clock.Clock
}

func defaultOptions() *options {
	return &options{
		limit:        rate.Inf,
		defaultBurst: math.MaxInt,
	}
}

type OptionFunc func(o *options)

func WithClock(clock clock.Clock) OptionFunc {
	return func(o *options) { o.clock = clock }
}

func WithBurst(burst int) OptionFunc {
	return func(o *options) { o.defaultBurst = burst }
}

func WithLimit(limit float64) OptionFunc {
	return func(o *options) { o.limit = rate.Limit(limit) }
}

type tokenBucketLimiter struct {
	*options
	limiters sync.Map
}

var _ limiter.Limiter = (*tokenBucketLimiter)(nil)

func NewTokenBucket(opts ...OptionFunc) *tokenBucketLimiter {
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	return &tokenBucketLimiter{
		options:  opt,
		limiters: sync.Map{},
	}
}

func (tb *tokenBucketLimiter) Allow(ctx context.Context, key string, opts ...limiter.AllowOptionFunc) (bool, error) {
	opt := &limiter.AllowOptions{}
	for _, o := range opts {
		o(opt)
	}

	count := 1
	if opt.Count > 0 {
		count = opt.Count
	}

	return tb.getLimiter(key).AllowN(tb.now(), count), nil
}

func (tb *tokenBucketLimiter) SetLimit(ctx context.Context, key string, limit rate.Limit) {
	tb.getLimiter(key).SetLimit(limit)
}

func (tb *tokenBucketLimiter) SetBurst(ctx context.Context, key string, burst int) {
	tb.getLimiter(key).SetBurst(burst)
}

func (tb *tokenBucketLimiter) getLimiter(key string) *rate.Limiter {
	val, ok := tb.limiters.Load(key)
	if !ok {
		l := rate.NewLimiter(tb.limit, tb.defaultBurst)
		tb.limiters.Store(key, l)
		return l
	}

	return val.(*rate.Limiter)
}

func (sl *tokenBucketLimiter) now() time.Time {
	if assert.IsNil(sl.clock) {
		return time.Now()
	}
	return sl.clock.Now()
}
