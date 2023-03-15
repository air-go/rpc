package slidingwindow

import (
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/emirpasic/gods/queues/arrayqueue"
	"github.com/why444216978/go-util/assert"
)

type SlidingWindow interface {
	Allow() bool
	SetLimit(rate int)
	SetWindow(window time.Duration)
}

type node struct {
	t time.Time
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

type slidingWindow struct {
	mu     sync.RWMutex
	opts   *option
	window time.Duration
	limit  int
	q      *arrayqueue.Queue
}

func NewSlidingWindow(limit int, window time.Duration, opts ...Option) *slidingWindow {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	return &slidingWindow{
		opts:   opt,
		window: window,
		limit:  limit,
		q:      arrayqueue.New(),
	}
}

func (sw *slidingWindow) Allow() bool {
	sw.mu.RLock()
	limit := sw.limit
	window := sw.window
	sw.mu.RUnlock()

	now := sw.now()

	// not full, access allowed
	if sw.q.Size() < limit {
		sw.q.Enqueue(&node{t: now})
		return true
	}

	// take out the earliest one
	early, _ := sw.q.Peek()
	first := early.(*node)

	// the first request is still in the time window, access denied
	if now.Add(-window).Before(first.t) {
		return false
	}

	// pop the first request
	_, _ = sw.q.Dequeue()
	sw.q.Enqueue(&node{t: now})

	return true
}

func (t *slidingWindow) SetLimit(limit int) {
	t.mu.Lock()
	t.limit = limit
	t.mu.Unlock()
}

func (t *slidingWindow) SetWindow(window time.Duration) {
	t.mu.Lock()
	t.window = window
	t.mu.Unlock()
}

func (sl *slidingWindow) now() time.Time {
	if assert.IsNil(sl.opts.clock) {
		return time.Now()
	}
	return sl.opts.clock.Now()
}
