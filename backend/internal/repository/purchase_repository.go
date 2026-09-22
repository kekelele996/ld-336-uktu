package repository

import (
	"errors"
	"fmt"

	"github.com/medasset/medasset/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PurchaseRepository 采购申请仓储。
type PurchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

// DB 返回底层数据库句柄（供 service 层开启事务）。
func (r *PurchaseRepository) DB() *gorm.DB { return r.db }

// Create 创建采购申请。
func (r *PurchaseRepository) Create(p *model.PurchaseRequest) error {
	return r.db.Create(p).Error
}

// FindByID 按 ID 查询。
func (r *PurchaseRepository) FindByID(id uint) (*model.PurchaseRequest, error) {
	var p model.PurchaseRequest
	err := r.db.First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// FindByIDForUpdate 加锁查询（并发审批安全）。
func (r *PurchaseRepository) FindByIDForUpdate(tx *gorm.DB, id uint) (*model.PurchaseRequest, error) {
	var p model.PurchaseRequest
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// List 分页查询采购申请。
func (r *PurchaseRepository) List(page, pageSize int, status, department string) ([]model.PurchaseRequest, int64, error) {
	var list []model.PurchaseRequest
	var total int64
	q := r.db.Model(&model.PurchaseRequest{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if department != "" {
		q = q.Where("department = ?", department)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Update 更新采购申请。
func (r *PurchaseRepository) Update(p *model.PurchaseRequest) error {
	return r.db.Save(p).Error
}

// UpdateTx 在指定事务中更新更新采购申请。
func (r *PurchaseRepository) UpdateTx(tx *gorm.DB, p *model.PurchaseRequest) error {
	return tx.Save(p).Error
}

// UpdateStatusTx 在事务中更新采购申请状态。
func (r *PurchaseRepository) UpdateStatusTx(tx *gorm.DB, id uint, status string) error {
	res := tx.Model(&model.PurchaseRequest{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("update purchase status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// IsRequestNoTaken 判断RequestNo是否已存在。
func (r *PurchaseRepository) IsRequestNoTaken(v string) (bool, error) {
	var n int64
	if err := r.db.Model(&model.PurchaseRequest{}).Where("request_no = ?", v).Count(&n).Error; err != nil {
		return false, fmt.Errorf("check request_no: %w", err)
	}
	return n > 0, nil
}
