package controller

import (
	"strconv"
	"strings"

	"manage_system/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// parseServiceError 解析 service 层返回的错误，提取错误码和消息
// service 层错误格式: [错误码] 消息
func parseServiceError(err error) (int, string) {
	if err == nil {
		return 0, ""
	}
	msg := err.Error()
	// 尝试解析 [code] message 格式
	if strings.HasPrefix(msg, "[") {
		if idx := strings.Index(msg, "] "); idx > 1 {
			codeStr := msg[1:idx]
			if code, err := strconv.Atoi(codeStr); err == nil {
				return code, msg[idx+2:]
			}
		}
	}
	return errcode.ErrInternalServer, "系统内部错误"
}

// codeToHTTP 业务错误码 → HTTP 状态码
func codeToHTTP(code int) int {
	switch {
	case code == errcode.ErrAuthFailed || code == errcode.ErrAccountDisabled:
		return 401
	case code == errcode.ErrTokenMissing || code == errcode.ErrTokenInvalid:
		return 401
	case code == errcode.ErrPermissionDenied:
		return 403
	case code == errcode.ErrNotFound || code == errcode.ErrUserNotFound || code == errcode.ErrEquipmentNotFound:
		return 404
	case code == errcode.ErrConflict || code == errcode.ErrUserExists || code == errcode.ErrDuplicateBorrow:
		return 409
	case code >= 5000:
		return 500
	default:
		return 400
	}
}

// extractToken 从 Authorization Header 提取 Bearer Token
func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return ""
}
