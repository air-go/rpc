package redis

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracerLog "github.com/opentracing/opentracing-go/log"

	libraryOpentracing "github.com/air-go/rpc/library/opentracing"
	"github.com/why444216978/go-util/assert"
)

const (
	operationRedis = "Redis-"
	logCmdName     = "command"
	logCmdArgs     = "args"
	logCmdResult   = "result"
)

type contextKey int

const (
	cmdStart contextKey = iota
)

// jaegerHook is go-redis jaeger hook
type jaegerHook struct{}

// NewJaegerHook return jaegerHook
func NewJaegerHook() redis.Hook {
	return &jaegerHook{}
}

// BeforeProcess redis before execute action do something
func (jh *jaegerHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if assert.IsNil(libraryOpentracing.Tracer) {
		return ctx, nil
	}
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, libraryOpentracing.Tracer, operationRedis+cmd.Name())

	libraryOpentracing.SetCommonTag(ctx, span)

	ctx = opentracing.ContextWithSpan(ctx, span)
	return ctx, nil
}

// AfterProcess redis after execute action do something
func (jh *jaegerHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if assert.IsNil(libraryOpentracing.Tracer) {
		return nil
	}
	span := opentracing.SpanFromContext(ctx)
	if assert.IsNil(span) {
		return nil
	}
	defer span.Finish()

	span.LogFields(tracerLog.String(logCmdName, cmd.Name()))
	span.LogFields(tracerLog.Object(logCmdArgs, cmd.Args()))
	span.LogFields(tracerLog.Object(logCmdResult, cmd.String()))

	if err := cmd.Err(); isRedisError(err) {
		span.LogFields(tracerLog.Error(err))
		span.SetTag(string(ext.Error), true)
	}

	return nil
}

// BeforeProcessPipeline before command process handle
func (jh *jaegerHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	if assert.IsNil(libraryOpentracing.Tracer) {
		return ctx, nil
	}

	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, libraryOpentracing.Tracer, operationRedis+"pipeline")

	ctx = context.WithValue(ctx, cmdStart, span)

	return ctx, nil
}

// AfterProcessPipeline after command process handle
func (jh *jaegerHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if assert.IsNil(libraryOpentracing.Tracer) {
		return nil
	}

	span, ok := ctx.Value(cmdStart).(opentracing.Span)
	if !ok {
		return nil
	}
	defer span.Finish()

	hasErr := false
	for idx, cmd := range cmds {
		if err := cmd.Err(); isRedisError(err) {
			hasErr = true
		}
		span.LogFields(tracerLog.String(jh.getPipeLineLogKey(logCmdName, idx), cmd.Name()))
		span.LogFields(tracerLog.Object(jh.getPipeLineLogKey(logCmdArgs, idx), cmd.Args()))
		span.LogFields(tracerLog.String(jh.getPipeLineLogKey(logCmdResult, idx), cmd.String()))
	}
	if hasErr {
		span.SetTag(string(ext.Error), true)
	}

	return nil
}

func (jh *jaegerHook) getPipeLineLogKey(logField string, idx int) string {
	return logField + "-" + strconv.Itoa(idx)
}

func isRedisError(err error) bool {
	if err == redis.Nil {
		return false
	}
	_, ok := err.(redis.Error)
	return ok
}
