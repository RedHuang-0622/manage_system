package safego

import "go.uber.org/zap"

// Go 安全启动 goroutine，自动 recover panic 并记录日志。
// 防止因 panic 导致整个进程崩溃。
// logger 为 nil 时静默丢弃 — 避免 recover 自身二次 panic。
func Go(logger *zap.Logger, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if logger != nil {
					logger.Error("goroutine_panic_recovered",
						zap.Any("panic", r),
						zap.Stack("stack"),
					)
				}
			}
		}()
		fn()
	}()
}
