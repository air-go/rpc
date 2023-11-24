package trace

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go/ext"
	opentracingLog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/why444216978/go-util/assert"

	lc "github.com/air-go/rpc/library/context"
	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/opentracing"
	jaegerHTTP "github.com/air-go/rpc/library/opentracing/http"
)

// OpentracingMiddleware is opentracing
// Register before LoggerMiddleware
// OpentracingMiddleware Before c.Next() > LoggerMiddleware Before c.Next() >  LoggerMiddleware After c.Next() > OpentracingMiddleware after c.Next()
//
// The code before next takes effect in the order of use
// The code after next takes effect in the reverse order
func OpentracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := logger.InitFieldsContainer(c.Request.Context())

		ctx, span, traceID := jaegerHTTP.ExtractHTTP(ctx, c.Request)
		if !assert.IsNil(span) {
			defer span.Finish()
		}
		ctx = lc.WithTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		ctx = c.Request.Context()

		opentracing.SetLogID(ctx, span)
		opentracing.SetRequest(span, logger.FindField(ctx, logger.Request).Value())
		opentracing.SetResponse(span, logger.FindField(ctx, logger.Response).Value())

		if len(c.Errors) > 0 {
			span.LogFields(opentracingLog.Error(errors.New(c.Errors.String())))
			span.SetTag(string(ext.Error), true)
		}

		c.Request = c.Request.WithContext(ctx)
	}
}
