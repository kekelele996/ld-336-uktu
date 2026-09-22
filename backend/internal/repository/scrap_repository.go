package repository

import (
	"errors"
	"fmt"

	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// CreateTx 在指定事务中创建报废申请。
func (r *ScrapRepository) CreateTx(tx *gorm.DB, s *model.ScrapRequest) error {
	return tx.Create(s).Error
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

// FindByIDForUpdateTx 事务内加锁查询报废申请。
func (r *ScrapRepository) FindByIDForUpdateTx(tx *gorm.DB, id uint) (*model.ScrapRequest, error) {
	var s model.ScrapRequest
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&s, id).Error
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

// UpdateIfPendingTx 仅当申请仍为待审批时在事务内更新（CAS，防止并发重复审批）。
func (r *ScrapRepository) UpdateIfPendingTx(tx *gorm.DB, s *model.ScrapRequest) (bool, error) {
	res := tx.Model(&model.ScrapRequest{}).
		Where("id = ? AND status = ?", s.ID, constants.ScrapStatusPending).
		Updates(map[string]any{
			"status":          s.Status,
			"approver":        s.Approver,
			"approve_comment": s.ApproveComment,
			"approve_at":      s.ApproveAt,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

// IsScrapNoTaken 判断ScrapNo是否已存在。
func (r *ScrapRepository) IsScrapNoTaken(v string) (bool, error) {
	var n int64
	if err := r.db.Model(&model.ScrapRequest{}).Where("scrap_no = ?", v).Count(&n).Error; err != nil {
		return false, fmt.Errorf("check scrap_no: %w", err)
	}
	return n > 0, nil
}
