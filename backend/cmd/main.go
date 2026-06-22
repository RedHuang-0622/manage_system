package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"manage_system/controller"
	"manage_system/dao"
	"manage_system/models"
	"manage_system/pkg/config"
	"manage_system/pkg/jwt"
	redispool "manage_system/pkg/redis"
	zaplog "manage_system/pkg/zap"
	"manage_system/router"
	"manage_system/service"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("conf/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 初始化日志
	logger, err := zaplog.InitLogger(cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	// 3. 初始化数据库
	db := initDB(cfg, logger)
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("获取底层sql.DB失败", zap.Error(err))
	}
	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MySQL.ConnMaxLifetime) * time.Second)

	// 4. AutoMigrate
	if err := db.AutoMigrate(
		&models.SysUser{},
		&models.SysRole{},
		&models.LabEquipment{},
		&models.BorrowRecord{},
	); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// 5. 初始化 Redis
	rdb, err := redispool.InitRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("Redis初始化失败", zap.Error(err))
	}
	defer rdb.Close()

	// 6. 初始化 JWT Service
	jwtService := jwt.NewService(cfg.JWT, rdb)

	// 7. 初始化 Casbin
	enforcer := initCasbin(db, cfg.Casbin.ModelPath, logger)

	// 8. 种子数据
	seedData(db, enforcer, logger)

	// 9. 组装依赖
	// DAO 层
	userDAO := dao.NewUserDAO(db)
	roleDAO := dao.NewRoleDAO(db)
	equipDAO := dao.NewEquipmentDAO(db)
	borrowDAO := dao.NewBorrowDAO(db)

	// Service 层
	iamService := service.NewIAMService(db, userDAO, roleDAO, jwtService, rdb)
	equipService := service.NewEquipmentService(db, equipDAO, borrowDAO, rdb, logger)
	borrowService := service.NewBorrowService(db, borrowDAO, equipDAO, equipService, logger)

	// Controller 层
	authController := controller.NewAuthController(iamService)
	equipController := controller.NewEquipmentController(equipService)
	borrowController := controller.NewBorrowController(borrowService)

	// 10. 路由
	r := router.SetupRouter(router.Dependencies{
		AuthController:      authController,
		EquipmentController: equipController,
		BorrowController:    borrowController,
		JWTService:          jwtService,
		Enforcer:            enforcer,
		Logger:              logger,
		RedisClient:         rdb,
		DB:                  db,
		CORSAllowedOrigins:  cfg.CORS.AllowedOrigins,
	})

	// 11. 启动服务
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           r,
		ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(cfg.Server.ReadHeaderTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
	}

	go func() {
		logger.Info("server_started", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server_failed", zap.Error(err))
		}
	}()

	// 12. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("server_shutting_down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server_shutdown_error", zap.Error(err))
	}

	sqlDB.Close()
	logger.Info("server_stopped")
}

func initDB(cfg *config.Config, logger *zap.Logger) *gorm.DB {
	dsn := cfg.MySQL.DSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	return db
}

func initCasbin(db *gorm.DB, modelPath string, logger *zap.Logger) *casbin.Enforcer {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		logger.Fatal("Casbin适配器初始化失败", zap.Error(err))
	}

	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		logger.Fatal("Casbin初始化失败", zap.Error(err))
	}

	if err := enforcer.LoadPolicy(); err != nil {
		logger.Fatal("Casbin加载策略失败", zap.Error(err))
	}

	return enforcer
}

