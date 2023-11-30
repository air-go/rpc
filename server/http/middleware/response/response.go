package response

import (
	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"

	lc "github.com/air-go/rpc/library/context"
	"github.com/air-go/rpc/library/logger"
)

// BodyWriter inherit the ResponseWriter of Gin and add the body field to expose the response
type BodyWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

// Write for writing responses body
func (w BodyWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

// ResponseMiddleware return a gin.HandlerFunc.
// The code after next takes effect in the reverse order.
// So must register this middleware at first.
func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bw := &BodyWriter{Body: bytes.NewBuffer(nil), ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		if bw.Body.Len() == 0 {
			return
		}

		ctx := logger.InitFieldsContainer(c.Request.Context())
		ctx = lc.WithResponseWriter(ctx, bw)
		c.Request = c.Request.WithContext(ctx)

		resp := map[string]interface{}{}
		_ = json.Unmarshal(bw.Body.Bytes(), &resp)
		logger.AddField(ctx, logger.Reflect(logger.Response, resp))
	}
}
