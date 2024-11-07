package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	libraryOtel "github.com/air-go/rpc/library/otel"
)

// opentelemetryHook is go-redis opentelemetry hook
type opentelemetryHook struct{}

// NewOpentelemetryHook new go-redis opentelemetry hook
func NewOpentelemetryHook() redis.Hook {
	return &opentelemetryHook{}
}

// BeforeProcess redis before execute action do something
func (*opentelemetryHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if !libraryOtel.CheckHasTraceID(ctx) {
		return ctx, nil
	}

	ctx, _ = libraryOtel.Tracer(libraryOtel.TracerNameRedis).Start(ctx, semconv.DBSystemRedis.Value.AsString()+"-"+cmd.Name(), trace.WithSpanKind(trace.SpanKindClient))

	return ctx, nil
}

// AfterProcess redis after execute action do something
func (*opentelemetryHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if !libraryOtel.CheckHasTraceID(ctx) {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("command", trace.WithAttributes([]attribute.KeyValue{
		libraryOtel.AttributeRedisCmdName.String(cmd.Name()),
		libraryOtel.AttributeRedisCmdString.String(cmd.String()),
		libraryOtel.AttributeRedisCmdArgs.String(fmt.Sprintf("%s", cmd.Args())),
	}...))

	if err := cmd.Err(); isRedisError(err) {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(libraryOtel.AttributeRedisError.String(err.Error()))
	}

	return nil
}

// BeforeProcessPipeline before command process handle
func (*opentelemetryHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	if !libraryOtel.CheckHasTraceID(ctx) {
		return ctx, nil
	}

	ctx, _ = libraryOtel.Tracer(libraryOtel.TracerNameRedis).Start(ctx, semconv.DBSystemRedis.Value.AsString()+"-pipline", trace.WithSpanKind(trace.SpanKindClient))

	return ctx, nil
}

// AfterProcessPipeline after command process handle
func (*opentelemetryHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if !libraryOtel.CheckHasTraceID(ctx) {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	defer span.End()

	hasErr := false
	attrs := []attribute.KeyValue{}
	for idx, cmd := range cmds {
		arr := []string{}
		if err := cmd.Err(); isRedisError(err) {
			hasErr = true
			arr = append(arr, "redis.cmd.error - "+err.Error())
		}

		arr = append(arr, []string{
			"redis.cmd.name - " + cmd.Name(),
			"redis.cmd.string - " + cmd.String(),
			"redis.cmd.args - " + fmt.Sprintf("%s", cmd.Args()),
		}...)

		attrs = append(attrs, attribute.StringSlice(fmt.Sprintf("pipline-%d", idx), arr))
	}

	span.AddEvent("command", trace.WithAttributes(attrs...))

	if hasErr {
		span.SetStatus(codes.Error, "pipline error")
	}

	return nil
}

func isRedisError(err error) bool {
	if err == redis.Nil {
		return false
	}
	_, ok := err.(redis.Error)
	return ok
}
