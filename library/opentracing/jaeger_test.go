package opentracing

import (
	"context"
	"reflect"
	"testing"

	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	lc "github.com/air-go/rpc/library/context"
)

func TestNewJaegerTracer(t *testing.T) {
	convey.Convey("TestNewJaegerTracer", t, func() {
		convey.Convey("success", func() {
			tracer, closer, err := NewJaegerTracer(&Config{}, "test")
			assert.Nil(t, err)
			assert.NotNil(t, tracer)
			assert.NotNil(t, closer)
		})
	})
}

func TestSetBasicTags(t *testing.T) {
	convey.Convey("TestSetBasicTags", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			ctx := lc.WithLogID(context.Background(), "logid")
			SetBasicTags(ctx, span)

			mockSpan := span.(*mocktracer.MockSpan)
			assert.NotEmpty(t, mockSpan.Tag(tagTraceID))
			assert.NotEmpty(t, mockSpan.Tag(tagSpanID))
			assert.NotEmpty(t, mockSpan.Tag(tagLogID))
		})
	})
}

func TestSetLogID(t *testing.T) {
	convey.Convey("TestSetLogID", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			ctx := lc.WithLogID(context.Background(), "logid")
			SetLogID(ctx, span)

			mockSpan := span.(*mocktracer.MockSpan)
			assert.NotEmpty(t, mockSpan.Tag(tagLogID))
		})
	})
}

func TestSetTraceTag(t *testing.T) {
	convey.Convey("TestSetCommonTag", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			ctx := lc.WithLogID(context.Background(), "logid")
			SetTraceTag(ctx, span)

			mockSpan := span.(*mocktracer.MockSpan)
			assert.NotEmpty(t, mockSpan.Tag(tagTraceID))
			assert.NotEmpty(t, mockSpan.Tag(tagSpanID))
		})
	})
}

func TestSetRequest(t *testing.T) {
	convey.Convey("TestSetRequest", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			SetRequest(span, "request")

			mockSpan := span.(*mocktracer.MockSpan)
			records := mockSpan.Logs()
			assert.Equal(t, 1, len(records))
			fields := records[0].Fields
			assert.Equal(t, 1, len(fields))
			f := fields[0]
			assert.Equal(t, logFieldsRequest, f.Key)
			assert.Equal(t, reflect.String, f.ValueKind)
			assert.Equal(t, "request", f.ValueString)
		})
	})
}

func TestSetResponse(t *testing.T) {
	convey.Convey("TestSetResponse", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			SetResponse(span, "response")

			mockSpan := span.(*mocktracer.MockSpan)
			records := mockSpan.Logs()
			assert.Equal(t, 1, len(records))
			fields := records[0].Fields
			assert.Equal(t, 1, len(fields))
			f := fields[0]
			assert.Equal(t, logFieldsResponse, f.Key)
			assert.Equal(t, reflect.String, f.ValueKind)
			assert.Equal(t, "response", f.ValueString)
		})
	})
}

func TestGetTraceID(t *testing.T) {
	convey.Convey("TestGetTraceID", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			id := GetTraceID(span)
			assert.NotEmpty(t, id)
		})
	})
}

func TestGetSpanID(t *testing.T) {
	convey.Convey("TestGetSpanID", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			id := GetSpanID(span)
			assert.NotEmpty(t, id)
		})
	})
}

func Test_spanContextToJaegerContext(t *testing.T) {
	convey.Convey("Test_spanContextToJaegerContext", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")
			span.SetTag("logid", "logid")

			spanContext := spanContextToJaegerContext(span.Context())
			assert.NotNil(t, spanContext)
		})
	})
}
