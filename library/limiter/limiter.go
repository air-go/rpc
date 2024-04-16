package limiter

import (
	"context"
)

type Limiter interface {
	Allow(ctx context.Context, key string, opts ...AllowOptionFunc) (bool, error)
}

type ParallelLimiter interface {
	Allow(ctx context.Context, key string, opts ...AllowOptionFunc) (bool, error)
	Finish(ctx context.Context, key string)
}

type AllowOptions struct {
	Count int
}

type AllowOptionFunc func(*AllowOptions)

func OptionCount(count int) AllowOptionFunc {
	return func(opts *AllowOptions) { opts.Count = count }
}
