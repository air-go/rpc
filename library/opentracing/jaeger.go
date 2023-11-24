package opentracing

import (
	"context"
	"io"

	"github.com/opentracing/opentracing-go"
	opentracingLog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	lc "github.com/air-go/rpc/library/context"
)

const (
	tagLogID   = "Log-Id"
	tagTraceID = "Trace-Id"
	tagSpanID  = "Span-Id"
)

const (
	logFieldsRequest  = "request"
	logFieldsResponse = "response"
	logFieldsArgs     = "args"
)

var Tracer opentracing.Tracer

type Config struct {
	Host string
	Port string
}

func NewJaegerTracer(connCfg *Config, serviceName string) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const", // 固定采样
			Param: 1,       // 1=全采样、0=不采样
		},

		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: connCfg.Host + ":" + connCfg.Port,
		},

		ServiceName: serviceName,
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	Tracer = tracer
	return tracer, closer, nil
}

func SetBasicTags(ctx context.Context, span opentracing.Span) {
	SetTraceTag(ctx, span)
	SetLogID(ctx, span)
}

func SetLogID(ctx context.Context, span opentracing.Span) {
	span.SetTag(tagLogID, lc.ValueLogID(ctx))
}

func SetTraceTag(ctx context.Context, span opentracing.Span) {
	jaegerSpanContext := spanContextToJaegerContext(span.Context())
	span.SetTag(tagTraceID, jaegerSpanContext.TraceID().String())
	span.SetTag(tagSpanID, jaegerSpanContext.SpanID().String())
}

func SetRequest(span opentracing.Span, request interface{}) {
	span.LogFields(opentracingLog.Object(logFieldsRequest, request))
}

func SetResponse(span opentracing.Span, response interface{}) {
	span.LogFields(opentracingLog.Object(logFieldsResponse, response))
}

func GetTraceID(span opentracing.Span) string {
	jaegerSpanContext := spanContextToJaegerContext(span.Context())
	return jaegerSpanContext.TraceID().String()
}

func GetSpanID(span opentracing.Span) string {
	jaegerSpanContext := spanContextToJaegerContext(span.Context())
	return jaegerSpanContext.SpanID().String()
}

func spanContextToJaegerContext(spanContext opentracing.SpanContext) jaeger.SpanContext {
	if sc, ok := spanContext.(jaeger.SpanContext); ok {
		return sc
	}

	return jaeger.SpanContext{}
}
