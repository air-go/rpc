package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisClient(t *testing.T) {
	convey.Convey("NewRedisClient", t, func() {
		convey.Convey("success", func() {
			cli, err := NewRedisClient(&Config{ServiceName: "default_redis"})
			assert.Nil(t, err)
			assert.NotNil(t, cli)
		})
	})
}

// TestRedisClient_Dialer guards against the data race where redis.NewClient's
// MinIdleConns warm-up goroutines call Dialer before cli.Client is assigned.
func TestRedisClient_Dialer(t *testing.T) {
	for i := 0; i < 100; i++ {
		cli, err := NewRedisClient(&Config{
			ServiceName:  "default_redis",
			Host:         "127.0.0.1",
			Port:         6390, // intentionally unreachable
			MinIdleConns: 10,
			PoolSize:     20,
			DialTimeout:  50,
		})
		assert.Nil(t, err)
		time.Sleep(time.Millisecond)
		_ = cli.Close()
	}
}

func TestRedisClient_OnConnect(t *testing.T) {
	_ = (&RedisClient{}).OnConnect(context.Background(), &redis.Conn{})
}
