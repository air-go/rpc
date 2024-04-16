package slidingwindow

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/emirpasic/gods/queues/arrayqueue"
	"github.com/why444216978/go-util/assert"

	"github.com/air-go/rpc/library/limiter"
)

type node struct {
	t time.Time
}

type options struct {
	window time.Duration
	limit  int
	clock  clock.Clock
}

func defaultOptions() *options {
	return &options{
		window: time.Second,
		limit:  math.MaxInt,
	}
}

type OptionFunc func(o *options)

func WithClock(clock clock.Clock) OptionFunc {
	return func(o *options) { o.clock = clock }
}

func WithWindow(window time.Duration) OptionFunc {
	return func(o *options) { o.window = window }
}

func WithLimit(limit int) OptionFunc {
	return func(o *options) { o.limit = limit }
}

type slidingWindow struct {
	*options
	limiters sync.Map
}

var _ limiter.Limiter = (*slidingWindow)(nil)

func NewSlidingWindow(opts ...OptionFunc) *slidingWindow {
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	return &slidingWindow{
		options: opt,
	}
}

func (sw *slidingWindow) Allow(ctx context.Context, key string, opts ...limiter.AllowOptionFunc) (bool, error) {
	opt := &limiter.AllowOptions{}
	for _, o := range opts {
		o(opt)
	}

	l := sw.getLimiter(key)

	return l.allow(sw.now()), nil
}

func (sw *slidingWindow) SetWindow(ctx context.Context, key string, window time.Duration) {
	sw.getLimiter(key).setWindow(window)
}

func (sw *slidingWindow) SetLimit(ctx context.Context, key string, limit int) {
	sw.getLimiter(key).setLimit(limit)
}

func (sw *slidingWindow) getLimiter(key string) *keyLimiter {
	l, ok := sw.limiters.Load(key)
	if ok {
		return l.(*keyLimiter)
	}

	lim := newKeyLimiter(sw.window, sw.limit)
	sw.limiters.Store(key, lim)
	return lim
}

func (sl *slidingWindow) now() time.Time {
	if assert.IsNil(sl.clock) {
		return time.Now()
	}
	return sl.clock.Now()
}

type keyLimiter struct {
	mu     sync.RWMutex
	window time.Duration
	limit  int
	q      *arrayqueue.Queue
}

func newKeyLimiter(window time.Duration, limit int) *keyLimiter {
	return &keyLimiter{
		window: window,
		limit:  limit,
		q:      arrayqueue.New(),
	}
}

func (l *keyLimiter) allow(now time.Time) bool {
	l.mu.RLock()
	limit := l.limit
	window := l.window
	l.mu.RUnlock()

	// not full, access allowed
	if l.q.Size() < limit {
		l.q.Enqueue(&node{t: now})
		return true
	}

	// take out the earliest one
	early, _ := l.q.Peek()
	first := early.(*node)

	// the first request is still in the time window, access denied
	if now.Add(-window).Before(first.t) {
		return false
	}

	// pop the first request
	_, _ = l.q.Dequeue()
	l.q.Enqueue(&node{t: now})

	return true
}

func (l *keyLimiter) setLimit(limit int) {
	l.mu.Lock()
	l.limit = limit
	l.mu.Unlock()
}

func (l *keyLimiter) setWindow(window time.Duration) {
	l.mu.Lock()
	l.window = window
	l.mu.Unlock()
}
