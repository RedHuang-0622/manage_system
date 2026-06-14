package middleware

import (
	"manage_system/pkg/errcode"
	"manage_system/pkg/response"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic_recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code: errcode.ErrInternalServer,
					Msg:  errcode.GetMsg(errcode.ErrInternalServer),
					Data: nil,
				})
			}
		}()
		c.Next()
	}
}

