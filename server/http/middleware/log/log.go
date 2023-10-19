package log

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/why444216978/go-util/sys"

	"github.com/air-go/rpc/library/app"
	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/server/http/util"
)

func LoggerMiddleware(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		start := time.Now()

		serverIP, _ := sys.LocalIP()

		logID := logger.ExtractLogID(c.Request)
		ctx = logger.WithLogID(ctx, logID)

		// req := logger.GetRequestBody(c.Request)
		req, _ := httputil.DumpRequest(c.Request, true)

		responseWriter := &util.BodyWriter{Body: bytes.NewBuffer(nil), ResponseWriter: c.Writer}
		c.Writer = responseWriter

		fields := []logger.Field{
			logger.Reflect(logger.LogID, logID),
			logger.Reflect(logger.TraceID, logger.ValueTraceID(ctx)),
			logger.Reflect(logger.Header, c.Request.Header),
			logger.Reflect(logger.Method, c.Request.Method),
			logger.Reflect(logger.Request, base64.StdEncoding.EncodeToString(req)),
			logger.Reflect(logger.Response, make(map[string]interface{})),
			logger.Reflect(logger.ClientIP, c.ClientIP()),
			logger.Reflect(logger.ClientPort, 0),
			logger.Reflect(logger.ServerIP, serverIP),
			logger.Reflect(logger.ServerPort, app.Port()),
			logger.Reflect(logger.API, c.Request.URL.Path),
			logger.Reflect(logger.URI, c.Request.RequestURI),
		}
		// Next之前这里需要写入ctx，否则会丢失log、断开trace
		ctx = logger.WithFields(ctx, fields)
		c.Request = c.Request.WithContext(ctx)

		var doneFlag int32
		done := make(chan struct{}, 1)
		defer func() {
			done <- struct{}{}
			atomic.StoreInt32(&doneFlag, 1)

			ctx := c.Request.Context()

			resp := responseWriter.Body.Bytes()
			if responseWriter.Body.Len() > 0 {
				logResponse := map[string]interface{}{}
				_ = json.Unmarshal(resp, &logResponse)
				ctx = logger.AddField(ctx, logger.Reflect(logger.Response, logResponse))
			}

			ctx = logger.AddField(ctx,
				logger.Reflect(logger.Code, c.Writer.Status()),
				logger.Reflect(logger.Cost, time.Since(start).Milliseconds()),
			)

			c.Request = c.Request.WithContext(ctx)

			l.Info(ctx, "request info")
		}()

		go func() {
			select {
			case <-done:
			case <-ctx.Done():
				if atomic.LoadInt32(&doneFlag) == 1 {
					return
				}

				code := 499
				err := ctx.Err()

				if errors.Is(err, context.DeadlineExceeded) {
					code = http.StatusGatewayTimeout
				}

				ctx = logger.AddField(ctx,
					logger.Reflect(logger.Code, code),
					logger.Reflect(logger.Cost, time.Since(start).Milliseconds()),
				)

				if err == nil {
					l.Warn(ctx, "client context Done")
				} else {
					l.Warn(ctx, err.Error())
				}
			}
		}()

		c.Next()
	}
}
