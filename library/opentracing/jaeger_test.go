package opentracing

import (
	"context"
	"testing"

	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
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

func TestSetRequest(t *testing.T) {
	convey.Convey("TestSetRequest", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			SetRequest(span, "request")
		})
	})
}

func TestSetResponse(t *testing.T) {
	convey.Convey("TestSetResponse", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			SetResponse(span, "response")
		})
	})
}

func TestSetCommonTag(t *testing.T) {
	convey.Convey("TestSetCommonTag", t, func() {
		convey.Convey("success", func() {
			tracer := mocktracer.New()
			span := tracer.StartSpan("test")

			SetCommonTag(context.Background(), span)
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

			spanContext := spanContextToJaegerContext(span.Context())
			assert.NotNil(t, spanContext)
		})
	})
}
