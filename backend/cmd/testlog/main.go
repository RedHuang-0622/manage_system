//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"manage_system/pkg/config"
	zaplog "manage_system/pkg/zap"

	"go.uber.org/zap"
)

func main() {
	tmpDir := filepath.Join(os.TempDir(), "zap-daily-test")
	os.RemoveAll(tmpDir)
	os.MkdirAll(tmpDir, 0755)
	defer fmt.Printf("\n📁 test files kept at: %s (clean up manually)\n", tmpDir)

	cfg := config.LogConfig{
		Path:       filepath.Join(tmpDir, "app.log"),
		Level:      "debug",
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     7,
	}

	logger, err := zaplog.InitLogger(cfg)
	if err != nil {
		panic(err)
	}

	// Write logs
	logger.Info("用户登录", zap.String("user", "alice"), zap.String("role", "super_admin"))
	logger.Info("设备查询", zap.String("equipment", "laptop-01"), zap.String("action", "list"))
	logger.Warn("慢查询告警", zap.String("sql", "SELECT * FROM borrow_record"), zap.Duration("elapsed", 350*time.Millisecond))
	logger.Error("越权拦截", zap.String("user", "member_001"), zap.String("path", "/api/v1/borrow/approve"), zap.Int("errcode", 4001))

	_ = logger.Sync()

	// List files
	entries, _ := os.ReadDir(tmpDir)
	fmt.Println("📋 当前日志文件：")
	for _, e := range entries {
		info, _ := e.Info()
		fmt.Printf("   %s  (%d bytes)\n", e.Name(), info.Size())
	}

	fmt.Println("\n✅ 如果文件名包含今天的日期 (app-2026-06-21.log)，则按天分片生效。")
}
