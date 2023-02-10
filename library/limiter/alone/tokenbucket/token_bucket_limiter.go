package tokenbucket

import (
	"context"
	"sync"

	"golang.org/x/time/rate"

	"github.com/air-go/rpc/library/limiter"
)

type tokenBucketEntry struct {
	allow bool
}

var _ limiter.Entry = (*tokenBucketEntry)(nil)

func (e *tokenBucketEntry) Allow() bool {
	return e.allow
}

func (e *tokenBucketEntry) Finish() {}

func (e *tokenBucketEntry) Error() error { return nil }

type tokenBucketLimiter struct {
	limiters sync.Map // key resource name,value *rate.limiter
}

var _ limiter.Limiter = (*tokenBucketLimiter)(nil)

func NewLimiter() *tokenBucketLimiter {
	return &tokenBucketLimiter{
		limiters: sync.Map{},
	}
}

func (l *tokenBucketLimiter) Check(ctx context.Context, r limiter.Resource) limiter.Entry {
	return &tokenBucketEntry{allow: l.getLimiter(r).Allow()}
}

func (l *tokenBucketLimiter) SetLimit(ctx context.Context, r limiter.Resource) {
	l.getLimiter(r).SetLimit(rate.Limit(r.Limit))
}

func (l *tokenBucketLimiter) SetBurst(ctx context.Context, r limiter.Resource) {
	l.getLimiter(r).SetBurst(r.Burst)
}

func (l *tokenBucketLimiter) getLimiter(r limiter.Resource) (lim *rate.Limiter) {
	val, ok := l.limiters.Load(r.Name)
	if !ok {
		lim = rate.NewLimiter(rate.Limit(r.Limit), r.Burst)
		l.limiters.Store(r.Name, lim)
		return
	}

	if lim, ok = val.(*rate.Limiter); !ok {
		lim = rate.NewLimiter(rate.Limit(r.Limit), r.Burst)
		l.limiters.Store(r.Name, lim)
		return
	}

	return
}
