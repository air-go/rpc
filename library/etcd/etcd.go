package etcd

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Config struct {
	Endpoints   string
	DialTimeout time.Duration
}

type Option struct {
	dialTimeout time.Duration
}

// Etcd
type Etcd struct {
	*clientv3.Client
	opts      *Option
	endpoints []string
}

type OptionFunc func(*Option)

func WithDialTimeout(duration time.Duration) OptionFunc {
	return func(o *Option) { o.dialTimeout = duration * time.Second }
}

func defaultOption() *Option {
	return &Option{dialTimeout: time.Second * 10}
}

// NewClient
func NewClient(endpoints []string, opts ...OptionFunc) (*Etcd, error) {
	var err error

	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	cli := &Etcd{
		opts:      opt,
		endpoints: endpoints,
	}
	cli.Client, err = clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: opt.dialTimeout,
	})
	if err != nil {
		return nil, err
	}

	return cli, nil
}
