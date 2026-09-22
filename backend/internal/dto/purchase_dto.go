package dto

import "time"

// CreatePurchaseReq 创建采购申请请求。
type CreatePurchaseReq struct {
	Department    string  `json:"department" binding:"required,max=128"`
	ApplicantName string  `json:"applicant_name" binding:"required,max=64"`
	DeviceName    string  `json:"device_name" binding:"required,max=128"`
	Model         string  `json:"model" binding:"omitempty,max=128"`
	Manufacturer  string  `json:"manufacturer" binding:"omitempty,max=128"`
	Quantity      int     `json:"quantity" binding:"required,min=1,max=999"`
	BudgetAmount  float64 `json:"budget_amount" binding:"omitempty,min=0"`
	Reason        string  `json:"reason" binding:"omitempty,max=512"`
}

// ApproveReq 审批请求。
type ApproveReq struct {
	Comment string `json:"comment" binding:"omitempty,max=512"`
}

// AcceptReq 验收登记请求。
type AcceptReq struct {
	AcceptancePerson string     `json:"acceptance_person" binding:"required,max=64"`
	AcceptanceDate   *time.Time `json:"acceptance_date"`
	PartsList        string     `json:"parts_list" binding:"omitempty,max=1024"`
	CertificateNo    string     `json:"certificate_no" binding:"omitempty,max=128"`
	RegistrationNo   string     `json:"registration_no" binding:"omitempty,max=128"`
	AssetCode        string     `json:"asset_code" binding:"required,max=64"`
	Category         string     `json:"category" binding:"omitempty,max=64"`
	ResponsiblePerson string    `json:"responsible_person" binding:"omitempty,max=64"`
	Location         string     `json:"location" binding:"omitempty,max=128"`
	WarrantyMonths   int        `json:"warranty_months" binding:"omitempty,min=0"`
	CalibrationRequired bool    `json:"calibration_required"`
}
