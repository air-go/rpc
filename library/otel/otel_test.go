package otel

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/air-go/rpc/library/otel/exporters/stdout"
)

func TestNewTracer(t *testing.T) {
	convey.Convey("TestNewTracer", t, func() {
		convey.Convey("success", func() {
			ctx := context.Background()
			serviceName := "test"

			s, err := stdout.NewStdout()
			assert.Nil(t, err)

			err = NewTracer(serviceName, s.Exporter,
				WithSampler(s.Sampler),
				WithMaxQueueSize(1024),
				WithMaxExportBatchSize(1024),
				WithRawSpanLimits(tracesdk.SpanLimits{
					AttributeCountLimit:         1024,
					EventCountLimit:             1024,
					LinkCountLimit:              1024,
					AttributePerEventCountLimit: 1024,
					AttributePerLinkCountLimit:  1024,
					AttributeValueLengthLimit:   1024,
				}),
				WithResource([]attribute.KeyValue{attribute.String("a", "a")}),
				WithExportTimeout(time.Second),
				WithBatchTimeout(time.Second),
			)
			assert.Nil(t, err)

			ctx, span := otel.Tracer("test").Start(
				ctx,
				serviceName,
				trace.WithSpanKind(trace.SpanKindClient),
			)

			fmt.Println("trace_id", CheckHasTraceID(ctx))

			span.SetStatus(codes.Error, "test error")

			span.SetAttributes(attribute.String("id", "abc"))

			span.End(trace.WithStackTrace(true))
		})
	})
}