func seedData(db *gorm.DB, enforcer *casbin.Enforcer, logger *zap.Logger) {
	// 种子角色（幂等）
	roles := []models.SysRole{
		{RoleName: "super_admin", Description: "超级管理员（指导老师）", IsSystem: 1},
		{RoleName: "lab_admin", Description: "实验室负责人", IsSystem: 1},
		{RoleName: "equipment_manager", Description: "设备管理员", IsSystem: 1},
		{RoleName: "member", Description: "普通成员", IsSystem: 1},
		{RoleName: "viewer", Description: "观察员（只读）", IsSystem: 1},
	}

	for _, role := range roles {
		var existing models.SysRole
		if err := db.Where("role_name = ?", role.RoleName).First(&existing).Error; err != nil {
			db.Create(&role)
		}
	}

	// 种子管理员用户（幂等）
	var admin models.SysUser
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
		var superAdminRole models.SysRole
		db.Where("role_name = ?", "super_admin").First(&superAdminRole)
		newAdmin := models.SysUser{
			Username:     "admin",
			PasswordHash: string(hash),
			RealName:     "系统管理员",
			RoleID:       superAdminRole.ID,
			Status:       1,
		}
		db.Create(&newAdmin)
		logger.Info("种子管理员已创建", zap.String("username", "admin"))
	}

	// 种子 Casbin 策略（逐条幂等，不会因已有旧策略而跳过新角色）
	policies := [][]string{
		// super_admin: 全局管理
		{"super_admin", "/api/v1/*", ".*"},
		// lab_admin: 实验室负责人
		{"lab_admin", "/api/v1/users*", ".*"},
		{"lab_admin", "/api/v1/equipments*", ".*"},
		{"lab_admin", "/api/v1/borrows*", ".*"},
		{"lab_admin", "/api/v1/roles*", "GET"},
		{"lab_admin", "/api/v1/auth/logout", "POST"},
		{"lab_admin", "/api/v1/auth/refresh", "POST"},
		// equipment_manager: 设备管理员
		{"equipment_manager", "/api/v1/equipments*", ".*"},
		{"equipment_manager", "/api/v1/borrows*", ".*"},
		{"equipment_manager", "/api/v1/roles", "GET"},
		{"equipment_manager", "/api/v1/auth/logout", "POST"},
		{"equipment_manager", "/api/v1/auth/refresh", "POST"},
		{"equipment_manager", "/api/v1/users/\\d+/password", "PUT"},
		// member: 普通成员
		{"member", "/api/v1/equipments*", "GET"},
		{"member", "/api/v1/borrows/apply", "POST"},
		{"member", "/api/v1/borrows/my", "GET"},
		{"member", "/api/v1/borrows/\\d+/return", "POST"},
		{"member", "/api/v1/borrows/\\d+/cancel", "POST"},
		{"member", "/api/v1/users/\\d+/password", "PUT"},
		{"member", "/api/v1/roles", "GET"},
		{"member", "/api/v1/auth/logout", "POST"},
		{"member", "/api/v1/auth/refresh", "POST"},
		// viewer: 观察员（只读）
		{"viewer", "/api/v1/equipments*", "GET"},
		{"viewer", "/api/v1/borrows/my", "GET"},
		{"viewer", "/api/v1/borrows/pending", "GET"},
		{"viewer", "/api/v1/roles", "GET"},
		{"viewer", "/api/v1/users/\\d+/password", "PUT"},
		{"viewer", "/api/v1/auth/logout", "POST"},
		{"viewer", "/api/v1/auth/refresh", "POST"},
	}
	for _, p := range policies {
		if !enforcer.HasPolicy(p[0], p[1], p[2]) {
			enforcer.AddPolicy(p[0], p[1], p[2])
		}
	}

	// 角色自映射 + 继承链
	for _, r := range []string{"viewer", "member", "equipment_manager", "lab_admin", "super_admin"} {
		if !enforcer.HasGroupingPolicy(r, r) {
			enforcer.AddGroupingPolicy(r, r)
		}
	}
	if !enforcer.HasGroupingPolicy("super_admin", "lab_admin") {
		enforcer.AddGroupingPolicy("super_admin", "lab_admin")
	}
	if !enforcer.HasGroupingPolicy("lab_admin", "equipment_manager") {
		enforcer.AddGroupingPolicy("lab_admin", "equipment_manager")
	}
	enforcer.SavePolicy()
	logger.Info("Casbin策略已同步")
}


