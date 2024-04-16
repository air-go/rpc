// Reference uber ratelimit, the different is support no wait take func
package leakybucket

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/why444216978/go-util/assert"

	"github.com/air-go/rpc/library/limiter"
)

type options struct {
	defaultVolume int
	defaultPer    time.Duration
	defaultRate   int
	clock         clock.Clock
}

type OptionFunc func(o *options)

func WithVolume(volume int) OptionFunc {
	return func(o *options) { o.defaultVolume = volume }
}

func WithRate(per time.Duration, rate int) OptionFunc {
	return func(o *options) {
		o.defaultPer = per
		o.defaultRate = rate
	}
}

func WithClock(clock clock.Clock) OptionFunc {
	return func(o *options) { o.clock = clock }
}

func defaultOptions() *options {
	return &options{
		defaultVolume: math.MaxInt,
		defaultPer:    time.Second,
		defaultRate:   math.MaxInt,
	}
}

type leakyBucket struct {
	*options
	limiters sync.Map
}

var _ limiter.Limiter = (*leakyBucket)(nil)

func NewLeakyBucket(opts ...OptionFunc) *leakyBucket {
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	l := &leakyBucket{
		options:  opt,
		limiters: sync.Map{},
	}

	return l
}

func (lb *leakyBucket) Allow(ctx context.Context, key string, opts ...limiter.AllowOptionFunc) (bool, error) {
	opt := &limiter.AllowOptions{}
	for _, o := range opts {
		o(opt)
	}

	count := 1
	if opt.Count != 0 {
		count = opt.Count
	}

	return lb.getLimiter(key).allow(lb.now(), count), nil
}

func (lb *leakyBucket) SetVolume(ctx context.Context, key string, volume int) {
	lb.getLimiter(key).setVolume(volume)
}

func (lb *leakyBucket) SetRate(ctx context.Context, key string, per time.Duration, rate int) {
	lb.getLimiter(key).setRate(per, rate)
}

func (lb *leakyBucket) getLimiter(key string) *keyLimiter {
	l, ok := lb.limiters.Load(key)
	if ok {
		return l.(*keyLimiter)
	}

	lim := newKeyLimiter(lb.defaultVolume, lb.defaultRate, lb.defaultPer)
	lb.limiters.Store(key, lim)
	return lim
}

func (lb *leakyBucket) now() time.Time {
	if assert.IsNil(lb.clock) {
		return time.Now()
	}
	return lb.clock.Now()
}

type keyLimiter struct {
	mu              sync.Mutex
	last            time.Time
	current         int
	volume          int
	rate            int
	per             time.Duration
	requestInterval time.Duration
}

func newKeyLimiter(volume, rate int, per time.Duration) *keyLimiter {
	return &keyLimiter{
		volume:          volume,
		requestInterval: per / time.Duration(rate),
	}
}

func (l *keyLimiter) allow(now time.Time, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	defer func() {
		l.last = now
	}()

	if l.last.IsZero() {
		l.last = now
	}

	// constant rate outflow
	c := int(now.Sub(l.last) / l.requestInterval)
	l.current -= c
	if l.current < 0 {
		l.current = 0
	}

	// bucket full overflow, access denied
	if l.current+n > l.volume {
		return false
	}

	// add request number
	l.current += n

	return true
}

func (l *keyLimiter) setRate(per time.Duration, rate int) {
	l.mu.Lock()
	l.per = per
	l.rate = rate
	// Flexible control request interval.
	// Every per time interval outflow rate.
	l.requestInterval = per / time.Duration(rate)
	l.mu.Unlock()
}

func (l *keyLimiter) setVolume(volume int) {
	l.mu.Lock()
	l.volume = volume
	l.mu.Unlock()
}
