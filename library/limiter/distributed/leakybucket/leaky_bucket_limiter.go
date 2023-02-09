package leakybucket

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"

	"github.com/air-go/rpc/library/limiter"
)

type leakyBucketEntry struct {
	allow bool
	err   error
}

var _ limiter.Entry = (*leakyBucketEntry)(nil)

func (e *leakyBucketEntry) Allow() bool {
	return e.allow
}

func (e *leakyBucketEntry) Finish() {}

func (e *leakyBucketEntry) Error() error { return e.err }

type leakyBucketLimiter struct {
	client   *redis.Client
	limiters sync.Map // key resource name,value *rate.limiter
}

var _ limiter.Limiter = (*leakyBucketLimiter)(nil)

func NewLimiter(client *redis.Client) *leakyBucketLimiter {
	return &leakyBucketLimiter{
		client:   client,
		limiters: sync.Map{},
	}
}

func (l *leakyBucketLimiter) Check(ctx context.Context, r limiter.Resource) limiter.Entry {
	ok, err := l.getLimiter(r).Allow(ctx, r.Name, 1)
	return &leakyBucketEntry{
		allow: ok,
		err:   err,
	}
}

func (l *leakyBucketLimiter) SetLimit(ctx context.Context, r limiter.Resource) {
	l.getLimiter(r).SetLimit(r.Limit)
}

func (l *leakyBucketLimiter) SetBurst(ctx context.Context, r limiter.Resource) {
	l.getLimiter(r).SetBurst(r.Burst)
}

func (l *leakyBucketLimiter) getLimiter(r limiter.Resource) (lim LeakyBucket) {
	val, ok := l.limiters.Load(r.Name)
	if !ok {
		lim = NewLeakyBucket(r.Limit, r.Burst, l.client)
		l.limiters.Store(r.Name, lim)
		return
	}

	if lim, ok = val.(LeakyBucket); !ok {
		lim = NewLeakyBucket(r.Limit, r.Burst, l.client)
		l.limiters.Store(r.Name, lim)
		return
	}

	return
}
