package slidingwindow

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/errgroup"
)

type RateLimiter interface {
	ShouldAllow(ctx context.Context, key string, size time.Duration, limit int64) (allow bool, err error)
}

// DefaultLimiter implements Sliding-Windows-Log rate limiter algorithm
type DefaultLimiter struct {
	RedisClient redis.Cmdable
}

func NewDefaultLimiter(redisClient redis.Cmdable) RateLimiter {
	return &DefaultLimiter{
		RedisClient: redisClient,
	}
}

// ShouldAllow returns if current request is allowed or not.
func (l *DefaultLimiter) ShouldAllow(ctx context.Context, key string, size time.Duration, limit int64) (allow bool, err error) {
	var count int64

	nowTime := time.Now()
	periodBoundary := nowTime.Add(-size)

	count, err = l.RedisClient.ZCount(ctx, key, strconv.FormatInt(periodBoundary.UnixMicro(), 10), strconv.FormatInt(nowTime.UnixMicro(), 10)).Result()
	if err != nil {
		return
	}

	if count >= limit {
		return
	}

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() (err error) {
		err = l.RedisClient.ZAdd(ctx, key, &redis.Z{Score: float64(nowTime.UnixMicro()), Member: nowTime.UnixMicro()}).Err()
		if err != nil {
			return
		}
		return
	})

	g.Go(func() (err error) {
		err = l.RedisClient.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(periodBoundary.UnixMicro()-1, 10)).Err()
		if err != nil {
			return
		}
		return
	})

	if err = g.Wait(); err != nil {
		return
	}

	allow = true

	return
}
