package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	// "github.com/air-go/rpc/library/servicer"
)

type Config struct {
	ServiceName string

	// The network type, either tcp or unix. Default is tcp.
	Network string
	Host    string
	Port    int

	// Dialer creates new network connection and has priority over Network and Addr options.
	// Default is the dialer of RedisClient.
	Dialer func(ctx context.Context, network, addr string) (net.Conn, error)

	// Hook that is called when new connection is established.
	// Default is the hook of RedisClient.
	OnConnect func(ctx context.Context, cn *redis.Conn) error

	Username string
	Password string
	DB       int

	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int
	// Minimum backoff between each retry, millisecond.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff int
	// Maximum backoff between each retry, millisecond.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff int

	// Dial timeout for establishing new connections, millisecond.
	DialTimeout int
	// Timeout for socket reads, millisecond.
	// Use value -1 for no timeout and 0 for default.
	ReadTimeout int
	// Timeout for socket writes, millisecond.
	WriteTimeout int

	// Type of connection pool.
	// true for FIFO pool, false for LIFO pool.
	PoolFIFO bool
	// Maximum number of socket connections.
	PoolSize int
	// Minimum number of idle connections.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection, millisecond.
	MaxConnAge int
	// Amount of time client waits for connection if all connections are busy, millisecond.
	PoolTimeout int
	// Amount of time after which client closes idle connections, millisecond.
	// -1 disables idle timeout check.
	IdleTimeout int
	// Frequency of idle checks made by idle connections reaper, millisecond.
	// -1 disables idle connections reaper.
	IdleCheckFrequency int

	// TLS Config to use. When set TLS will be negotiated.
	TLSConfig *tls.Config

	// Limiter interface used to implemented circuit breaker or rate limiter.
	Limiter redis.Limiter
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

	dialer := cfg.Dialer
	if dialer == nil {
		dialer = cli.Dialer
	}

	onConnect := cfg.OnConnect
	if onConnect == nil {
		onConnect = cli.OnConnect
	}

	cli.Client = redis.NewClient(&redis.Options{
		Network:            cfg.Network,
		Addr:               fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Dialer:             dialer,
		OnConnect:          onConnect,
		Username:           cfg.Username,
		Password:           cfg.Password,
		DB:                 cfg.DB,
		MaxRetries:         cfg.MaxRetries,
		MinRetryBackoff:    milliseconds(cfg.MinRetryBackoff),
		MaxRetryBackoff:    milliseconds(cfg.MaxRetryBackoff),
		DialTimeout:        milliseconds(cfg.DialTimeout),
		ReadTimeout:        milliseconds(cfg.ReadTimeout),
		WriteTimeout:       milliseconds(cfg.WriteTimeout),
		PoolFIFO:           cfg.PoolFIFO,
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxConnAge:         milliseconds(cfg.MaxConnAge),
		PoolTimeout:        milliseconds(cfg.PoolTimeout),
		IdleTimeout:        milliseconds(cfg.IdleTimeout),
		IdleCheckFrequency: milliseconds(cfg.IdleCheckFrequency),
		TLSConfig:          cfg.TLSConfig,
		Limiter:            cfg.Limiter,
	})

	return cli, nil
}

// milliseconds converts config value to duration,
// -1 is kept as is because go-redis uses it to disable the related feature.
func milliseconds(v int) time.Duration {
	if v == -1 {
		return -1
	}
	return time.Duration(v) * time.Millisecond
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

	// Don't read rc.Options() here: when MinIdleConns > 0, redis.NewClient warms up
	// the pool in background goroutines before rc.Client is assigned, so the embedded
	// client is still nil. Read the timeout from config instead.
	netDialer := &net.Dialer{
		Timeout:   rc.dialTimeout(),
		KeepAlive: 5 * time.Minute,
	}
	if rc.config == nil || rc.config.TLSConfig == nil {
		return netDialer.DialContext(ctx, network, addr)
	}
	return tls.DialWithDialer(netDialer, network, addr, rc.config.TLSConfig)
}

func (rc *RedisClient) OnConnect(ctx context.Context, cn *redis.Conn) (err error) {
	return
}

// dialTimeout keeps the same semantics as go-redis Options.DialTimeout:
// 0 means the 5s default, -1 means no timeout.
func (rc *RedisClient) dialTimeout() time.Duration {
	if rc.config == nil || rc.config.DialTimeout == 0 {
		return 5 * time.Second
	}
	if rc.config.DialTimeout == -1 {
		return 0
	}
	return milliseconds(rc.config.DialTimeout)
}
