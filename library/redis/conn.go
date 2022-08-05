package redis

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	// "github.com/air-go/rpc/library/servicer"
)

type Config struct {
	ServiceName  string
	Host         string
	Port         int
	Username     string
	Password     string
	DB           int
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
	MaxConnAge   int
	PoolSize     int
	MinIdleConns int
}

type Option struct{}

type OptionFunc func(*Option)

type RedisClient struct {
	*redis.Client
	opts        *Option
	config      *Config
	serviceName string
}

func NewRedisClient(cfg *Config, opts ...OptionFunc) (*RedisClient, error) {
	options := &Option{}
	for _, o := range opts {
		o(options)
	}

	cli := &RedisClient{
		opts:        options,
		config:      cfg,
		serviceName: cfg.ServiceName,
	}

	cli.Client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Dialer:       cli.Dialer,
		OnConnect:    cli.OnConnect,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Millisecond,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Millisecond,
		MaxConnAge:   time.Duration(cfg.MaxConnAge) * time.Millisecond,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	return cli, nil
}

func (rc *RedisClient) Dialer(ctx context.Context, network, addr string) (net.Conn, error) {
	// TODO user servicer discover addr
	// s, ok := servicer.GetServicer(rc.serviceName)
	// if !ok {
	// 	return nil, errors.Errorf("NewRedisClient GetServicer %s not found", rc.serviceName)
	// }

	// node, err := s.Pick(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// addr := fmt.Sprintf("%s:%d", node.Host(), node.Port())

	netDialer := &net.Dialer{
		Timeout:   rc.Options().DialTimeout,
		KeepAlive: 5 * time.Minute,
	}
	return netDialer.DialContext(ctx, network, addr)
}

func (rc *RedisClient) OnConnect(ctx context.Context, cn *redis.Conn) (err error) {
	return
}
