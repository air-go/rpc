package trace

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go/ext"
	opentracingLog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/why444216978/go-util/assert"
	"github.com/why444216978/go-util/conversion"

	"github.com/air-go/rpc/library/logger"
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
		ctx := c.Request.Context()

		ctx, span, traceID := jaegerHTTP.ExtractHTTP(ctx, c.Request)
		if !assert.IsNil(span) {
			defer span.Finish()
		}
		ctx = logger.WithTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		ctx = c.Request.Context()

		fields := logger.ValueFields(ctx)
		request := logger.Find(logger.Request, fields)
		req, _ := request.Value().(string)

		response := logger.Find(logger.Response, fields)
		resp, _ := conversion.JsonEncode(response.Value())

		logID := logger.ValueLogID(ctx)
		jaegerHTTP.SetHTTPLog(span, logID, req, resp)

		if len(c.Errors) > 0 {
			span.LogFields(opentracingLog.Error(errors.New(c.Errors.String())))
			span.SetTag(string(ext.Error), true)
		}

		c.Request = c.Request.WithContext(ctx)
	}
}
