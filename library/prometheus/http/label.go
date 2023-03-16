package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
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
		Label: "method",
		GetValue: func(c *gin.Context) string {
			return c.FullPath()
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
