package leakybucket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-redis/redis/v8"
	"github.com/why444216978/go-util/assert"
)

type LeakyBucket interface {
	Allow(ctx context.Context, key string, requestCount int) (bool, error)
	SetRate(rate int)
	SetVolume(burst int)
}

var script = `local key           = KEYS[1] --键名，格式hash
local volume        = tonumber(KEYS[2]) --桶容量
local rate          = tonumber(KEYS[3]) --桶流出速率
local request_count = tonumber(KEYS[4]) --请求的增长数
local current_time  = tonumber(KEYS[5]) --当前时间戳
local ttl = math.floor((volume/rate)*2) --有效期

--容量小于流出速率, 只需要拦住1s的请求即可
if ttl < 1 then
	ttl = 1
end

--不存在则设置
if tonumber(redis.call('exists', key)) == 0 then
    redis.call('hset', key, 'last_time', current_time)
    redis.call('hset', key, 'count', 0)
end
 
--延迟漏出：（当前时间 - 上次时间）* 每秒流出速率
local last_time     = tonumber(redis.call('hget', key, 'last_time'))
local current_count = tonumber(redis.call('hget', key, 'count'))
local leak_count    = (current_time - last_time) * rate
local remain_count  = current_count + request_count - leak_count
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

type option struct {
	clock clock.Clock
}

type Option func(o *option)

func WithClock(clock clock.Clock) Option {
	return func(o *option) { o.clock = clock }
}

func defaultOption() *option {
	return &option{}
}

type leakyBucket struct {
	opts   *option
	mu     sync.RWMutex
	volume int
	rate   int
	client *redis.Client
}

func NewLeakyBucket(rate, volume int, client *redis.Client, opts ...Option) LeakyBucket {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	return &leakyBucket{
		opts:   opt,
		client: client,
		rate:   rate,
		volume: volume,
	}
}

func (lb *leakyBucket) Allow(ctx context.Context, key string, requestCount int) (ok bool, err error) {
	lb.mu.RLock()
	rate := lb.rate
	volume := lb.volume
	lb.mu.RUnlock()

	res, err := lb.client.Do(ctx, "EVAL", script, 5, key, volume, rate, requestCount, lb.now().Unix()).Result()
	if err != nil {
		return
	}

	if fmt.Sprint(res) == "1" {
		return true, nil
	}

	return
}

func (lb *leakyBucket) SetRate(rate int) {
	lb.mu.Lock()
	lb.rate = rate
	lb.mu.Unlock()
}

func (lb *leakyBucket) SetVolume(burst int) {
	lb.mu.Lock()
	lb.volume = burst
	lb.mu.Unlock()
}

func (sl *leakyBucket) now() time.Time {
	if assert.IsNil(sl.opts.clock) {
		return time.Now()
	}
	return sl.opts.clock.Now()
}
