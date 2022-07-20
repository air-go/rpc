package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/why444216978/go-util/assert"

	libraryOpentracing "github.com/air-go/rpc/library/opentracing"
)

const (
	httpClientComponentPrefix = "HTTP-Client-"
	httpServerComponentPrefix = "HTTP-Server-"
)

var ErrTracerNil = errors.New("Tracer is nil")

// ExtractHTTP is used to extract span context by HTTP middleware
func ExtractHTTP(ctx context.Context, req *http.Request) (context.Context, opentracing.Span, string) {
	if assert.IsNil(libraryOpentracing.Tracer) {
		return ctx, nil, ""
	}

	var span opentracing.Span

	parentSpanContext, err := libraryOpentracing.Tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	if assert.IsNil(parentSpanContext) || err == opentracing.ErrSpanContextNotFound {
		span, ctx = opentracing.StartSpanFromContextWithTracer(ctx, libraryOpentracing.Tracer, httpServerComponentPrefix+req.URL.Path, ext.SpanKindRPCServer)
	} else {
		span = libraryOpentracing.Tracer.StartSpan(
			httpServerComponentPrefix+req.URL.Path,
			ext.RPCServerOption(parentSpanContext),
			ext.SpanKindRPCServer,
		)
	}
	span.SetTag(string(ext.Component), httpServerComponentPrefix+req.URL.Path)

	libraryOpentracing.SetCommonTag(ctx, span)

	ctx = opentracing.ContextWithSpan(ctx, span)

	return ctx, span, libraryOpentracing.GetTraceID(span)
}

// InjectHTTP is used to inject HTTP span
func InjectHTTP(ctx context.Context, req *http.Request, logID string) error {
	if assert.IsNil(libraryOpentracing.Tracer) {
		return nil
	}

	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, libraryOpentracing.Tracer, httpClientComponentPrefix+req.URL.Path, ext.SpanKindRPCClient)
	defer span.Finish()
	span.SetTag(string(ext.Component), httpClientComponentPrefix+req.URL.Path)
	span.SetTag(libraryOpentracing.FieldLogID, logID)
	libraryOpentracing.SetCommonTag(ctx, span)

	return libraryOpentracing.Tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
}

func SetHTTPLog(span opentracing.Span, logID, req, resp string) {
	if assert.IsNil(span) {
		return
	}
	span.SetTag(libraryOpentracing.FieldLogID, logID)
	libraryOpentracing.SetRequest(span, req)
	libraryOpentracing.SetResponse(span, resp)
}
