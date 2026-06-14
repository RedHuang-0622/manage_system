package middleware

import (
	"context"
	"net/http"
	"time"

	"manage_system/pkg/errcode"
	"manage_system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit 基于 Redis 的简单固定窗口限流中间件。
// keyPrefix: Redis key 前缀；limit: 窗口内最大请求数；window: 窗口时长。
func RateLimit(rdb *redis.Client, keyPrefix string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ratelimit:" + keyPrefix + ":" + c.ClientIP()
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redis 不可用时放行（应配合告警）
			c.Next()
			return
		}

		// 首次请求时设置 TTL
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.Response{
				Code: errcode.ErrRateLimited,
				Msg:  "请求过于频繁，请稍后再试",
				Data: nil,
			})
			return
		}

		c.Next()
	}
}
