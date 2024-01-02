package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/why444216978/go-util/assert"

	lc "github.com/air-go/rpc/library/context"
)

// Errno is response errno
type Errno int

const (
	ErrnoSuccess     Errno = 0
	ErrnoParams      Errno = http.StatusBadRequest
	ErrnoForbidden   Errno = http.StatusForbidden
	ErrnoNotFound    Errno = http.StatusNotFound
	ErrnoServer      Errno = http.StatusInternalServerError
	ErrnoUnavailable Errno = http.StatusServiceUnavailable
	ErrnoTimeout     Errno = http.StatusGatewayTimeout
)

var codeToast = map[Errno]string{
	ErrnoSuccess:     "success",
	ErrnoParams:      "参数错误",
	ErrnoForbidden:   "暂无权限",
	ErrnoNotFound:    "资源不存在",
	ErrnoUnavailable: "服务器暂时不可用",
	ErrnoTimeout:     "请求超时",
	ErrnoServer:      "服务器错误",
}

// Response is json response struct
type Response struct {
	Errno   Errno       `json:"errno"`
	Toast   string      `json:"toast"`
	ErrMsg  string      `json:"errmsg"`
	Data    interface{} `json:"data"`
	LogID   string      `json:"log_id"`
	TraceID string      `json:"trace_id"`
}

// ResponseJSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func ResponseJSON(c *gin.Context, errno Errno, data ...interface{}) {
	m := codeToast[errno]
	resp := Response{
		Errno:   errno,
		Toast:   m,
		ErrMsg:  m,
		Data:    struct{}{},
		LogID:   lc.ValueLogID(c.Request.Context()),
		TraceID: lc.ValueTraceID(c.Request.Context()),
	}

	defer func() {
		c.JSON(http.StatusOK, resp)
		c.Abort()
	}()

	if len(data) <= 0 {
		return
	}

	switch t := (data[0]).(type) {
	case error:
		err := &ResponseError{}
		if errors.As(t, &err) {
			resp.Toast = err.Toast()
			resp.ErrMsg = err.Error()
		} else {
			resp.Toast = t.Error()
			resp.ErrMsg = t.Error()
		}
	default:
		// interface{} !=nil 比较 <type,value> 两者都是nil才是
		// 实际使用中会存在传入data type != nil，value == nil
		if !assert.IsNil(data[0]) {
			resp.Data = data[0]
		}
	}
}
