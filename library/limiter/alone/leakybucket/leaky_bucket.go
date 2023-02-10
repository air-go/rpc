// Reference uber ratelimit, the different is support no wait take func
package leakybucket

import (
	"sync"
	"time"

	"github.com/benbjohnson/clock"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen -package leakybucket -destination=./leaky_bucket_mock.go  -source=leaky_bucket.go -build_flags=-mod=mod
type LeakyBucket interface {
	Allow() bool
	SetLimit(rate int)
	SetBurst(burst int)
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
		per:   time.Second,
		clock: clock.New(),
	}
}

type leakyBucket struct {
	opts            *option
	mu              sync.Mutex
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
		last:   time.Now(),
		volume: volume,
	}

	l.SetLimit(rate)
	l.SetBurst(volume)

	return l
}

func (lb *leakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := lb.opts.clock.Now()
	defer func() {
		lb.last = now
	}()

	// constant rate outflow
	c := int(now.Sub(lb.last) / lb.requestInterval)
	lb.current -= c
	if lb.current < 0 {
		lb.current = 0
	}

	// bucket full overflow, access denied
	if lb.current >= lb.volume {
		return false
	}

	// only one can pass at a time
	lb.current += 1

	return true
}

func (lb *leakyBucket) SetLimit(rate int) {
	lb.mu.Lock()
	requestInterval := lb.opts.per / time.Duration(rate)
	lb.requestInterval = requestInterval
	lb.mu.Unlock()
}

func (lb *leakyBucket) SetBurst(burst int) {
	lb.mu.Lock()
	lb.volume = burst
	lb.mu.Unlock()
}
