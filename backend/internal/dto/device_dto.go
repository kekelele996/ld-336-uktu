package dto

import (
	"time"

	"github.com/medasset/medasset/internal/model"
)

// CreateDeviceReq 创建设备请求。
type CreateDeviceReq struct {
	AssetCode           string     `json:"asset_code" binding:"required,max=64"`
	Barcode             string     `json:"barcode" binding:"omitempty,max=64"`
	Name                string     `json:"name" binding:"required,max=128"`
	Model               string     `json:"model" binding:"omitempty,max=128"`
	Manufacturer        string     `json:"manufacturer" binding:"omitempty,max=128"`
	SerialNumber        string     `json:"serial_number" binding:"omitempty,max=128"`
	Category            string     `json:"category" binding:"omitempty,max=64"`
	Department          string     `json:"department" binding:"omitempty,max=128"`
	ResponsiblePerson   string     `json:"responsible_person" binding:"omitempty,max=64"`
	Location            string     `json:"location" binding:"omitempty,max=128"`
	Supplier            string     `json:"supplier" binding:"omitempty,max=128"`
	PurchaseDate        *time.Time `json:"purchase_date"`
	PurchaseAmount      float64    `json:"purchase_amount" binding:"omitempty,min=0"`
	WarrantyMonths      int        `json:"warranty_months" binding:"omitempty,min=0"`
	RegistrationNo      string     `json:"registration_no" binding:"omitempty,max=128"`
	CertificateNo       string     `json:"certificate_no" binding:"omitempty,max=128"`
	Status              string     `json:"status" binding:"omitempty,oneof=in_storage in_use under_maintenance disabled scrapped"`
	CalibrationRequired bool       `json:"calibration_required"`
	PurchaseRequestID   uint       `json:"purchase_request_id"`
}

// UpdateDeviceReq 更新设备请求。
type UpdateDeviceReq struct {
	Name                string     `json:"name" binding:"required,max=128"`
	Model               string     `json:"model" binding:"omitempty,max=128"`
	Manufacturer        string     `json:"manufacturer" binding:"omitempty,max=128"`
	SerialNumber        string     `json:"serial_number" binding:"omitempty,max=128"`
	Category            string     `json:"category" binding:"omitempty,max=64"`
	Department          string     `json:"department" binding:"omitempty,max=128"`
	ResponsiblePerson   string     `json:"responsible_person" binding:"omitempty,max=64"`
	Location            string     `json:"location" binding:"omitempty,max=128"`
	Supplier            string     `json:"supplier" binding:"omitempty,max=128"`
	PurchaseAmount      float64    `json:"purchase_amount" binding:"omitempty,min=0"`
	WarrantyMonths      int        `json:"warranty_months" binding:"omitempty,min=0"`
	RegistrationNo      string     `json:"registration_no" binding:"omitempty,max=128"`
	CertificateNo       string     `json:"certificate_no" binding:"omitempty,max=128"`
	CalibrationRequired bool       `json:"calibration_required"`
}

// DeviceDetail 设备详情响应。
type DeviceDetail struct {
	model.Device
	WarrantyExpired bool `json:"warranty_expired"`
}
