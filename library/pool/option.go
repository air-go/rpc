package pool

import (
	"time"
)

type Options struct {
	PoolSize         uint32
	IdleSize         uint32
	BreakerThreshold uint32 // 连接失败熔断阈值
	GetQuickFail     bool
	GetConnTimeout   time.Duration
}

func defaultOptions() *Options {
	opt := &Options{
		PoolSize:         10,
		IdleSize:         5,
		BreakerThreshold: 1,
		GetQuickFail:     false,
		GetConnTimeout:   time.Second * 10,
	}
	return opt
}

type Option func(*Options)

func WithPoolSize(s uint32) Option {
	return func(o *Options) { o.PoolSize = s }
}

func WithIdleSize(s uint32) Option {
	return func(o *Options) { o.IdleSize = s }
}

func WithBreakerThreshold(s uint32) Option {
	return func(o *Options) { o.BreakerThreshold = s }
}

func WithGetQuickFail() Option {
	return func(o *Options) { o.GetQuickFail = true }
}

func WithGetConnTimeout(t time.Duration) Option {
	return func(o *Options) { o.GetConnTimeout = t }
}
