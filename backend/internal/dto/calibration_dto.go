package dto

import "time"

// CreateCalibrationReq 建立计量台账请求。
type CreateCalibrationReq struct {
	InstrumentNo           string     `json:"instrument_no" binding:"required,max=64"`
	DeviceID               uint       `json:"device_id" binding:"required,min=1"`
	CalibrationCycleMonths int        `json:"calibration_cycle_months" binding:"required,min=1,max=120"`
	LastCalibrationDate    *time.Time `json:"last_calibration_date"`
	CertificateNo          string     `json:"certificate_no" binding:"omitempty,max=128"`
	CalibrationOrg         string     `json:"calibration_org" binding:"omitempty,max=128"`
	Remark                 string     `json:"remark" binding:"omitempty,max=512"`
}

// CalibrationResultReq 登记计量结果请求。
type CalibrationResultReq struct {
	Result               string     `json:"result" binding:"required,oneof=qualified unqualified"`
	NextCalibrationDate  *time.Time `json:"next_calibration_date"`
	CertificateNo        string     `json:"certificate_no" binding:"omitempty,max=128"`
	CalibrationOrg       string     `json:"calibration_org" binding:"omitempty,max=128"`
	Remark               string     `json:"remark" binding:"omitempty,max=512"`
}
