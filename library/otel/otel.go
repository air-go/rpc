package otel

import (
	"context"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/why444216978/go-util/assert"
	"github.com/why444216978/go-util/sys"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/air-go/rpc/library/app"
)

type Option struct {
	sampler            tracesdk.Sampler
	propagation        propagation.TextMapPropagator
	maxQueueSize       int
	maxExportBatchSize int
	limits             tracesdk.SpanLimits
	idGenerator        tracesdk.IDGenerator
	attrs              []attribute.KeyValue
	exportTimeout      time.Duration
	batchTimeout       time.Duration
}

type OptionFunc func(*Option)

func WithSampler(s tracesdk.Sampler) OptionFunc {
	return func(o *Option) { o.sampler = s }
}

func WithClientPropagation(p propagation.TextMapPropagator) OptionFunc {
	return func(o *Option) { o.propagation = p }
}

func WithMaxQueueSize(s int) OptionFunc {
	return func(o *Option) { o.maxQueueSize = s }
}

func WithMaxExportBatchSize(s int) OptionFunc {
	return func(o *Option) { o.maxExportBatchSize = s }
}

func WithRawSpanLimits(l tracesdk.SpanLimits) OptionFunc {
	return func(o *Option) { o.limits = l }
}

func WithIDGenerator(g tracesdk.IDGenerator) OptionFunc {
	return func(o *Option) { o.idGenerator = g }
}

func WithResource(attrs []attribute.KeyValue) OptionFunc {
	return func(o *Option) { o.attrs = attrs }
}

func WithExportTimeout(timeout time.Duration) OptionFunc {
	return func(o *Option) { o.exportTimeout = timeout }
}

func WithBatchTimeout(timeout time.Duration) OptionFunc {
	return func(o *Option) { o.batchTimeout = timeout }
}

func defaultOption() *Option {
	return &Option{
		sampler:            tracesdk.AlwaysSample(),
		maxQueueSize:       20480,
		maxExportBatchSize: 4096,
		limits: tracesdk.SpanLimits{
			AttributeCountLimit:         1024,
			EventCountLimit:             1024,
			LinkCountLimit:              1024,
			AttributePerEventCountLimit: 1024,
			AttributePerLinkCountLimit:  1024,
			AttributeValueLengthLimit:   1024,
		},
		exportTimeout: time.Second,
		batchTimeout:  time.Second,
	}
}

type otelTracer struct {
	TP *tracesdk.TracerProvider
}

var once sync.Once

func NewTracer(serviceName string, exporter tracesdk.SpanExporter, opts ...OptionFunc) (err error) {
	once.Do(func() {
		options := defaultOption()
		for _, o := range opts {
			o(options)
		}

		tp := tracesdk.NewTracerProvider(
			tracesdk.WithSampler(options.sampler),
			tracesdk.WithBatcher(
				exporter,
				tracesdk.WithMaxQueueSize(options.maxQueueSize),
				tracesdk.WithMaxExportBatchSize(options.maxExportBatchSize),
				tracesdk.WithExportTimeout(options.exportTimeout),
				tracesdk.WithBatchTimeout(options.batchTimeout),
			),
			tracesdk.WithIDGenerator(options.idGenerator),
			tracesdk.WithRawSpanLimits(options.limits),
			tracesdk.WithResource(setResource(serviceName, options.attrs...)),
		)

		otel.SetTracerProvider(tp)

		textMapPropagators := []propagation.TextMapPropagator{
			propagation.TraceContext{},
			propagation.Baggage{},
			b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader | b3.B3SingleHeader)),
		}

		if !assert.IsNil(options.propagation) {
			textMapPropagators = append(textMapPropagators, options.propagation)
		}

		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(textMapPropagators...))
	})

	return
}

func Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(name, opts...)
}

func TraceID(span trace.Span) string {
	traceID := span.SpanContext().TraceID()
	if traceID.IsValid() {
		return traceID.String()
	}
	return ""
}

func SpanID(span trace.Span) string {
	spanID := span.SpanContext().SpanID()
	if spanID.IsValid() {
		return spanID.String()
	}
	return ""
}

func CheckHasTraceID(ctx context.Context) bool {
	return trace.SpanFromContext(ctx).SpanContext().HasTraceID()
}

func setResource(serviceName string, attrs ...attribute.KeyValue) *resource.Resource {
	hostName, _ := os.Hostname()
	localIP, _ := sys.LocalIP()
	attrs = append(attrs, []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
		semconv.HostNameKey.String(hostName),
		semconv.NetHostIPKey.String(localIP),
		semconv.NetHostPortKey.Int(app.Port()),
		semconv.ProcessPIDKey.Int(syscall.Getpid()),
	}...)
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)
}
