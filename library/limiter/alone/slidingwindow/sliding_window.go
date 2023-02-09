package slidingwindow

import (
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/emirpasic/gods/queues/arrayqueue"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen -package slidingwindow -destination=./sliding_window_mock.go  -source=sliding_window.go -build_flags=-mod=mod
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
	return &option{
		clock: clock.New(),
	}
}

type slidingWindow struct {
	mu     sync.Mutex
	opts   *option
	window time.Duration
	count  int
	q      *arrayqueue.Queue
}

func NewSlidingWindow(count int, window time.Duration, opts ...Option) *slidingWindow {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	return &slidingWindow{
		opts:   opt,
		window: window,
		count:  count,
		q:      arrayqueue.New(),
	}
}

func (sw *slidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := sw.opts.clock.Now()

	// not full, access allowed
	if sw.q.Size() < sw.count {
		sw.q.Enqueue(&node{t: now})
		return true
	}

	// take out the earliest one
	early, _ := sw.q.Peek()
	first := early.(*node)

	// the first request is still in the time window, access denied
	if now.Add(-sw.window).Before(first.t) {
		return false
	}

	// pop the first request
	_, _ = sw.q.Dequeue()
	sw.q.Enqueue(&node{t: now})

	return true
}

func (t *slidingWindow) SetLimit(count int) {
	t.mu.Lock()
	t.count = count
	t.mu.Unlock()
}

func (t *slidingWindow) SetWindow(window time.Duration) {
	t.mu.Lock()
	t.window = window
	t.mu.Unlock()
}
