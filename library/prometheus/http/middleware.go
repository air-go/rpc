package http

import (
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPMetricsMiddleware return a gin.HandlerFunc.
// Panic and response is executed after c.Next.
// The code after next takes effect in the reverse order.
// So must register after PanicMiddleware and LoggerMiddleware.
func HTTPMetricsMiddleware(opts ...OptionFunc) gin.HandlerFunc {
	metrics := NewHTTPMetrics(opts...)
	metrics.Register(nil)
	return func(c *gin.Context) {
		// handle filters
		for _, filter := range metrics.getFilters() {
			if !filter(c) {
				return
			}
		}

		defer func() {
			if err := recover(); err != nil {
				metrics.withPanicValues()
				// keep panic, used by PanicMiddleware
				panic(err)
			}
		}()

		start := time.Now()
		c.Next()

		metrics.withLabelValues(c, time.Now().Sub(start))
	}
}
