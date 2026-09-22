package repository

import (
	"errors"
	"fmt"

	"github.com/medasset/medasset/internal/model"
	"gorm.io/gorm"
)

// ScrapRepository 报废申请仓储。
type ScrapRepository struct {
	db *gorm.DB
}

func NewScrapRepository(db *gorm.DB) *ScrapRepository {
	return &ScrapRepository{db: db}
}

// DB 返回底层数据库句柄（供 service 层开启事务）。
func (r *ScrapRepository) DB() *gorm.DB { return r.db }

// Create 创建报废申请。
func (r *ScrapRepository) Create(s *model.ScrapRequest) error {
	return r.db.Create(s).Error
}

// FindByID 按 ID 查询。
func (r *ScrapRepository) FindByID(id uint) (*model.ScrapRequest, error) {
	var s model.ScrapRequest
	err := r.db.First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}

// List 分页查询报废申请。
func (r *ScrapRepository) List(page, pageSize int, status string) ([]model.ScrapRequest, int64, error) {
	var list []model.ScrapRequest
	var total int64
	q := r.db.Model(&model.ScrapRequest{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Update 更新报废申请。
func (r *ScrapRepository) Update(s *model.ScrapRequest) error {
	return r.db.Save(s).Error
}

// UpdateTx 在指定事务中更新更新报废申请。
func (r *ScrapRepository) UpdateTx(tx *gorm.DB, s *model.ScrapRequest) error {
	return tx.Save(s).Error
}

// IsScrapNoTaken 判断ScrapNo是否已存在。
func (r *ScrapRepository) IsScrapNoTaken(v string) (bool, error) {
	var n int64
	if err := r.db.Model(&model.ScrapRequest{}).Where("scrap_no = ?", v).Count(&n).Error; err != nil {
		return false, fmt.Errorf("check scrap_no: %w", err)
	}
	return n > 0, nil
}
