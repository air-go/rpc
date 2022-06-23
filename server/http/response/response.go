package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/why444216978/go-util/assert"

	"github.com/air-go/rpc/library/logger"
)

// Code is response code
type Code int

// response is json response struct
type response struct {
	Code    Code        `json:"code"`
	Toast   string      `json:"toast"`
	Data    interface{} `json:"data"`
	ErrMsg  string      `json:"errmsg"`
	TraceID string      `json:"trace_id"`
}

// ResponseJSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func ResponseJSON(c *gin.Context, code Code, data interface{}, err *ResponseError) {
	if assert.IsNil(data) {
		data = make(map[string]interface{})
	}

	// prevent panic
	if err == nil {
		err = WrapToast(nil, "toast")
	}

	c.JSON(http.StatusOK, response{
		Code:    code,
		Toast:   err.Toast(),
		Data:    data,
		ErrMsg:  err.Error(),
		TraceID: logger.ValueTraceID(c.Request.Context()),
	})
	c.Abort()
}
