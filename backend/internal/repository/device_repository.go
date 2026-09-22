package repository

import (
	"errors"
	"fmt"

	"github.com/medasset/medasset/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeviceRepository 设备台账仓储。
type DeviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// DB 返回底层数据库句柄（供 service 层开启事务）。
func (r *DeviceRepository) DB() *gorm.DB { return r.db }

// Create 创建设备。
func (r *DeviceRepository) Create(d *model.Device) error {
	return r.db.Create(d).Error
}

// FindByID 按 ID 查询设备。
func (r *DeviceRepository) FindByID(id uint) (*model.Device, error) {
	var d model.Device
	err := r.db.First(&d, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, err
}

// FindByIDForUpdate 按 ID 加锁查询设备（并发安全）。
func (r *DeviceRepository) FindByIDForUpdate(tx *gorm.DB, id uint) (*model.Device, error) {
	var d model.Device
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&d, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, err
}

// FindByAssetCode 按资产编号查询。
func (r *DeviceRepository) FindByAssetCode(code string) (*model.Device, error) {
	var d model.Device
	err := r.db.Where("asset_code = ?", code).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, err
}

// List 分页检索设备，支持科室/类型/状态/关键字多维过滤。
func (r *DeviceRepository) List(page, pageSize int, department, category, status, keyword string) ([]model.Device, int64, error) {
	var list []model.Device
	var total int64
	q := r.db.Model(&model.Device{})
	if department != "" {
		q = q.Where("department = ?", department)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		q = q.Where("name LIKE ? OR asset_code LIKE ? OR serial_number LIKE ? OR manufacturer LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// Update 更新设备。
func (r *DeviceRepository) Update(d *model.Device) error {
	return r.db.Save(d).Error
}

// UpdateTx 在指定事务中更新设备。
func (r *DeviceRepository) UpdateTx(tx *gorm.DB, d *model.Device) error {
	return tx.Save(d).Error
}

// Count 统计设备数量（被列表接口与统计接口复用）。
func (r *DeviceRepository) Count(status string) (int64, error) {
	var n int64
	q := r.db.Model(&model.Device{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Count(&n).Error
	return n, err
}

// SumAmount 统计资产总值。
func (r *DeviceRepository) SumAmount() (float64, error) {
	var sum float64
	err := r.db.Model(&model.Device{}).Select("COALESCE(SUM(purchase_amount),0)").Scan(&sum).Error
	return sum, err
}

// GroupCount 按字段分组统计。
func (r *DeviceRepository) GroupCount(field string) (map[string]int64, error) {
	type row struct {
		GroupKey string
		Count    int64
	}
	var rows []row
	err := r.db.Model(&model.Device{}).Select(field+" AS group_key, COUNT(*) AS count").Group(field).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, rw := range rows {
		out[rw.GroupKey] = rw.Count
	}
	return out, nil
}

// Delete 删除设备。
func (r *DeviceRepository) Delete(id uint) error {
	res := r.db.Delete(&model.Device{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateStatusTx 在事务中更新设备状态。
func (r *DeviceRepository) UpdateStatusTx(tx *gorm.DB, id uint, status string) error {
	res := tx.Model(&model.Device{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("update device status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
