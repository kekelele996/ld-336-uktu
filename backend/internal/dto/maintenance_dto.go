package dto

import (
	"time"

	"github.com/medasset/medasset/internal/model"
)

// CreateMaintenanceReq 创建保养/维修工单请求。
type CreateMaintenanceReq struct {
	DeviceID         uint       `json:"device_id" binding:"required,min=1"`
	Type             string     `json:"type" binding:"required,oneof=daily weekly monthly yearly repair"`
	PlannedDate      *time.Time `json:"planned_date"`
	Content          string     `json:"content" binding:"omitempty,max=1024"`
	FaultDescription string     `json:"fault_description" binding:"omitempty,max=1024"`
	Engineer         string     `json:"engineer" binding:"omitempty,max=64"`
}

// StartMaintenanceReq 开始执行请求。
type StartMaintenanceReq struct {
	Engineer string `json:"engineer" binding:"required,max=64"`
}

// CompleteMaintenanceReq 完成工单请求。
type CompleteMaintenanceReq struct {
	Content       string  `json:"content" binding:"required,max=1024"`
	ReplacedParts string  `json:"replaced_parts" binding:"omitempty,max=1024"`
	WorkHours     float64 `json:"work_hours" binding:"omitempty,min=0"`
	Cost          float64 `json:"cost" binding:"omitempty,min=0"`
	RepairResult  string  `json:"repair_result" binding:"omitempty,max=1024"`
}

// CancelMaintenanceReq 取消工单请求。
type CancelMaintenanceReq struct {
	Reason string `json:"reason" binding:"omitempty,max=512"`
}

// CompleteMaintenanceResp 完成工单响应。
// DeviceRestored=false 时表示设备未恢复使用（如最新计量不合格保持禁用），Notice 为页面提示。
type CompleteMaintenanceResp struct {
	Record         *model.MaintenanceRecord `json:"record"`
	DeviceStatus   string                   `json:"device_status"`
	DeviceRestored bool                     `json:"device_restored"`
	Notice         string                   `json:"notice,omitempty"`
}
