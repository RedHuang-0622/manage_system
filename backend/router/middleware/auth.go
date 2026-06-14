package middleware

import (
	"manage_system/pkg/errcode"
	"manage_system/pkg/response"
	"net/http"
	"strings"

	"manage_system/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// Auth JWT 认证中间件
func Auth(jwtService *jwt.Service) gin.HandlerFunc {
	// 白名单路径
	skipPaths := map[string]bool{
		"POST /api/v1/auth/login": true,
		"GET /api/v1/health":      true,
	}

	return func(c *gin.Context) {
		// 检查白名单
		key := c.Request.Method + " " + c.Request.URL.Path
		if skipPaths[key] {
			c.Next()
			return
		}

		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
				Code: errcode.ErrTokenMissing,
				Msg:  errcode.GetMsg(errcode.ErrTokenMissing),
				Data: nil,
			})
			return
		}

		// 检查黑名单
		if jwtService.IsInBlacklist(tokenStr) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
				Code: errcode.ErrTokenInvalid,
				Msg:  "Token已失效",
				Data: nil,
			})
			return
		}

		// 解析并验证 Token
		claims, err := jwtService.ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
				Code: errcode.ErrTokenInvalid,
				Msg:  errcode.GetMsg(errcode.ErrTokenInvalid),
				Data: nil,
			})
			return
		}

		// 注入 Context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role_id", claims.RoleID)
		c.Set("role_name", claims.RoleName)

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return ""
}
