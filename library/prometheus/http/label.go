package http

import (
	"strconv"

	"github.com/air-go/rpc/library/app"
	"github.com/gin-gonic/gin"
	"github.com/why444216978/go-util/sys"
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
			ip, _ := sys.LocalIP()
			return ip
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
}
