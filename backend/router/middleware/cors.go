package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowedOrigins []string // 允许的源，空数组默认允许 localhost 开发地址
}

// CORS 跨域中间件，根据配置限制允许的源。
// 未配置时默认允许 localhost:5173 / localhost:3000（开发环境）。
func CORS(cfg *CORSConfig) gin.HandlerFunc {
	allowed := make(map[string]bool)
	for _, o := range cfg.AllowedOrigins {
		if o == "*" {
			allowed["*"] = true
			break
		}
		allowed[strings.TrimRight(o, "/")] = true
	}

	// 如果未配置任何 origin，默认允许常见开发地址
	if len(allowed) == 0 {
		allowed["http://localhost:5173"] = true
		allowed["http://localhost:3000"] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if allowed["*"] {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
