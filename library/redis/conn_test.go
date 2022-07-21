package redis

import (
	"context"
	"testing"

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

func TestRedisClient_Dialer(t *testing.T) {}

func TestRedisClient_OnConnect(t *testing.T) {
	_ = (&RedisClient{}).OnConnect(context.Background(), &redis.Conn{})
}
