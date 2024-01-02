package limiter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/server/http/response"
)

func Limiter(maxBurstSize int, l logger.Logger) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(time.Second*1), maxBurstSize)
	return func(c *gin.Context) {
		if limiter.Allow() {
			c.Next()
			return
		}

		ctx := logger.InitFieldsContainer(c.Request.Context())

		logger.AddField(ctx,
			logger.Reflect(logger.Status, http.StatusInternalServerError),
		)
		c.Request = c.Request.WithContext(ctx)

		l.Error(ctx, "limiter") // 这里不能打Fatal和Panic，否则程序会退出
		response.ResponseJSON(c, http.StatusServiceUnavailable, response.WrapToast(http.StatusText(http.StatusServiceUnavailable)))
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}
