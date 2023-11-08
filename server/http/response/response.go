package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/why444216978/go-util/assert"

	"github.com/air-go/rpc/library/logger"
)

// Code is response code
type Code int

const (
	CodeSuccess     Code = 0
	CodeParams      Code = http.StatusBadRequest
	CodeForbidden   Code = http.StatusForbidden
	CodeNotFound    Code = http.StatusNotFound
	CodeServer      Code = http.StatusInternalServerError
	CodeUnavailable Code = http.StatusServiceUnavailable
	CodeTimeout     Code = http.StatusGatewayTimeout
)

var codeToast = map[Code]string{
	CodeSuccess:     "success",
	CodeParams:      "参数错误",
	CodeForbidden:   "暂无权限",
	CodeNotFound:    "资源不存在",
	CodeUnavailable: "服务器暂时不可用",
	CodeTimeout:     "请求超时",
	CodeServer:      "服务器错误",
}

// response is json response struct
type response struct {
	Code    Code        `json:"code"`
	Toast   string      `json:"toast"`
	ErrMsg  string      `json:"errmsg"`
	Data    interface{} `json:"data"`
	LogID   string      `json:"log_id"`
	TraceID string      `json:"trace_id"`
}

// ResponseJSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func ResponseJSON(c *gin.Context, code Code, data ...interface{}) {
	m := codeToast[code]
	resp := response{
		Code:    code,
		Toast:   m,
		ErrMsg:  m,
		Data:    struct{}{},
		LogID:   logger.ValueLogID(c.Request.Context()),
		TraceID: logger.ValueTraceID(c.Request.Context()),
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
