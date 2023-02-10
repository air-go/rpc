package slidingwindow

import (
	"context"
	"sync"
	"time"

	"github.com/air-go/rpc/library/limiter"
)

type slidingWindowEntry struct {
	allow bool
}

var _ limiter.Entry = (*slidingWindowEntry)(nil)

func (e *slidingWindowEntry) Allow() bool {
	return e.allow
}

func (e *slidingWindowEntry) Finish() {}

func (e *slidingWindowEntry) Error() error { return nil }

type slidingWindowLimiter struct {
	limiters sync.Map // key resource name,value *rate.limiter
}

var _ limiter.Limiter = (*slidingWindowLimiter)(nil)

func NewLimiter() *slidingWindowLimiter {
	return &slidingWindowLimiter{
		limiters: sync.Map{},
	}
}

func (l *slidingWindowLimiter) Check(ctx context.Context, r limiter.Resource) limiter.Entry {
	return &slidingWindowEntry{allow: l.getLimiter(r).Allow()}
}

func (l *slidingWindowLimiter) SetLimit(ctx context.Context, r limiter.Resource) {
	l.getLimiter(r).SetLimit(r.Limit)
}

func (l *slidingWindowLimiter) SetBurst(ctx context.Context, r limiter.Resource) {
	l.getLimiter(r).SetWindow(l.burst2Window(r.Burst))
}

func (l *slidingWindowLimiter) getLimiter(r limiter.Resource) (lim SlidingWindow) {
	val, ok := l.limiters.Load(r.Name)
	if !ok {
		lim = NewSlidingWindow(r.Limit, l.burst2Window(r.Burst))
		l.limiters.Store(r.Name, lim)
		return
	}

	if lim, ok = val.(SlidingWindow); !ok {
		lim = NewSlidingWindow(r.Limit, l.burst2Window(r.Burst))
		l.limiters.Store(r.Name, lim)
		return
	}

	return
}

func (l *slidingWindowLimiter) burst2Window(burst int) time.Duration {
	return time.Duration(burst) * time.Millisecond
}
