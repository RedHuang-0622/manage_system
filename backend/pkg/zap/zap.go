package zaplog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"manage_system/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// dailyWriter wraps lumberjack.Logger to add daily rotation on top of
// Lumberjack's size-based rotation.  On every Write it checks whether the
// date has changed and, if so, forces a rotation so logs naturally split
// per day.  A background goroutine also triggers rotation at midnight to
// cover idle periods between writes.
type dailyWriter struct {
	logger *lumberjack.Logger
	dir    string // e.g. "logs"
	base   string // e.g. "app"
	ext    string // e.g. ".log"
	today  string // "2006-01-02"
	mu     sync.Mutex
}

func (w *dailyWriter) Write(p []byte) (int, error) {
	date := time.Now().Format("2006-01-02")

	w.mu.Lock()
	needRotate := date != w.today
	if needRotate {
		w.today = date
		w.logger.Filename = filepath.Join(w.dir, fmt.Sprintf("%s-%s%s", w.base, date, w.ext))
	}
	w.mu.Unlock()

	if needRotate {
		_ = w.logger.Rotate() // safe to ignore: lumberjack handles races internally
	}
	return w.logger.Write(p)
}

// startDailyRotate runs a background goroutine that forces rotation at
// midnight, so the filename always reflects the current date even when
// the application is idle around the day boundary.
func (w *dailyWriter) startDailyRotate() {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 1, 0, now.Location())
			time.Sleep(next.Sub(now))

			date := time.Now().Format("2006-01-02")
			w.mu.Lock()
			if date != w.today {
				w.today = date
				w.logger.Filename = filepath.Join(w.dir, fmt.Sprintf("%s-%s%s", w.base, date, w.ext))
				w.mu.Unlock()
				_ = w.logger.Rotate()
			} else {
				w.mu.Unlock()
			}
		}
	}()
}

func InitLogger(cfg config.LogConfig) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level.SetLevel(zapcore.InfoLevel)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Derive a date-stamped filename from the configured path.
	// "logs/app.log" → "logs/app-2026-06-21.log"
	dir := filepath.Dir(cfg.Path)
	ext := filepath.Ext(cfg.Path)
	base := strings.TrimSuffix(filepath.Base(cfg.Path), ext)
	today := time.Now().Format("2006-01-02")

	dw := &dailyWriter{
		logger: &lumberjack.Logger{
			Filename:   filepath.Join(dir, fmt.Sprintf("%s-%s%s", base, today, ext)),
			MaxSize:    cfg.MaxSize,    // MB — size-based rotation within a day
			MaxBackups: cfg.MaxBackups, // keep up to N old daily files
			MaxAge:     cfg.MaxAge,     // days before an old file is purged
			Compress:   true,
			LocalTime:  true,
		},
		dir:   dir,
		base:  base,
		ext:   ext,
		today: today,
	}
	dw.startDailyRotate()

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(dw)),
		level,
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}
