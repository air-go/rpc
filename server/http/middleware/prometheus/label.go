package http

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"github.com/air-go/rpc/library/app"
	lc "github.com/air-go/rpc/library/context"
	"github.com/air-go/rpc/library/logger"
	mr "github.com/air-go/rpc/server/http/middleware/response"
	sr "github.com/air-go/rpc/server/http/response"
)

type Label struct {
	Label    string
	GetValue func(*gin.Context) string
}

func (l *Label) Name() string {
	return l.Label
}

func (l *Label) Value(c *gin.Context) string {
	return l.GetValue(c)
}

var DefaultLabels = []Label{
	{
		Label: "service_name",
		GetValue: func(ctx *gin.Context) string {
			return app.Name()
		},
	},
	{
		Label: "node",
		GetValue: func(ctx *gin.Context) string {
			return app.LocalIP()
		},
	},
	{
		Label: "method",
		GetValue: func(c *gin.Context) string {
			return c.Request.Method
		},
	},
	{
		Label: "api",
		GetValue: func(c *gin.Context) string {
			return c.Request.URL.Path
		},
	},
	{
		Label: "error_number",
		GetValue: func(c *gin.Context) string {
			return strconv.Itoa(len(c.Errors))
		},
	},
	{
		Label: "http_status",
		GetValue: func(c *gin.Context) string {
			return strconv.Itoa(c.Writer.Status())
		},
	},
	{
		Label: "errno",
		GetValue: func(c *gin.Context) string {
			ctx := c.Request.Context()

			rw := lc.ValueResponseWriter(c.Request.Context())
			bw, ok := rw.(*mr.BodyWriter)
			if !ok {
				return "-1"
			}

			resp := &sr.Response{}
			if err := json.Unmarshal(bw.Body.Bytes(), resp); err != nil {
				return "-2"
			}

			code := cast.ToString(int(resp.Errno))

			logger.AddField(ctx, logger.Reflect(logger.Errno, code))

			c.Request = c.Request.WithContext(ctx)

			return code
		},
	},
}
