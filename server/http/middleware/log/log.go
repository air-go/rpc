package log

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httputil"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/why444216978/go-util/sys"

	"github.com/air-go/rpc/library/app"
	lc "github.com/air-go/rpc/library/context"
	"github.com/air-go/rpc/library/logger"
)

func LoggerMiddleware(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := logger.InitFieldsContainer(c.Request.Context())

		start := time.Now()

		serverIP, _ := sys.LocalIP()

		logID := logger.ExtractLogID(c.Request)
		ctx = lc.WithLogID(ctx, logID)

		// req := logger.GetRequestBody(c.Request)
		req, _ := httputil.DumpRequest(c.Request, true)

		// Next之前这里需要写入ctx，否则会丢失log、断开trace
		logger.AddField(ctx,
			logger.Reflect(logger.LogID, logID),
			logger.Reflect(logger.TraceID, lc.ValueTraceID(ctx)),
			logger.Reflect(logger.RequestHeader, c.Request.Header),
			logger.Reflect(logger.Method, c.Request.Method),
			logger.Reflect(logger.Request, base64.StdEncoding.EncodeToString(req)),
			logger.Reflect(logger.Response, make(map[string]interface{})),
			logger.Reflect(logger.ClientIP, c.ClientIP()),
			logger.Reflect(logger.ClientPort, 0),
			logger.Reflect(logger.ServerIP, serverIP),
			logger.Reflect(logger.ServerPort, app.Port()),
			logger.Reflect(logger.API, c.Request.URL.Path),
			logger.Reflect(logger.URI, c.Request.RequestURI))
		c.Request = c.Request.WithContext(ctx)

		var doneFlag int32
		done := make(chan struct{}, 1)
		defer func() {
			done <- struct{}{}
			atomic.StoreInt32(&doneFlag, 1)

			ctx := c.Request.Context()

			logger.AddField(ctx,
				logger.Reflect(logger.Status, c.Writer.Status()),
				logger.Reflect(logger.Cost, time.Since(start).Milliseconds()))

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

				status := 499
				err := ctx.Err()

				if errors.Is(err, context.DeadlineExceeded) {
					status = http.StatusGatewayTimeout
				}

				logger.AddField(ctx,
					logger.Reflect(logger.Status, status),
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
