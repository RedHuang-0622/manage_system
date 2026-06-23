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
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redis 不可用时放行（fail-open），避免阻断所有登录
			c.Next()
			return
		}

		// 首次请求时设置 TTL。如果 Expire 失败（Redis 闪断），key 会永久
		// 存在导致该 IP 被永久限流。失败时立即删除 key 做 fail-safe 自愈。
		if count == 1 {
			if err := rdb.Expire(ctx, key, window).Err(); err != nil {
				rdb.Del(context.Background(), key) // best-effort cleanup
				c.Next()
				return
			}
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
