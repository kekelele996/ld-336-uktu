package database

import (
	"fmt"

	"github.com/medasset/medasset/internal/config"
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New 建立 MySQL 连接并执行自动迁移。
func New(cfg *config.Config) (*gorm.DB, error) {
	level := logger.Warn
	if cfg.RunMode == "debug" {
		level = logger.Info
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
	if err != nil {
		util.Log.Error(fmt.Sprintf("数据库连接失败: %v", err))
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	if err := db.AutoMigrate(
		&model.User{},
		&model.Device{},
		&model.PurchaseRequest{},
		&model.MaintenanceRecord{},
		&model.CalibrationRecord{},
		&model.TransferRequest{},
		&model.ScrapRequest{},
		&model.AuditLog{},
	); err != nil {
		util.Log.Error(fmt.Sprintf("数据库迁移失败: %v", err))
		return nil, err
	}
	return db, nil
}
