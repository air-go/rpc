package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/why444216978/go-util/assert"

	"github.com/air-go/rpc/library/lock"
)

const (
	lockSuccess = 1
	lockFail    = 0
	lockLua     = `if redis.call("GET", KEYS[1]) == ARGV[1] then redis.call("DEL", KEYS[1]) return 1 else return 0 end`
)

var _ lock.Locker = (*RedisLock)(nil)

type RedisLock struct {
	c redis.Cmdable
}

func New(c redis.Cmdable) (*RedisLock, error) {
	if assert.IsNil(c) {
		return nil, lock.ErrClientNil
	}
	return &RedisLock{
		c: c,
	}, nil
}

// Lock lock
func (rl *RedisLock) Lock(ctx context.Context, key string, random interface{}, duration time.Duration, try int) (ok bool, err error) {
	count := 0
	for {
		if count >= try {
			return
		}

		count += 1

		ok, err = rl.c.SetNX(ctx, key, random, duration).Result()
		if err != nil {
			return
		}
		if ok {
			return
		}

		wait, err := rl.c.TTL(ctx, key).Result()
		if err != nil {
			continue
		}
		time.Sleep(wait)
	}
}

// UnLock unlock
func (rl *RedisLock) Unlock(ctx context.Context, key string, random interface{}) (err error) {
	res, err := rl.c.Eval(ctx, lockLua, []string{key}, random).Result()
	if err != nil {
		return
	}

	if res == lockFail {
		err = lock.ErrUnLock
		return
	}

	return
}
