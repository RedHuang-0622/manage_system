package router

import (
	"context"
	"time"

	"manage_system/controller"
	"manage_system/pkg/jwt"
	"manage_system/router/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Dependencies struct {
	AuthController      *controller.AuthController
	EquipmentController *controller.EquipmentController
	BorrowController    *controller.BorrowController
	JWTService          *jwt.Service
	Enforcer            *casbin.Enforcer
	Logger              *zap.Logger
	RedisClient         *redis.Client
	DB                  *gorm.DB
	CORSAllowedOrigins  []string
}

func SetupRouter(deps Dependencies) *gin.Engine {
	r := gin.New()

	// 全局中间件链
	r.Use(middleware.Recovery(deps.Logger))
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(&middleware.CORSConfig{
		AllowedOrigins: deps.CORSAllowedOrigins,
	}))
	r.Use(middleware.Logger(deps.Logger))

	// 健康检查（验证 DB + Redis 连通性）
	r.GET("/api/v1/health", func(c *gin.Context) {
		result := gin.H{"status": "ok", "db": "ok", "redis": "ok"}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if deps.DB != nil {
			sqlDB, err := deps.DB.DB()
			if err != nil || sqlDB.PingContext(ctx) != nil {
				result["db"] = "unhealthy"
				result["status"] = "degraded"
			}
		}
		if deps.RedisClient != nil {
			if deps.RedisClient.Ping(ctx).Err() != nil {
				result["redis"] = "unhealthy"
				result["status"] = "degraded"
			}
		}
		if result["status"] == "degraded" {
			c.JSON(503, result)
			return
		}
		c.JSON(200, result)
	})

	// 公开路由（登录限流：同一 IP 1 分钟内最多 10 次）
	auth := r.Group("/api/v1/auth")
	{
		if deps.RedisClient != nil {
			auth.POST("/login", middleware.RateLimit(deps.RedisClient, "login", 10, 1*time.Minute), deps.AuthController.Login)
		} else {
			auth.POST("/login", deps.AuthController.Login)
		}
		// Refresh 不放在受保护组内 — 刷新端点必须接受已过期的 token
		//（过期 token 会被 Auth 中间件直接 reject，导致刷新死锁）。
		// 签名验证和黑名单检查在 RefreshToken handler 内部完成。
		auth.POST("/refresh", deps.AuthController.RefreshToken)
	}

	// 受保护路由组（中间件链: Recovery → Logger → Auth → Casbin）
	protected := r.Group("/api/v1")
	protected.Use(middleware.Auth(deps.JWTService, deps.Logger))
	protected.Use(middleware.Casbin(deps.Enforcer, deps.Logger))
	{
		// 认证
		protected.POST("/auth/logout", deps.AuthController.Logout)

		// 角色
		protected.GET("/roles", deps.AuthController.ListRoles)

		// 用户管理
		protected.GET("/users", deps.AuthController.ListUsers)
		protected.GET("/users/:id", deps.AuthController.GetUser)
		protected.POST("/users", deps.AuthController.CreateUser)
		protected.PUT("/users/:id", deps.AuthController.UpdateUser)
		protected.POST("/users/:id/disable", deps.AuthController.DisableUser)
		protected.PUT("/users/:id/password", deps.AuthController.ChangePassword)

		// 设备管理
		protected.GET("/equipments", deps.EquipmentController.List)
		protected.GET("/equipments/:id", deps.EquipmentController.GetByID)
		protected.POST("/equipments", deps.EquipmentController.Create)
		protected.PUT("/equipments/:id", deps.EquipmentController.Update)
		protected.DELETE("/equipments/:id", deps.EquipmentController.Disable)

		// 借阅工单
		protected.POST("/borrows/apply", deps.BorrowController.Apply)
		protected.POST("/borrows/:id/approve", deps.BorrowController.Approve)
		protected.POST("/borrows/:id/return", deps.BorrowController.Return)
		protected.POST("/borrows/:id/cancel", deps.BorrowController.Cancel)
		protected.GET("/borrows/my", deps.BorrowController.ListMyRecords)
		protected.GET("/borrows/pending", deps.BorrowController.ListPending)
		protected.GET("/borrows", deps.BorrowController.ListAll)
	}

	return r
}
