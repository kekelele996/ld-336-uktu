package dto

// CreateScrapReq 创建报废申请请求。
type CreateScrapReq struct {
	DeviceID       uint    `json:"device_id" binding:"required,min=1"`
	Reason         string  `json:"reason" binding:"required,max=512"`
	EstimatedValue float64 `json:"estimated_value" binding:"omitempty,min=0"`
}

// ScrapApproveReq 报废审批请求。
type ScrapApproveReq struct {
	Comment string `json:"comment" binding:"omitempty,max=512"`
}
