package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisLogger(t *testing.T) {
	convey.Convey("TestNewRedisLogger", t, func() {
		convey.Convey("success", func() {
			l, err := NewRedisLogger(&RedisConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)
		})
	})
}

func TestRedisLogger_BeforeProcess(t *testing.T) {
	l, err := NewRedisLogger(&RedisConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, l)

	convey.Convey("TestRedisLogger_BeforeProcess", t, func() {
		convey.Convey("success", func() {
			ctx := context.TODO()
			ctx, err := l.BeforeProcess(ctx, redis.NewBoolCmd(ctx))

			assert.Nil(t, err)

			start := ctx.Value(cmdStart)
			assert.NotNil(t, start)
		})
	})
}

func TestRedisLogger_AfterProcess(t *testing.T) {
	convey.Convey("TestRedisLogger_AfterProcess", t, func() {
		convey.Convey("Logger nil", func() {
			ctx := context.TODO()
			l := &RedisLogger{}
			err := l.AfterProcess(ctx, redis.NewBoolCmd(ctx))
			assert.Nil(t, err)
		})
		convey.Convey("cmd error", func() {
			l, err := NewRedisLogger(&RedisConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			ctx := context.TODO()
			cmd := redis.NewBoolCmd(ctx)
			cmd.SetErr(errors.New("error"))

			err = l.AfterProcess(ctx, cmd)
			assert.Nil(t, err)
		})
		convey.Convey("info", func() {
			l, err := NewRedisLogger(&RedisConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			ctx := context.TODO()
			cmd := redis.NewBoolCmd(ctx)

			err = l.AfterProcess(ctx, cmd)
			assert.Nil(t, err)
		})
	})
}

func TestRedisLogger_BeforeProcessPipeline(t *testing.T) {
	l, err := NewRedisLogger(&RedisConfig{})
	assert.Nil(t, err)
	assert.NotNil(t, l)

	convey.Convey("TestRedisLogger_BeforeProcessPipeline", t, func() {
		convey.Convey("success", func() {
			ctx := context.TODO()
			ctx, err := l.BeforeProcessPipeline(ctx, []redis.Cmder{redis.NewBoolCmd(ctx)})

			assert.Nil(t, err)

			start := ctx.Value(cmdStart)
			assert.NotNil(t, start)
		})
	})
}

func TestRedisLogger_AfterProcessPipeline(t *testing.T) {
	convey.Convey("TestRedisLogger_AfterProcessPipeline", t, func() {
		convey.Convey("Logger nil", func() {
			ctx := context.TODO()
			l := &RedisLogger{}
			err := l.AfterProcessPipeline(ctx, []redis.Cmder{redis.NewBoolCmd(ctx)})
			assert.Nil(t, err)
		})
		convey.Convey("cmd error", func() {
			l, err := NewRedisLogger(&RedisConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			ctx := context.TODO()
			cmd := redis.NewBoolCmd(ctx)
			cmd.SetErr(errors.New("error"))

			err = l.AfterProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Nil(t, err)
		})
		convey.Convey("info", func() {
			l, err := NewRedisLogger(&RedisConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			ctx := context.TODO()
			cmd := redis.NewBoolCmd(ctx)

			err = l.AfterProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Nil(t, err)
		})
	})
}
