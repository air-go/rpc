package localbloom

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/why444216978/go-util/assert"
	"github.com/why444216978/go-util/nopanic"

	lbloom "github.com/air-go/rpc/library/bloom"
)

type options struct {
	n               uint
	fp              float64
	clock           clock.Clock
	releaseDuration time.Duration
}

func defaultOptions() *options {
	return &options{
		n:               10000,
		fp:              0.01,
		releaseDuration: time.Minute,
	}
}

type OptionFunc func(*options)

func SetEstimateParameters(n uint, fp float64) OptionFunc {
	return func(o *options) {
		o.n = n
		o.fp = fp
	}
}

func SetClock(c clock.Clock) OptionFunc {
	return func(o *options) { o.clock = c }
}

func SetReleaseDuration(d time.Duration) OptionFunc {
	return func(o *options) { o.releaseDuration = d }
}

type MemoryBloom struct {
	*options
	blooms sync.Map
}

var _ lbloom.Bloom = (*MemoryBloom)(nil)

func NewMemoryBloom(opts ...OptionFunc) *MemoryBloom {
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	mb := &MemoryBloom{
		options: opt,
		blooms:  sync.Map{},
	}
	mb.tryRelease()

	return mb
}

func (mb *MemoryBloom) Add(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	mb.getBloom(key).add(data, mb.now(), ttl)
	return nil
}

func (mb *MemoryBloom) Check(ctx context.Context, key string, data []byte, ttl time.Duration) (bool, error) {
	return mb.getBloom(key).check(data, mb.now(), ttl), nil
}

func (mb *MemoryBloom) CheckAndAdd(ctx context.Context, key string, data []byte, ttl time.Duration) (bool, error) {
	return mb.getBloom(key).checkAndAdd(data, mb.now(), ttl), nil
}

func (mb *MemoryBloom) getBloom(k string) *keyBloom {
	v, ok := mb.blooms.Load(k)
	if ok {
		return v.(*keyBloom)
	}

	b := newKeyBloom(k, mb.n, mb.fp)
	mb.blooms.Store(k, b)
	return b
}

func (mb *MemoryBloom) tryRelease() {
	go nopanic.GoVoid(context.Background(), func() {
		t := time.NewTicker(mb.releaseDuration)
		for range t.C {
			mb.blooms.Range(func(k, v any) bool {
				if mb.now().After(v.(*keyBloom).getExpireAt()) {
					mb.blooms.Delete(k)
				}
				return true
			})
		}
	})
}

func (mb *MemoryBloom) now() time.Time {
	if assert.IsNil(mb.clock) {
		return time.Now()
	}
	return mb.clock.Now()
}

type keyBloom struct {
	mu       sync.Mutex
	k        string
	n        uint
	fp       float64
	b        *bloom.BloomFilter
	expireAt time.Time
}

func newKeyBloom(k string, n uint, fp float64) *keyBloom {
	return &keyBloom{
		k:  k,
		n:  n,
		fp: fp,
		b:  bloom.NewWithEstimates(n, fp),
	}
}

func (b *keyBloom) add(data []byte, now time.Time, ttl time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.expireAt = now.Add(ttl)
	b.b.Add(data)
}

func (b *keyBloom) check(data []byte, now time.Time, ttl time.Duration) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.expireAt = now.Add(ttl)
	return b.b.Test(data)
}

func (b *keyBloom) checkAndAdd(data []byte, now time.Time, ttl time.Duration) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.expireAt = now.Add(ttl)
	return b.b.TestAndAdd(data)
}

func (b *keyBloom) getExpireAt() time.Time {
	b.mu.Lock()
	t := b.expireAt
	b.mu.Unlock()
	return t
}
