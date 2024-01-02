package panic

import (
	"net/http"

	"github.com/gin-gonic/gin"
	uruntime "github.com/why444216978/go-util/runtime"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/server/http/response"
)

func PanicMiddleware(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func(c *gin.Context) {
			if r := recover(); r != nil {
				ctx := logger.InitFieldsContainer(c.Request.Context())

				se := uruntime.WrapStackError(r)

				logger.AddField(ctx,
					logger.Error(se.Error()),
					logger.Stack(se.Stack()),
					logger.Reflect(logger.Code, http.StatusInternalServerError),
				)
				c.Request = c.Request.WithContext(ctx)

				l.DPanic(ctx, "httpPanic") // Don't Panic or Fatal.
				response.ResponseJSON(c, http.StatusInternalServerError, response.WrapToast(http.StatusText(http.StatusInternalServerError)))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}(c)
		c.Next()
	}
}
