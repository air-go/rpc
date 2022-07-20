package http

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	libraryOpentracing "github.com/air-go/rpc/library/opentracing"
)

func TestExtractHTTP(t *testing.T) {
	ctx := context.Background()
	req := &http.Request{
		Header: http.Header{},
		URL:    &url.URL{},
	}

	convey.Convey("TestExtractHTTP", t, func() {
		convey.Convey("Tracer nil", func() {
			libraryOpentracing.Tracer = nil

			_, span, spanID := ExtractHTTP(ctx, req)
			assert.Equal(t, span, nil)
			assert.Equal(t, spanID, "")
		})
		convey.Convey("success no parentSpanContext", func() {
			tracer := mocktracer.New()
			libraryOpentracing.Tracer = tracer

			ctx, span, _ := ExtractHTTP(ctx, req)
			span.Finish()

			_, ok := span.Context().(mocktracer.MockSpanContext)

			assert.Equal(t, ok, true)
			assert.Equal(t, span, opentracing.SpanFromContext(ctx))
			assert.Len(t, tracer.FinishedSpans(), 1)
		})
		convey.Convey("success has parentSpanContext", func() {
			tracer := mocktracer.New()
			libraryOpentracing.Tracer = tracer

			span := tracer.StartSpan(httpServerComponentPrefix + req.URL.Path)
			span.Finish()
			_ = tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

			ctx, span, _ = ExtractHTTP(ctx, req)
			span.Finish()

			_, ok := span.Context().(mocktracer.MockSpanContext)

			assert.Equal(t, ok, true)
			assert.Equal(t, span, opentracing.SpanFromContext(ctx))
			assert.Len(t, tracer.FinishedSpans(), 2)
		})
	})
}

func TestInjectHTTP(t *testing.T) {
	ctx := context.Background()
	req := &http.Request{
		Header: http.Header{},
		URL:    &url.URL{},
	}
	logID := "logID"

	convey.Convey("TestInjectHTTP", t, func() {
		convey.Convey("Tracer nil", func() {
			libraryOpentracing.Tracer = nil

			err := InjectHTTP(ctx, req, logID)
			assert.Equal(t, err, nil)
		})
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			libraryOpentracing.Tracer = tracer

			span := tracer.StartSpan(httpServerComponentPrefix + req.URL.Path)
			span.Finish()
			_ = tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

			err := InjectHTTP(ctx, req, logID)
			assert.Equal(t, err, nil)
			assert.Len(t, tracer.FinishedSpans(), 2)
		})
	})
}

func TestSetHTTPLog(t *testing.T) {
	convey.Convey("TestSetHTTPLog", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan(httpServerComponentPrefix + "uri")
			span.Finish()
			SetHTTPLog(span, "logID", "req", "resp")
		})
	})
}
