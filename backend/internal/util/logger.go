package util

import (
	"log/slog"
	"os"
)

// 全局 slog 日志器，所有 handler/service/middleware 统一引用。
var Log *slog.Logger

// InitLogger 初始化结构化日志器。
func InitLogger(level slog.Level) {
	opts := &slog.HandlerOptions{Level: level}
	Log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
