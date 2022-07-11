package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/why444216978/go-util/assert"
	utilCtx "github.com/why444216978/go-util/context"
	"github.com/why444216978/go-util/snowflake"

	"github.com/air-go/rpc/library/cache"
	"github.com/air-go/rpc/library/lock"
)

type RedisCache struct {
	*Option
	c    redis.Cmdable
	lock lock.Locker
}

type Option struct {
	try int
}

type OptionFunc func(*Option)

func WithTry(try int) OptionFunc {
	return func(o *Option) { o.try = try }
}

func defaultOption() *Option {
	return &Option{try: 3}
}

var _ cache.Cacher = (*RedisCache)(nil)

func New(c redis.Cmdable, locker lock.Locker, opts ...OptionFunc) (*RedisCache, error) {
	if assert.IsNil(c) {
		return nil, errors.New("redis is nil")
	}

	if assert.IsNil(locker) {
		return nil, errors.New("locker is nil")
	}

	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	return &RedisCache{
		Option: opt,
		c:      c,
		lock:   locker,
	}, nil
}

func (rc *RedisCache) GetData(ctx context.Context, key string, ttl time.Duration, virtualTTL time.Duration, f cache.LoadFunc, data interface{}) (err error) {
	cache, err := rc.getCache(ctx, key)
	if err != nil {
		return
	}

	// cache no-exists
	if cache.ExpireAt == 0 || cache.Data == "" {
		err = rc.FlushCache(ctx, key, ttl, virtualTTL, f, data)
		return
	}

	// cache expiration
	err = json.Unmarshal([]byte(cache.Data), data)
	if err != nil {
		return
	}
	if time.Now().Before(time.Unix(cache.ExpireAt, 0)) {
		return
	}

	ctxNew, _ := context.WithTimeout(utilCtx.RemoveCancel(ctx), time.Second*10)
	go rc.FlushCache(ctxNew, key, ttl, virtualTTL, f, data)

	return
}

func (rc *RedisCache) FlushCache(ctx context.Context, key string, ttl time.Duration, virtualTTL time.Duration, f cache.LoadFunc, data interface{}) (err error) {
	lockKey := "LOCK::" + key
	random := snowflake.Generate().String()

	ok, err := rc.lock.Lock(ctx, lockKey, random, time.Second*10, rc.try)
	if err != nil || !ok {
		return
	}
	defer rc.lock.Unlock(ctx, lockKey, random)

	// load data
	err = cache.HandleLoad(ctx, f, data)
	if err != nil {
		return
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		return
	}

	// save cache
	err = rc.setCache(ctx, key, string(dataStr), ttl, virtualTTL)

	return
}

func (rc *RedisCache) getCache(ctx context.Context, key string) (data *cache.CacheData, err error) {
	data = &cache.CacheData{}

	res, err := rc.c.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return data, nil
	}
	if err != nil {
		return
	}

	if res == "" {
		return
	}

	err = json.Unmarshal([]byte(res), data)
	if err != nil {
		return
	}

	return
}

func (rc *RedisCache) setCache(ctx context.Context, key, val string, ttl time.Duration, virtualTTL time.Duration) (err error) {
	_data := cache.CacheData{
		ExpireAt: time.Now().Add(virtualTTL).Unix(),
		Data:     val,
	}
	data, err := json.Marshal(_data)
	if err != nil {
		return
	}

	_, err = rc.c.Set(ctx, key, string(data), ttl).Result()
	if err != nil {
		return
	}

	return
}
