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
	clock  clock.Clock
	keyTTL time.Duration
	window time.Duration
	limit  int
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

	return sl.getLimiter(key).allow(ctx, sl.now(), n, sl.getClient)

	// end := sl.now()
	// begin := end.Add(-window)

	// newCtx := ucontext.RemoveCancel(ctx)
	// defer func() {
	// 	go nopanic.GoVoid(ucontext.RemoveCancel(ctx), func() {
	// 		timer := time.NewTimer(sl.keyTTL)
	// 		defer timer.Stop()

	// 		_ = sl.getClient().ExpireAt(newCtx, key, end.Add(sl.keyTTL))
	// 	})
	// }()

	// count, err := sl.getClient().ZCount(ctx, key, strconv.FormatInt(begin.UnixMicro(), 10), strconv.FormatInt(end.UnixMicro(), 10)).Result()
	// if err != nil {
	// 	return false, err
	// }

	// if count >= int64(limit) {
	// 	return false, err
	// }

	// g := nopanic.New(ctx)
	// g.Go(func() error {
	// 	return sl.getClient().ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(begin.UnixMicro()-1, 10)).Err()
	// })
	// g.Go(func() error {
	// 	return sl.getClient().ZAdd(ctx, key, &redis.Z{Score: float64(end.UnixMicro()), Member: end.UnixMicro()}).Err()
	// })
	// if err = g.Wait(); err != nil {
	// 	return false, err
	// }

	// return true, nil
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

	lim := newKeyLimiter(key, sl.keyTTL, sl.window, sl.limit)
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
	mu     sync.RWMutex
	key    string
	window time.Duration
	keyTTL time.Duration
	limit  int
}

func newKeyLimiter(key string, keyTTL, window time.Duration, limit int) *keyLimiter {
	return &keyLimiter{
		key:    key,
		window: window,
		keyTTL: keyTTL,
		limit:  limit,
	}
}

func (l *keyLimiter) allow(ctx context.Context, now time.Time, n int, getClient func() *redis.Client) (bool, error) {
	l.mu.RLock()
	limit := l.limit
	window := l.window
	l.mu.RUnlock()

	end := now
	begin := end.Add(-window)

	newCtx := ucontext.RemoveCancel(ctx)
	defer func() {
		go nopanic.GoVoid(ucontext.RemoveCancel(ctx), func() {
			timer := time.NewTimer(l.keyTTL)
			defer timer.Stop()

			_ = getClient().ExpireAt(newCtx, l.key, end.Add(l.keyTTL))
		})
	}()

	count, err := getClient().ZCount(ctx, l.key,
		strconv.FormatInt(begin.UnixMicro(), 10), strconv.FormatInt(end.UnixMicro(), 10)).Result()
	if err != nil {
		return false, err
	}

	if count >= int64(limit) {
		return false, err
	}

	g := nopanic.New(ctx)
	g.Go(func() error {
		return getClient().ZRemRangeByScore(ctx, l.key, "0", strconv.FormatInt(begin.UnixMicro()-1, 10)).Err()
	})
	g.Go(func() error {
		return getClient().ZAdd(ctx, l.key, &redis.Z{Score: float64(end.UnixMicro()), Member: end.UnixMicro()}).Err()
	})
	if err = g.Wait(); err != nil {
		return false, err
	}

	return true, nil
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
