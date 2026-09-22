package model

import "time"

// Device 医疗器械设备台账，覆盖采购验收后入库到报废的完整生命周期。
type Device struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	AssetCode           string    `gorm:"size:64;uniqueIndex;not null" json:"asset_code"`
	Barcode             string    `gorm:"size:64;index" json:"barcode"`
	Name                string    `gorm:"size:128;not null;index" json:"name"`
	Model               string    `gorm:"size:128" json:"model"`
	Manufacturer        string    `gorm:"size:128;index" json:"manufacturer"`
	SerialNumber        string    `gorm:"size:128" json:"serial_number"`
	Category            string    `gorm:"size:64;index" json:"category"`
	Department          string    `gorm:"size:128;index" json:"department"`
	ResponsiblePerson   string    `gorm:"size:64" json:"responsible_person"`
	Location            string    `gorm:"size:128" json:"location"`
	Supplier            string    `gorm:"size:128" json:"supplier"`
	PurchaseDate        *time.Time `json:"purchase_date"`
	PurchaseAmount      float64   `gorm:"type:decimal(14,2)" json:"purchase_amount"`
	WarrantyMonths      int       `json:"warranty_months"`
	WarrantyExpiry      *time.Time `json:"warranty_expiry"`
	RegistrationNo      string    `gorm:"size:128" json:"registration_no"`
	CertificateNo       string    `gorm:"size:128" json:"certificate_no"`
	Status              string    `gorm:"size:32;not null;default:in_storage;index" json:"status"`
	CalibrationRequired bool      `gorm:"default:false" json:"calibration_required"`
	LastMaintenanceAt   *time.Time `json:"last_maintenance_at"`
	PurchaseRequestID   uint      `gorm:"index" json:"purchase_request_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
