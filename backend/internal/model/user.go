package model

import "time"

// User 系统用户：覆盖设备科、院长、科室、工程师等角色。
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password     string    `gorm:"size:255;not null" json:"-"`
	RealName     string    `gorm:"size:64;not null" json:"real_name"`
	Role         string    `gorm:"size:32;not null;index" json:"role"`
	Department   string    `gorm:"size:128" json:"department"`
	Phone        string    `gorm:"size:32" json:"phone"`
	Email        string    `gorm:"size:128" json:"email"`
	Status       string    `gorm:"size:16;not null;default:active" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
