package redisbloom

import (
	"context"
	"errors"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/go-redis/redis/v8"
	"github.com/spaolacci/murmur3"
	"github.com/why444216978/go-util/assert"

	lbloom "github.com/air-go/rpc/library/bloom"
)

const (
	setScript = `
local key = KEYS[1]
local ttl = tonumber(KEYS[2])
if ttl < 60 then
	ttl = 60
end
redis.call("expire", key, ttl)
for _, offset in pairs(ARGV) do
	redis.call("setbit", key, offset, 1)
end
`
	checkScript = `
local key = KEYS[1]
local ttl = tonumber(KEYS[2])
if ttl < 60 then
	ttl = 60
end
redis.call("expire", key, ttl)
for _, offset in pairs(ARGV) do
	if tonumber(redis.call("getbit", key, offset)) == 0 then
		return "0"
	end
end
return "1"
`
	checkAndSetScript = `
local key      = KEYS[1]
local ttl      = tonumber(KEYS[2])
if ttl < 60 then
	ttl = 60
end
local isExists = true
redis.call("expire", key, ttl)
for _, offset in pairs(ARGV) do
	if tonumber(redis.call("getbit", key, offset)) == 0 then
		isExists = false
	end 
end
if isExists then
	return "1"
end
for _, offset in pairs(ARGV) do
	redis.call("setbit", key, offset, 1)
end
return "0"
`
)

type options struct {
	n uint
	p float64
}
type OptionFunc func(*options)

func defaultOptions() *options {
	return &options{
		n: 10000,
		p: 0.01,
	}
}

func SetEstimateParameters(n uint, p float64) OptionFunc {
	return func(o *options) {
		o.n = n
		o.p = p
	}
}

type RedisBloom struct {
	*options
	m, k     uint64
	getRedis func() *redis.Client
}

var _ lbloom.Bloom = (*RedisBloom)(nil)

func NewRedisBloom(getRedis func() *redis.Client, opts ...OptionFunc) (*RedisBloom, error) {
	if assert.IsNil(getRedis) {
		return nil, errors.New("getRedis is nil")
	}

	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	m, k := bloom.EstimateParameters(opt.n, opt.p)

	return &RedisBloom{
		options:  opt,
		m:        uint64(m),
		k:        uint64(k),
		getRedis: getRedis,
	}, nil
}

func (rb *RedisBloom) Add(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	locations := rb.getLocations(data)
	return rb.set(ctx, key, locations, ttl)
}

func (rb *RedisBloom) Check(ctx context.Context, key string, data []byte, ttl time.Duration) (bool, error) {
	locations := rb.getLocations(data)
	isExists, err := rb.check(ctx, key, locations, ttl)
	if err != nil {
		return false, err
	}

	return isExists, nil
}

// CheckAndAdd
// If exists return false, otherwise return true.
func (rb *RedisBloom) CheckAndAdd(ctx context.Context, key string, data []byte, ttl time.Duration) (bool, error) {
	locations := rb.getLocations(data)
	setSuccess, err := rb.checkAndAdd(ctx, key, locations, ttl)
	if err != nil {
		return false, err
	}

	return setSuccess, nil
}

func (rb *RedisBloom) checkAndAdd(ctx context.Context, key string, offsets []uint64, ttl time.Duration) (bool, error) {
	resp, err := rb.getRedis().Do(ctx, rb.buildArgs(key, checkAndSetScript, offsets, ttl)...).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return resp == "1", nil
}

func (rb *RedisBloom) check(ctx context.Context, key string, offsets []uint64, ttl time.Duration) (bool, error) {
	resp, err := rb.getRedis().Do(ctx, rb.buildArgs(key, checkScript, offsets, ttl)...).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return resp == "1", nil
}

func (rb *RedisBloom) set(ctx context.Context, key string, offsets []uint64, ttl time.Duration) error {
	_, err := rb.getRedis().Do(ctx, rb.buildArgs(key, setScript, offsets, ttl)...).Result()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	return err
}

func (rb *RedisBloom) buildArgs(key, script string, offsets []uint64, ttl time.Duration) []interface{} {
	args := []interface{}{
		"eval",
		script,
		2,
		key,
		int64(ttl / time.Second),
	}
	for _, offset := range offsets {
		args = append(args, offset)
	}
	return args
}

func (rb *RedisBloom) getLocations(data []byte) []uint64 {
	h := rb.baseHashes(data)
	locations := []uint64{}
	for i := uint64(0); i < rb.k; i++ {
		locations = append(locations, rb.location(h, i))
	}
	return locations
}

// baseHashes reference:https://github.com/bits-and-blooms/bloom/blob/master/bloom.go#L111
func (rb *RedisBloom) baseHashes(data []byte) [4]uint64 {
	a1 := []byte{1}
	hasher := murmur3.New128()
	hasher.Write(data)
	v1, v2 := hasher.Sum128()
	hasher.Write(a1)
	v3, v4 := hasher.Sum128()
	return [4]uint64{
		v1, v2, v3, v4,
	}
}

// location reference:https://github.com/bits-and-blooms/bloom/blob/master/bloom.go#L126
func (rb *RedisBloom) location(h [4]uint64, i uint64) uint64 {
	l := location(h, i)
	return uint64(l % rb.m)
}

func location(h [4]uint64, i uint64) uint64 {
	return h[i%2] + i*h[2+(((i+(i%2))%4)/2)]
}
