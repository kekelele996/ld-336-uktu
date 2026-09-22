package repository

import (
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/util"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestDB 创建内存 SQLite 测试库并迁移核心表。
func newTestDB(t testingT) *gorm.DB {
	util.InitLogger(-4)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Device{}, &model.PurchaseRequest{},
		&model.MaintenanceRecord{}, &model.CalibrationRecord{}, &model.TransferRequest{},
		&model.ScrapRequest{}, &model.AuditLog{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	return db
}

type testingT interface {
	Fatalf(format string, args ...any)
}
