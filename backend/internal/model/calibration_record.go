package model

import "time"

// CalibrationRecord 计量器具台账与计量结果记录。
type CalibrationRecord struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	InstrumentNo          string     `gorm:"size:64;uniqueIndex;not null" json:"instrument_no"`
	DeviceID              uint       `gorm:"not null;index" json:"device_id"`
	DeviceName            string     `gorm:"size:128;not null" json:"device_name"`
	CalibrationCycleMonths int       `json:"calibration_cycle_months"`
	LastCalibrationDate   *time.Time `json:"last_calibration_date"`
	NextCalibrationDate   *time.Time `gorm:"index" json:"next_calibration_date"`
	Status                string     `gorm:"size:32;not null;default:normal;index" json:"status"`
	Result                string     `gorm:"size:16" json:"result"`
	CertificateNo         string     `gorm:"size:128" json:"certificate_no"`
	CalibrationOrg        string     `gorm:"size:128" json:"calibration_org"`
	Remark                string     `gorm:"size:512" json:"remark"`
	CreatedBy             string     `gorm:"size:64" json:"created_by"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
