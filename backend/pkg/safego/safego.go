package safego

import "go.uber.org/zap"

// Go 安全启动 goroutine，自动 recover panic 并记录日志。
// 防止因 panic 导致整个进程崩溃。
func Go(logger *zap.Logger, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("goroutine_panic_recovered",
					zap.Any("panic", r),
					zap.Stack("stack"),
				)
			}
		}()
		fn()
	}()
}
