package leakybucket

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/air-go/rpc/library/limiter"
	"github.com/benbjohnson/clock"
	"github.com/go-redis/redis/v8"
	"github.com/why444216978/go-util/assert"
)

var script = `local key           = KEYS[1] --键名，格式hash
local volume        = tonumber(KEYS[2]) --桶容量
local rate          = tonumber(KEYS[3]) --桶流出速率
local request_count = tonumber(KEYS[4]) --请求的增长数
local current_time  = tonumber(KEYS[5]) --当前时间戳
local ttl = math.floor((volume/rate)*2) --有效期

--最低有效期60s，避免请求处理时过期导致异常
if ttl < 60 then
	ttl = 60
end

--不存在则设置
if tonumber(redis.call('exists', key)) == 0 then
    redis.call('hset', key, 'last_time', current_time)
    redis.call('hset', key, 'count', 0)
end
--存在马上续期，避免下面处理时过期异常
redis.call("expire", key, ttl)
 
--延迟漏出：（当前时间 - 上次时间）* 每秒流出速率
local last_time     = tonumber(redis.call('hget', key, 'last_time'))
local current_count = tonumber(redis.call('hget', key, 'count'))
local leak_count    = (current_time - last_time) * rate
if leak_count > current_count then
    leak_count = current_count
end
local remain_count  = current_count - leak_count + request_count
if remain_count <= 0 then
    remain_count = 0 
end
 
redis.call('hset', key, 'volume', volume)
redis.call('hset', key, 'rate', rate)
redis.call('hset', key, 'last_time', current_time)
redis.call("expire", key, ttl)
 
if remain_count > volume then
    redis.call('hset', key, 'count', volume)
    return 0
end
 
redis.call('hset', key, 'count', remain_count)
 
return 1`

type options struct {
	defaultVolume int
	defaultRate   int
	clock         clock.Clock
}

type OptionFunc func(o *options)

func WithClock(clock clock.Clock) OptionFunc {
	return func(o *options) { o.clock = clock }
}

func WithVolume(volume int) OptionFunc {
	return func(o *options) { o.defaultVolume = volume }
}

func defaultOptions() *options {
	return &options{
		defaultVolume: 3000,
		defaultRate:   1,
	}
}

type leakyBucket struct {
	*options
	getClient func() *redis.Client
	limiters  sync.Map
}

var _ limiter.Limiter = (*leakyBucket)(nil)

func NewLeakyBucket(getClient func() *redis.Client, opts ...OptionFunc) (*leakyBucket, error) {
	if assert.IsNil(getClient) {
		return nil, errors.New("distributed sliding log limiter getClient is nil")
	}

	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	return &leakyBucket{
		options:   opt,
		getClient: getClient,
	}, nil
}

func (lb *leakyBucket) Allow(ctx context.Context, key string, opts ...limiter.AllowOptionFunc) (bool, error) {
	opt := &limiter.AllowOptions{}
	for _, o := range opts {
		o(opt)
	}

	count := 1
	if opt.Count > 0 {
		count = opt.Count
	}

	return lb.getLimiter(key).allow(ctx, lb.now(), count, lb.getClient)
}

func (lb *leakyBucket) SetVolume(ctx context.Context, key string, volume int) {
	lb.getLimiter(key).setVolume(volume)
}

func (lb *leakyBucket) SetRate(ctx context.Context, key string, rate int) {
	lb.getLimiter(key).setRate(rate)
}

func (lb *leakyBucket) getLimiter(key string) *keyLimiter {
	l, ok := lb.limiters.Load(key)
	if ok {
		return l.(*keyLimiter)
	}

	lim := newKeyLimiter(key, lb.defaultVolume, lb.defaultRate)
	lb.limiters.Store(key, lim)
	return lim
}

func (lb *leakyBucket) now() time.Time {
	if assert.IsNil(lb.clock) {
		return time.Now()
	}
	return lb.clock.Now()
}

type keyLimiter struct {
	mu     sync.RWMutex
	key    string
	volume int
	rate   int
}

func newKeyLimiter(key string, volume, rate int) *keyLimiter {
	return &keyLimiter{
		key:    key,
		volume: volume,
		rate:   rate,
	}
}

func (l *keyLimiter) allow(ctx context.Context, now time.Time, n int, getClient func() *redis.Client) (bool, error) {
	l.mu.RLock()
	volume := l.volume
	rate := l.rate
	l.mu.RUnlock()

	res, err := getClient().Do(ctx, "EVAL", script, 5, l.key,
		volume, rate, n, now.Unix()).Result()
	if err != nil {
		return false, err
	}

	if fmt.Sprint(res) == "1" {
		return true, nil
	}

	return false, nil
}

func (l *keyLimiter) setRate(rate int) {
	l.mu.Lock()
	l.rate = rate
	l.mu.Unlock()
}

func (l *keyLimiter) setVolume(volume int) {
	l.mu.Lock()
	l.volume = volume
	l.mu.Unlock()
}
