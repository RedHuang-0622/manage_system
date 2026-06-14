package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency_ms", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		}

		if rid, ok := c.Get("request_id"); ok {
			fields = append(fields, zap.String("request_id", rid.(string)))
		}
		if userID, ok := c.Get("user_id"); ok {
			fields = append(fields, zap.Any("user_id", userID))
		}

		switch {
		case status >= 500:
			logger.Error("request", fields...)
		case status >= 400:
			logger.Warn("request", fields...)
		default:
			logger.Info("request", fields...)
		}
	}
}
