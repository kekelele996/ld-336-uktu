package service

import (
	"log/slog"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestServiceEnv 构建测试环境（内存 SQLite + 审计服务）。
type testEnv struct {
	db     *gorm.DB
	audit  *AuditService
	logger *slog.Logger
}

func newTestServiceEnv(t *testing.T) *testEnv {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	// 内存 SQLite 每个连接持有独立的空库，连接池固定为 1 个连接，
	// 既保证 AutoMigrate 的表对后续事务可见，也让并发提交在测试中可串行复现。
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Device{}, &model.PurchaseRequest{},
		&model.MaintenanceRecord{}, &model.CalibrationRecord{}, &model.TransferRequest{},
		&model.ScrapRequest{}, &model.AuditLog{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	logger := slog.Default()
	audit := NewAuditService(repository.NewAuditRepository(db), logger)
	return &testEnv{db: db, audit: audit, logger: logger}
}
