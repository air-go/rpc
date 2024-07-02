package slidinglog

import (
	"context"
	"errors"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-redis/redis/v8"
	"github.com/why444216978/go-util/assert"
	ucontext "github.com/why444216978/go-util/context"
	"github.com/why444216978/go-util/nopanic"

	"github.com/air-go/rpc/library/limiter"
)

type SlidingLog interface {
	Allow(ctx context.Context, key string) (bool, error)
	SetLimit(limit int64)
	SetWindow(w time.Duration)
}

type option struct {
	clock       clock.Clock
	keyTTL      time.Duration
	window      time.Duration
	leftWindow  time.Time
	rightWindow time.Time
	limit       int
}

func defaultOption() *option {
	return &option{
		keyTTL: time.Hour,
		window: time.Second,
		limit:  math.MaxInt,
	}
}

type Option func(o *option)

func WithClock(clock clock.Clock) Option {
	return func(o *option) { o.clock = clock }
}

func WithKeyTTL(keyTTL time.Duration) Option {
	return func(o *option) { o.keyTTL = keyTTL }
}

func WithLimit(l int) Option {
	return func(o *option) { o.limit = l }
}

func WithWindow(w time.Duration) Option {
	return func(o *option) { o.window = w }
}

func WithFledWindow(l, r time.Time) Option {
	return func(o *option) {
		o.leftWindow = l
		o.rightWindow = r
	}
}

type slidingLog struct {
	*option
	limiters  sync.Map
	getClient func() *redis.Client
}

var _ limiter.Limiter = (*slidingLog)(nil)

func NewSlidingLog(getClient func() *redis.Client, opts ...Option) (*slidingLog, error) {
	if assert.IsNil(getClient) {
		return nil, errors.New("distributed sliding log limiter getClient is nil")
	}

	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	l := &slidingLog{
		option:    opt,
		getClient: getClient,
	}

	return l, nil
}

func (sl *slidingLog) Allow(ctx context.Context, key string, opts ...limiter.AllowOptionFunc) (bool, error) {
	opt := &limiter.AllowOptions{}
	for _, o := range opts {
		o(opt)
	}
	n := 1
	if opt.Count > 0 {
		n = opt.Count
	}

	return sl.getLimiter(key).allow(ctx, sl.now(), n, opt.FixedWindow, sl.getClient)
}

func (sl *slidingLog) SetWindow(ctx context.Context, key string, window time.Duration) {
	sl.getLimiter(key).setWindow(window)
}

func (sl *slidingLog) SetLimit(ctx context.Context, key string, limit int) {
	sl.getLimiter(key).setLimit(limit)
}

func (sl *slidingLog) getLimiter(key string) *keyLimiter {
	l, ok := sl.limiters.Load(key)
	if ok {
		return l.(*keyLimiter)
	}

	lim := newKeyLimiter(key, sl.option)
	sl.limiters.Store(key, lim)
	return lim
}

func (sl *slidingLog) now() time.Time {
	if assert.IsNil(sl.clock) {
		return time.Now()
	}
	return sl.clock.Now()
}

type keyLimiter struct {
	mu          sync.RWMutex
	key         string
	window      time.Duration
	leftWindow  time.Time
	rightWindow time.Time
	keyTTL      time.Duration
	limit       int
}

func newKeyLimiter(key string, o *option) *keyLimiter {
	return &keyLimiter{
		key:         key,
		window:      o.window,
		leftWindow:  o.leftWindow,
		rightWindow: o.rightWindow,
		keyTTL:      o.keyTTL,
		limit:       o.limit,
	}
}

func (l *keyLimiter) allow(ctx context.Context, now time.Time, n int, fixedWindow bool, getClient func() *redis.Client) (bool, error) {
	if fixedWindow {
		return l.allowFixedWindow(ctx, now, n, getClient)
	}

	return l.allowDuration(ctx, now, n, getClient)
}

func (l *keyLimiter) duration2Window(now time.Time, d time.Duration) (left, right time.Time) {
	return now.Add(-d), now
}

func (l *keyLimiter) allowFixedWindow(ctx context.Context, now time.Time, n int, getClient func() *redis.Client) (bool, error) {
	l.mu.RLock()
	limit := l.limit
	left := l.leftWindow
	right := l.rightWindow
	l.mu.RUnlock()

	if left.IsZero() || right.IsZero() {
		return false, errors.New("left window or right window empty")
	}

	if now.Before(left) || now.After(right) {
		return false, errors.New("now time illegal")
	}

	return l.handle(ctx, left, right, now, limit, n, getClient)
}

func (l *keyLimiter) allowDuration(ctx context.Context, now time.Time, n int, getClient func() *redis.Client) (bool, error) {
	l.mu.RLock()
	limit := l.limit
	window := l.window
	l.mu.RUnlock()

	left, right := l.duration2Window(now, window)

	return l.handle(ctx, left, right, now, limit, n, getClient)
}

func (l *keyLimiter) handle(ctx context.Context, left, right, now time.Time, limit, n int, getClient func() *redis.Client) (bool, error) {
	newCtx := ucontext.RemoveCancel(ctx)
	defer func() {
		go nopanic.GoVoid(ucontext.RemoveCancel(ctx), func() {
			timer := time.NewTimer(l.keyTTL)
			defer timer.Stop()

			_ = getClient().ExpireAt(newCtx, l.key, right.Add(l.keyTTL))
		})
	}()

	if err := getClient().ZRemRangeByScore(ctx, l.key, "0", strconv.FormatInt(left.UnixMicro()-1, 10)).Err(); err != nil {
		return false, err
	}

	count, err := getClient().ZCount(ctx, l.key,
		strconv.FormatInt(left.UnixMicro(), 10), strconv.FormatInt(right.UnixMicro(), 10)).Result()
	if err != nil {
		return false, err
	}

	if count >= int64(limit) {
		return false, err
	}

	for i := 1; i <= n; i++ {
		e := getClient().ZAdd(ctx, l.key, &redis.Z{Score: float64(now.UnixMicro()), Member: now.UnixMicro()}).Err()
		if e != nil {
			err = e
		}
	}

	return true, err
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

func (l *keyLimiter) setLeftWindow(lf time.Time) {
	l.mu.Lock()
	l.leftWindow = lf
	l.mu.Unlock()
}

func (l *keyLimiter) setRightWindow(rw time.Time) {
	l.mu.Lock()
	l.rightWindow = rw
	l.mu.Unlock()
}
