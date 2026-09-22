package dto

// OverviewResp 资产统计总览响应。
type OverviewResp struct {
	TotalDevices      int64            `json:"total_devices"`
	TotalAmount       float64          `json:"total_amount"`
	InUseDevices      int64            `json:"in_use_devices"`
	UnderMaintenance  int64            `json:"under_maintenance"`
	ScrappedDevices   int64            `json:"scrapped_devices"`
	MaintenanceCost   float64          `json:"maintenance_cost"`
	DepartmentDist    map[string]int64 `json:"department_dist"`
	ManufacturerDist  map[string]int64 `json:"manufacturer_dist"`
	CategoryDist      map[string]int64 `json:"category_dist"`
	CalibrationDue    int64            `json:"calibration_due"`
	PendingPurchases  int64            `json:"pending_purchases"`
}
