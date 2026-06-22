package middleware

import (
	"manage_system/pkg/errcode"
	"manage_system/pkg/response"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Casbin 权限鉴权中间件
func Casbin(enforcer *casbin.Enforcer, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleName, exists := c.Get("role_name")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Response{
				Code: errcode.ErrPermissionDenied,
				Msg:  errcode.GetMsg(errcode.ErrPermissionDenied),
				Data: nil,
			})
			return
		}

		sub := roleName.(string)  // 如 "member"
		obj := c.Request.URL.Path // 如 "/api/v1/borrows/1/approve"
		act := c.Request.Method   // 如 "POST"

		ok, err := enforcer.Enforce(sub, obj, act)
		if err != nil {
			logger.Error("casbin_enforce_error",
				zap.Error(err),
				zap.String("sub", sub),
				zap.String("obj", obj),
				zap.String("act", act),
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
				Code: errcode.ErrInternalServer,
				Msg:  errcode.GetMsg(errcode.ErrInternalServer),
				Data: nil,
			})
			return
		}

		if !ok {
			userID, _ := c.Get("user_id")
			username, _ := c.Get("username")
			logger.Warn("permission_denied",
				zap.Any("user_id", userID),
				zap.String("username", username.(string)),
				zap.String("role", sub),
				zap.String("path", obj),
				zap.String("method", act),
				zap.String("ip", c.ClientIP()),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, response.Response{
				Code: errcode.ErrPermissionDenied,
				Msg:  errcode.GetMsg(errcode.ErrPermissionDenied),
				Data: nil,
			})
			return
		}

		c.Next()
	}
}
