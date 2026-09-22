package model

import "time"

// TransferRequest 科室间设备调拨申请。
type TransferRequest struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TransferNo      string     `gorm:"size:64;uniqueIndex;not null" json:"transfer_no"`
	DeviceID        uint       `gorm:"not null;index" json:"device_id"`
	DeviceName      string     `gorm:"size:128;not null" json:"device_name"`
	FromDepartment  string     `gorm:"size:128;not null" json:"from_department"`
	ToDepartment    string     `gorm:"size:128;not null" json:"to_department"`
	FromPerson      string     `gorm:"size:64" json:"from_person"`
	ToPerson        string     `gorm:"size:64" json:"to_person"`
	Reason          string     `gorm:"size:512" json:"reason"`
	Status          string     `gorm:"size:32;not null;default:pending;index" json:"status"`
	Applicant       string     `gorm:"size:64" json:"applicant"`
	Approver        string     `gorm:"size:64" json:"approver"`
	ApproveComment  string     `gorm:"size:512" json:"approve_comment"`
	ApproveAt       *time.Time `json:"approve_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
