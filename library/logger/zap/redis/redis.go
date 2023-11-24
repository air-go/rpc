package redis

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/air-go/rpc/library/app"
	"github.com/air-go/rpc/library/logger"
	zapLogger "github.com/air-go/rpc/library/logger/zap"
)

type contextKey int

const (
	cmdStart contextKey = iota
)

// RedisConfig is used to parse configuration file
// logger should be controlled with Options
type RedisConfig struct {
	InfoFile    string
	ErrorFile   string
	Level       string
	ServiceName string
	Host        string
	Port        int
}

// RedisLogger is go-redis logger Hook
type RedisLogger struct {
	*zapLogger.ZapLogger
	config *RedisConfig
}

type RedisOption func(rl *RedisLogger)

// NewRedisLogger
func NewRedisLogger(config *RedisConfig, opts ...RedisOption) (rl *RedisLogger, err error) {
	rl = &RedisLogger{config: config}

	for _, o := range opts {
		o(rl)
	}

	infoWriter, errWriter, err := logger.RotateWriter(config.InfoFile, config.ErrorFile)
	if err != nil {
		return
	}

	l, err := zapLogger.NewLogger(
		zapLogger.WithModule(logger.ModuleRedis),
		zapLogger.WithServiceName(config.ServiceName),
		zapLogger.WithCallerSkip(6),
		zapLogger.WithInfoWriter(infoWriter),
		zapLogger.WithErrorWriter(errWriter),
		zapLogger.WithLevel(config.Level),
	)
	if err != nil {
		return
	}
	rl.ZapLogger = l

	return
}

// BeforeProcess redis before execute action do something
func (rl *RedisLogger) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	ctx = rl.setCmdStart(ctx)
	return ctx, nil
}

// AfterProcess redis after execute action do something
func (rl *RedisLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if rl.logger() == nil {
		return nil
	}

	cost := rl.getCmdCost(ctx)
	if err := cmd.Err(); err != nil && err != redis.Nil {
		rl.Error(ctx, false, []redis.Cmder{cmd}, cost)
		return nil
	}

	rl.Info(ctx, false, []redis.Cmder{cmd}, cost)

	return nil
}

// BeforeProcessPipeline before command process handle
func (rl *RedisLogger) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	ctx = rl.setCmdStart(ctx)
	return ctx, nil
}

// AfterProcessPipeline after command process handle
func (rl *RedisLogger) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if rl.logger() == nil {
		return nil
	}
	cost := rl.getCmdCost(ctx)

	hasErr := false
	for _, cmd := range cmds {
		if err := cmd.Err(); err != nil && err != redis.Nil {
			hasErr = true
		}
	}

	if hasErr {
		rl.Error(ctx, true, cmds, cost)
		return nil
	}
	rl.Info(ctx, true, cmds, cost)

	return nil
}

func (rl *RedisLogger) Info(ctx context.Context, isPipeline bool, cmds []redis.Cmder, cost int64) {
	ctx = rl.fields(ctx, isPipeline, cmds, cost)
	rl.logger().Info(ctx, "info")
}

func (rl *RedisLogger) Error(ctx context.Context, isPipeline bool, cmds []redis.Cmder, cost int64) {
	errs := make([]string, 0)
	for idx, cmd := range cmds {
		err := cmd.Err()
		if err == nil {
			return
		}
		errs = append(errs, strconv.Itoa(idx)+"-"+err.Error())
	}

	ctx = rl.fields(ctx, isPipeline, cmds, cost)
	rl.logger().Error(ctx, strings.Join(errs, ","))
}

func (rl *RedisLogger) fields(ctx context.Context, isPipeline bool, cmds []redis.Cmder, cost int64) context.Context {
	ctx = logger.ForkContextOnlyMeta(ctx)

	l := len(cmds)
	names := make([]string, l)
	args := make([]interface{}, l)
	response := make([]string, l)
	for idx, cmd := range cmds {
		names[idx] = cmd.Name()
		args[idx] = cmd.Args()
		response[idx] = cmd.String()
	}

	method := "pipeline"
	if !isPipeline {
		method = cmds[0].Name()
	}

	logger.AddField(ctx,
		logger.Reflect(logger.Method, method),
		logger.Reflect(logger.Request, args),
		logger.Reflect(logger.Response, response),
		logger.Reflect(logger.Code, 0),
		logger.Reflect(logger.ClientIP, app.LocalIP()),
		logger.Reflect(logger.ClientPort, app.Port()),
		logger.Reflect(logger.ServerIP, rl.config.Host),
		logger.Reflect(logger.ServerPort, rl.config.Port),
		logger.Reflect(logger.API, method),
		logger.Reflect(logger.Cost, cost))

	return ctx
}

func (rl *RedisLogger) logger() *zapLogger.ZapLogger {
	return rl.ZapLogger
}

func (rl *RedisLogger) setCmdStart(ctx context.Context) context.Context {
	return context.WithValue(ctx, cmdStart, time.Now())
}

func (rl *RedisLogger) getCmdCost(ctx context.Context) int64 {
	t, ok := ctx.Value(cmdStart).(time.Time)
	if !ok {
		t = time.Now()
	}
	return time.Since(t).Milliseconds()
}
