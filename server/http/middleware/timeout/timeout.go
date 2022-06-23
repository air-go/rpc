package timeout

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type headerKey string

const (
	timeoutKey headerKey = "Timeout-Millisecond"
	startKey   headerKey = "Timeout-StartAt"
)

func (k headerKey) String() string {
	return string(k)
}

// TimeoutMiddleware is used to pass the remaining timeout
func TimeoutMiddleware(timeout time.Duration) func(c *gin.Context) {
	return func(c *gin.Context) {
		remain := timeout
		headerTimeout := c.Request.Header.Get(timeoutKey.String())
		if headerTimeout != "" {
			t, _ := strconv.ParseInt(headerTimeout, 10, 64)
			remain = time.Duration(t) * time.Millisecond
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), remain)
		_ = cancel

		ctx = SetStart(ctx, remain.Milliseconds())

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// SetStart for marking current service request start
func SetStart(ctx context.Context, timeout int64) context.Context {
	ctx = context.WithValue(ctx, timeoutKey, timeout)
	return context.WithValue(ctx, startKey, nowMillisecond())
}

// CalcRemainTimeout is used to calculate remain timeout
func CalcRemainTimeout(ctx context.Context) (int64, error) {
	timeout, ok := ctx.Value(timeoutKey).(int64)
	if !ok {
		return 0, nil
	}

	startAt, ok := ctx.Value(startKey).(int64)
	if !ok {
		return 0, errors.New("miss startAt")
	}

	remain := timeout - (nowMillisecond() - startAt)
	if remain < 0 {
		return 0, errors.New("timeout < diff, context deadline exceeded")
	}

	return remain, nil
}

// SetHeader save timeout field to http.Header
func SetHeader(ctx context.Context, header http.Header) (err error) {
	remain, err := CalcRemainTimeout(ctx)
	if err != nil {
		return
	}
	header.Set(timeoutKey.String(), strconv.FormatInt(remain, 10))
	return
}

// nowMillisecond return now time nanosecond
func nowMillisecond() int64 {
	return time.Now().UnixNano() / 1e6
}
