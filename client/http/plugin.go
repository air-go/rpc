package http

import (
	"context"
	"net/http"

	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/air-go/rpc/library/logger"
	jaeger "github.com/air-go/rpc/library/opentracing/http"
	libraryOtel "github.com/air-go/rpc/library/otel"
	"github.com/air-go/rpc/server/http/middleware/timeout"
)

type BeforeRequestPlugin interface {
	Handle(ctx context.Context, req *http.Request) (context.Context, error)
	Name() string
}

type AfterRequestPlugin interface {
	Handle(ctx context.Context, req *http.Request, resp *http.Response) (context.Context, error)
	Name() string
}

type OpentracingBeforePlugin struct{}

var _ BeforeRequestPlugin = (*OpentracingBeforePlugin)(nil)

func (*OpentracingBeforePlugin) Handle(ctx context.Context, req *http.Request) (context.Context, error) {
	return ctx, jaeger.InjectHTTP(ctx, req)
}

func (*OpentracingBeforePlugin) Name() string {
	return "OpentracingBeforePlugin"
}

type OpentelemetryBeforePlugin struct{}

var _ BeforeRequestPlugin = (*OpentelemetryBeforePlugin)(nil)

func (*OpentelemetryBeforePlugin) Handle(ctx context.Context, req *http.Request) (context.Context, error) {
	if !libraryOtel.CheckHasTraceID(ctx) {
		return ctx, nil
	}

	logID := logger.ValueLogID(ctx)
	req.Header.Set(logger.LogHeader, logID)

	libraryOtel.InjectHTTPBaggage(ctx, req.Header)

	path := req.URL.Path
	opts := []trace.SpanStartOption{
		trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", req)...),
		trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(req)...),
		trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest("rpc-example", path, req)...),
		trace.WithSpanKind(trace.SpanKindClient),
	}

	ctx, span := libraryOtel.Tracer().Start(ctx, path, opts...)
	defer span.End()

	return ctx, nil
}

func (*OpentelemetryBeforePlugin) Name() string {
	return "OpentelemetryBeforePlugin"
}

type LogBeforePlugin struct{}

var _ BeforeRequestPlugin = (*LogBeforePlugin)(nil)

func (*LogBeforePlugin) Handle(ctx context.Context, req *http.Request) (context.Context, error) {
	return ctx, logger.SetLogID(ctx, req.Header)
}

func (*LogBeforePlugin) Name() string {
	return "LogBeforePlugin"
}

type TimeoutBeforePlugin struct{}

var _ BeforeRequestPlugin = (*TimeoutBeforePlugin)(nil)

func (*TimeoutBeforePlugin) Handle(ctx context.Context, req *http.Request) (context.Context, error) {
	return ctx, timeout.SetHeader(ctx, req.Header)
}

func (*TimeoutBeforePlugin) Name() string {
	return "TimeoutBeforePlugin"
}
