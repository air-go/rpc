package redis

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/air-go/rpc/library/logger"
	"github.com/go-redis/redis/v8"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisLoggerWithWriter(t *testing.T) {
	ctx := logger.InitFieldsContainer(context.Background())

	convey.Convey("TestNewRedisLoggerWithWriter", t, func() {
		convey.Convey("injected writer receives command log with config host", func() {
			dir := t.TempDir()
			infoFile := filepath.Join(dir, "info.log")
			errFile := filepath.Join(dir, "error.log")
			buf := &bytes.Buffer{}

			l, err := NewRedisLogger(&RedisConfig{
				Level: "info", ServiceName: "test", Host: "127.0.0.1", Port: 6379,
				InfoFile: infoFile, ErrorFile: errFile,
			}, WithWriter(buf, buf))
			assert.Nil(t, err)

			err = l.AfterProcess(ctx, redis.NewBoolCmd(ctx))
			assert.Nil(t, err)

			assert.Contains(t, buf.String(), "127.0.0.1")
			assert.Contains(t, buf.String(), "6379")
			assert.NoFileExists(t, infoFile)
			assert.NoFileExists(t, errFile)
		})
	})
}

func TestNewRedisLoggerRejectsSingleWriter(t *testing.T) {
	l, err := NewRedisLogger(&RedisConfig{}, WithWriter(&bytes.Buffer{}, nil))

	assert.NotNil(t, err)
	assert.Nil(t, l)
}

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
			ctx := logger.InitFieldsContainer(context.TODO())
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
			ctx := logger.InitFieldsContainer(context.TODO())
			l := &RedisLogger{}
			err := l.AfterProcess(ctx, redis.NewBoolCmd(ctx))
			assert.Nil(t, err)
		})
		convey.Convey("cmd error", func() {
			l, err := NewRedisLogger(&RedisConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			ctx := logger.InitFieldsContainer(context.TODO())
			cmd := redis.NewBoolCmd(ctx)
			cmd.SetErr(errors.New("error"))

			err = l.AfterProcess(ctx, cmd)
			assert.Nil(t, err)
		})
		convey.Convey("info", func() {
			l, err := NewRedisLogger(&RedisConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			ctx := logger.InitFieldsContainer(context.TODO())
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
			ctx := logger.InitFieldsContainer(context.TODO())
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
			ctx := logger.InitFieldsContainer(context.TODO())
			l := &RedisLogger{}
			err := l.AfterProcessPipeline(ctx, []redis.Cmder{redis.NewBoolCmd(ctx)})
			assert.Nil(t, err)
		})
		convey.Convey("cmd error", func() {
			l, err := NewRedisLogger(&RedisConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			ctx := logger.InitFieldsContainer(context.TODO())
			cmd := redis.NewBoolCmd(ctx)
			cmd.SetErr(errors.New("error"))

			err = l.AfterProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Nil(t, err)
		})
		convey.Convey("info", func() {
			l, err := NewRedisLogger(&RedisConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			ctx := logger.InitFieldsContainer(context.TODO())
			cmd := redis.NewBoolCmd(ctx)

			err = l.AfterProcessPipeline(ctx, []redis.Cmder{cmd})
			assert.Nil(t, err)
		})
	})
}
