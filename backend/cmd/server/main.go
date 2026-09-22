package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fmt"

	"github.com/medasset/medasset/internal/config"
	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/database"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/router"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	level := slog.LevelInfo
	if cfg.RunMode == "debug" {
		level = slog.LevelDebug
	}
	util.InitLogger(level)

	// 初始化数据库。
	db, err := database.New(cfg)
	if err != nil {
		util.Log.Error(fmt.Sprintf(constants.LogDBInitFailed, err))
		os.Exit(1)
	}

	// 初始化 Redis（失败仅告警，限流降级为内存模式）。
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		util.Log.Warn("Redis 不可用，限流降级为内存模式", "err", err)
		rdb = nil
	}

	// 种子管理员。
	auditSvc := service.NewAuditService(repository.NewAuditRepository(db), util.Log)
	userRepo := repository.NewUserRepository(db)
	seedUserService := service.NewUserService(userRepo, auditSvc, cfg.JWTSecret, 24, util.Log)
	if err := seedUserService.SeedAdmin(); err != nil {
		util.Log.Error("初始化默认管理员失败", "err", err)
		os.Exit(1)
	}

	engine := router.New(router.Deps{DB: db, Cfg: cfg, Log: util.Log, RDB: rdb})

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: engine,
	}
	go func() {
		util.Log.Info(fmt.Sprintf(constants.LogServerStarted, cfg.ServerPort, cfg.RunMode))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			util.Log.Error("服务监听失败", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		util.Log.Error("服务关闭失败", "err", err)
	}
	util.Log.Info("服务已优雅退出")
}
