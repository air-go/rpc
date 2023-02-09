// Reference uber ratelimit, the different is support no wait take func
package leakybucket

import (
	"context"
	"fmt"
	"sync"

	"github.com/benbjohnson/clock"
	"github.com/go-redis/redis/v8"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen -package leakybucket -destination=./leaky_bucket_mock.go  -source=leaky_bucket.go -build_flags=-mod=mod
type LeakyBucket interface {
	Allow(ctx context.Context, key string, requestCount int) (bool, error)
	SetLimit(rate int)
	SetBurst(burst int)
}

var script = `local key           = KEYS[1] --键名，格式hash
local volumn        = tonumber(KEYS[2]) --桶容量
local rate          = tonumber(KEYS[3]) --桶流出速率
local request_count = tonumber(KEYS[4]) --请求的增长数
local current_time  = tonumber(KEYS[5]) --当前时间戳
local ttl = math.floor((volumn/rate)*2) --有效期

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
 
redis.call('hset', key, 'volumn', volumn)
redis.call('hset', key, 'rate', rate)
redis.call('hset', key, 'last_time', current_time)
redis.call("expire", key, ttl)
 
if remain_count > volumn then
    redis.call('hset', key, 'count', volumn)
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
	return &option{
		clock: clock.New(),
	}
}

type leakyBucket struct {
	opts   *option
	mu     sync.Mutex
	volumn int
	rate   int
	client *redis.Client
}

func NewLeakyBucket(rate, volume int, client *redis.Client, opts ...Option) LeakyBucket {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	l := &leakyBucket{
		opts:   opt,
		client: client,
	}

	l.SetLimit(rate)
	l.SetBurst(volume)

	return l
}

func (lb *leakyBucket) Allow(ctx context.Context, key string, requestCount int) (ok bool, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	res, err := lb.client.Do(ctx, "EVAL", script, 5, key, lb.volumn, lb.rate, requestCount, lb.opts.clock.Now().Unix()).Result()
	if err != nil {
		return
	}

	if fmt.Sprint(res) == "1" {
		return true, nil
	}

	return
}

func (lb *leakyBucket) SetLimit(rate int) {
	lb.mu.Lock()
	lb.rate = rate
	lb.mu.Unlock()
}

func (lb *leakyBucket) SetBurst(burst int) {
	lb.mu.Lock()
	lb.volumn = burst
	lb.mu.Unlock()
}
