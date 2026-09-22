package repository

import (
	"github.com/medasset/medasset/internal/model"
	"gorm.io/gorm"
)

// AuditRepository 审计日志仓储。
type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// DB 返回底层数据库句柄（供 service 层开启事务）。
func (r *AuditRepository) DB() *gorm.DB { return r.db }

// Create 写入审计日志。
func (r *AuditRepository) Create(a *model.AuditLog) error {
	return r.db.Create(a).Error
}

// List 分页查询审计日志。
func (r *AuditRepository) List(page, pageSize int, module, username string) ([]model.AuditLog, int64, error) {
	var list []model.AuditLog
	var total int64
	q := r.db.Model(&model.AuditLog{})
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if username != "" {
		q = q.Where("username = ?", username)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
