package dto

import (
	"time"

	"github.com/medasset/medasset/internal/model"
)

// RegisterReq 注册请求。
type RegisterReq struct {
	Username   string `json:"username" binding:"required,min=3,max=64"`
	Password   string `json:"password" binding:"required,min=6,max=64"`
	RealName   string `json:"real_name" binding:"required,max=64"`
	Role       string `json:"role" binding:"omitempty,oneof=SUPER_ADMIN DEVICE_ADMIN DEAN DEPARTMENT ENGINEER"`
	Department string `json:"department" binding:"omitempty,max=128"`
	Phone      string `json:"phone" binding:"omitempty,max=32"`
	Email      string `json:"email" binding:"omitempty,email,max=128"`
}

// LoginReq 登录请求。
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResp 登录响应。
type LoginResp struct {
	Token string     `json:"token"`
	User  *model.User `json:"user"`
}

// UpdateUserReq 更新用户请求。
type UpdateUserReq struct {
	RealName   string `json:"real_name" binding:"required,max=64"`
	Role       string `json:"role" binding:"required,oneof=SUPER_ADMIN DEVICE_ADMIN DEAN DEPARTMENT ENGINEER"`
	Department string `json:"department" binding:"omitempty,max=128"`
	Phone      string `json:"phone" binding:"omitempty,max=32"`
	Email      string `json:"email" binding:"omitempty,email,max=128"`
	Status     string `json:"status" binding:"omitempty,oneof=active disabled"`
}

// UserListItem 用户列表项。
type UserListItem struct {
	ID         uint      `json:"id"`
	Username   string    `json:"username"`
	RealName   string    `json:"real_name"`
	Role       string    `json:"role"`
	Department string    `json:"department"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// ToUserListItem 转换用户实体为列表项。
func ToUserListItem(u *model.User) UserListItem {
	return UserListItem{
		ID:         u.ID,
		Username:   u.Username,
		RealName:   u.RealName,
		Role:       u.Role,
		Department: u.Department,
		Phone:      u.Phone,
		Email:      u.Email,
		Status:     u.Status,
		CreatedAt:  u.CreatedAt,
	}
}
