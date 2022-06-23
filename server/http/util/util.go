package util

import (
	"bytes"

	"github.com/gin-gonic/gin"
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
