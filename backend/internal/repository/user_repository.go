package repository

import (
	"errors"
	"fmt"

	"github.com/medasset/medasset/internal/model"
	"gorm.io/gorm"
)

// UserRepository 用户仓储。
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// DB 返回底层数据库句柄（供 service 层开启事务）。
func (r *UserRepository) DB() *gorm.DB { return r.db }

// Create 创建用户。
func (r *UserRepository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

// FindByUsername 按用户名查询。
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

// FindByID 按 ID 查询。
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, err
}

// List 分页查询用户。
func (r *UserRepository) List(page, pageSize int, role, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	q := r.db.Model(&model.User{})
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if keyword != "" {
		q = q.Where("username LIKE ? OR real_name LIKE ? OR department LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// Update 更新用户（返回受影响行数）。
func (r *UserRepository) Update(u *model.User) error {
	return r.db.Save(u).Error
}

// Delete 删除用户。
func (r *UserRepository) Delete(id uint) error {
	res := r.db.Delete(&model.User{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CountByRole 按角色统计用户数。
func (r *UserRepository) CountByRole(role string) (int64, error) {
	var n int64
	err := r.db.Model(&model.User{}).Where("role = ?", role).Count(&n).Error
	return n, err
}

// IsUsernameTaken 判断Username是否已存在。
func (r *UserRepository) IsUsernameTaken(v string) (bool, error) {
	var n int64
	if err := r.db.Model(&model.User{}).Where("username = ?", v).Count(&n).Error; err != nil {
		return false, fmt.Errorf("check username: %w", err)
	}
	return n > 0, nil
}
