package trace

import (
	"github.com/gin-gonic/gin"
	"github.com/why444216978/go-util/conversion"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/air-go/rpc/library/app"
	"github.com/air-go/rpc/library/logger"
	libraryOtel "github.com/air-go/rpc/library/otel"
)

// OpentelemetryMiddleware is opentelemetry
// Register before LoggerMiddleware
// OpentelemetryMiddleware Before c.Next() > LoggerMiddleware Before c.Next() >  LoggerMiddleware After c.Next() > OpentelemetryMiddleware after c.Next()
//
// The code before next takes effect in the order of use
// The code after next takes effect in the reverse order
func OpentelemetryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := logger.InitFieldsContainer(c.Request.Context())

		ctx = libraryOtel.ExtractHTTPBaggage(ctx, c.Request.Header)
		opts := []trace.SpanStartOption{
			trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", c.Request)...),
			trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(c.Request)...),
			trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(app.Name(), c.FullPath(), c.Request)...),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		spanName := c.Request.URL.Path
		ctx, span := libraryOtel.Tracer().Start(ctx, spanName, opts...)
		defer span.End()

		traceID := libraryOtel.TraceID(span)
		spanID := libraryOtel.SpanID(span)

		ctx = logger.WithTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		ctx = c.Request.Context()

		request := logger.FindField(ctx, logger.Request)
		req, _ := request.Value().(string)

		response := logger.FindField(ctx, logger.Response)
		resp, _ := conversion.JsonEncode(response.Value())

		span.AddEvent("request", trace.WithAttributes([]attribute.KeyValue{
			libraryOtel.AttributeLogID.String(logger.ValueLogID(ctx)),
			libraryOtel.AttributeRequest.String(req),
			libraryOtel.AttributeResponse.String(resp),
		}...))

		status := c.Writer.Status()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(status)

		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(status, trace.SpanKindServer)
		attrs = append(attrs, []attribute.KeyValue{
			libraryOtel.AttributeTraceID.String(traceID),
			libraryOtel.AttributeSpanID.String(spanID),
		}...)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
		if len(c.Errors) > 0 {
			span.SetStatus(codes.Error, c.Errors.String())
			span.SetAttributes(libraryOtel.AttributeGinError.String(c.Errors.String()))
		}

		c.Request = c.Request.WithContext(ctx)
	}
}
