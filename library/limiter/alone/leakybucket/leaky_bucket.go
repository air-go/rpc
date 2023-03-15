// Reference uber ratelimit, the different is support no wait take func
package leakybucket

import (
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/why444216978/go-util/assert"
)

type LeakyBucket interface {
	Allow() bool
	SetRate(rate int)
	SetVolume(volume int)
}

type option struct {
	// flexible control request interval
	// if 3 times are allowed to pass in 3 seconds, per is 3 second
	per   time.Duration
	clock clock.Clock
}

type Option func(o *option)

func WithPer(per time.Duration) Option {
	return func(o *option) { o.per = per }
}

func WithClock(clock clock.Clock) Option {
	return func(o *option) { o.clock = clock }
}

func defaultOption() *option {
	return &option{
		per: time.Second,
	}
}

type leakyBucket struct {
	opts            *option
	mu              sync.RWMutex
	last            time.Time
	volume          int
	current         int
	requestInterval time.Duration // per / rate
}

func NewLeakyBucket(rate, volume int, opts ...Option) LeakyBucket {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	l := &leakyBucket{
		opts:   opt,
		volume: volume,
	}

	l.SetRate(rate)

	return l
}

func (lb *leakyBucket) Allow() bool {
	lb.mu.RLock()
	requestInterval := lb.requestInterval
	volume := lb.volume
	lb.mu.RUnlock()

	now := lb.now()
	defer func() {
		lb.last = now
	}()

	// constant rate outflow
	c := int(now.Sub(lb.last) / requestInterval)
	lb.current -= c
	if lb.current < 0 {
		lb.current = 0
	}

	// bucket full overflow, access denied
	if lb.current >= volume {
		return false
	}

	// only one can pass at a time
	lb.current += 1

	return true
}

func (lb *leakyBucket) SetRate(rate int) {
	lb.mu.Lock()
	requestInterval := lb.opts.per / time.Duration(rate)
	lb.requestInterval = requestInterval
	lb.mu.Unlock()
}

func (lb *leakyBucket) SetVolume(volume int) {
	lb.mu.Lock()
	lb.volume = volume
	lb.mu.Unlock()
}

func (sl *leakyBucket) now() time.Time {
	if assert.IsNil(sl.opts.clock) {
		return time.Now()
	}
	return sl.opts.clock.Now()
}
