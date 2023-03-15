package slidinglog

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"

	"github.com/air-go/rpc/library/limiter"
)

type slidingLogEntry struct {
	allow bool
	err   error
}

var _ limiter.Entry = (*slidingLogEntry)(nil)

func (e *slidingLogEntry) Allow() bool {
	return e.allow
}

func (e *slidingLogEntry) Finish() {}

func (e *slidingLogEntry) Error() error { return e.err }

type slidingLogLimiter struct {
	client   *redis.Client
	limiters sync.Map
}

var _ limiter.Limiter = (*slidingLogLimiter)(nil)

func NewLimiter(client *redis.Client) *slidingLogLimiter {
	return &slidingLogLimiter{
		client:   client,
		limiters: sync.Map{},
	}
}

func (l *slidingLogLimiter) Check(ctx context.Context, r limiter.Resource) limiter.Entry {
	ok, err := l.getLimiter(r).Allow(ctx, r.Name)
	return &slidingLogEntry{
		allow: ok,
		err:   err,
	}
}

func (l *slidingLogLimiter) SetLimit(ctx context.Context, r limiter.Resource) {
	l.getLimiter(r).SetLimit(int64(r.Limit))
}

func (l *slidingLogLimiter) SetBurst(ctx context.Context, r limiter.Resource) {}

func (l *slidingLogLimiter) SetWindow(ctx context.Context, r limiter.Resource) {
	l.getLimiter(r).SetWindow(r.Window)
}

func (l *slidingLogLimiter) getLimiter(r limiter.Resource) (lim SlidingLog) {
	val, ok := l.limiters.Load(r.Name)
	if !ok {
		lim = NewSlidingLog(int64(r.Limit), r.Window, l.client)
		l.limiters.Store(r.Name, lim)
		return
	}

	if lim, ok = val.(SlidingLog); !ok {
		lim = NewSlidingLog(int64(r.Limit), r.Window, l.client)
		l.limiters.Store(r.Name, lim)
		return
	}

	return
}
