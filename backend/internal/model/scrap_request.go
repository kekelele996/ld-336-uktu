package model

import "time"

// ScrapRequest 设备报废申请。
type ScrapRequest struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ScrapNo        string     `gorm:"size:64;uniqueIndex;not null" json:"scrap_no"`
	DeviceID       uint       `gorm:"not null;index" json:"device_id"`
	DeviceName     string     `gorm:"size:128;not null" json:"device_name"`
	Reason         string     `gorm:"size:512;not null" json:"reason"`
	EstimatedValue float64    `gorm:"type:decimal(14,2)" json:"estimated_value"`
	Status         string     `gorm:"size:32;not null;default:pending;index" json:"status"`
	Applicant      string     `gorm:"size:64" json:"applicant"`
	Approver       string     `gorm:"size:64" json:"approver"`
	ApproveComment string     `gorm:"size:512" json:"approve_comment"`
	ApproveAt      *time.Time `json:"approve_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
