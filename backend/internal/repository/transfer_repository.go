package repository

import (
	"errors"
	"fmt"

	"github.com/medasset/medasset/internal/model"
	"gorm.io/gorm"
)

// TransferRepository 调拨申请仓储。
type TransferRepository struct {
	db *gorm.DB
}

func NewTransferRepository(db *gorm.DB) *TransferRepository {
	return &TransferRepository{db: db}
}

// DB 返回底层数据库句柄（供 service 层开启事务）。
func (r *TransferRepository) DB() *gorm.DB { return r.db }

// Create 创建调拨申请。
func (r *TransferRepository) Create(t *model.TransferRequest) error {
	return r.db.Create(t).Error
}

// FindByID 按 ID 查询。
func (r *TransferRepository) FindByID(id uint) (*model.TransferRequest, error) {
	var t model.TransferRequest
	err := r.db.First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &t, err
}

// List 分页查询调拨申请。
func (r *TransferRepository) List(page, pageSize int, status string) ([]model.TransferRequest, int64, error) {
	var list []model.TransferRequest
	var total int64
	q := r.db.Model(&model.TransferRequest{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Update 更新调拨申请。
func (r *TransferRepository) Update(t *model.TransferRequest) error {
	return r.db.Save(t).Error
}

// UpdateTx 在指定事务中更新更新调拨申请。
func (r *TransferRepository) UpdateTx(tx *gorm.DB, t *model.TransferRequest) error {
	return tx.Save(t).Error
}

// IsTransferNoTaken 判断TransferNo是否已存在。
func (r *TransferRepository) IsTransferNoTaken(v string) (bool, error) {
	var n int64
	if err := r.db.Model(&model.TransferRequest{}).Where("transfer_no = ?", v).Count(&n).Error; err != nil {
		return false, fmt.Errorf("check transfer_no: %w", err)
	}
	return n > 0, nil
}
