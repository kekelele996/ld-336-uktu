package util

import (
	"fmt"
	"time"

	"github.com/medasset/medasset/internal/constants"
)

// 日期、状态文本、类型文本等格式化逻辑集中于此（多处耦合）。
const (
	dateLayout = "2006-01-02"
	dateTimeLayout = "2006-01-02 15:04:05"
)

// FormatDate 格式化日期为 yyyy-MM-dd。
func FormatDate(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format(dateLayout)
}

// FormatDateTime 格式化日期时间为 yyyy-MM-dd HH:mm:ss。
func FormatDateTime(t time.Time) string {
	return t.Format(dateTimeLayout)
}

// FormatAmount 格式化金额，保留两位小数。
func FormatAmount(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

// DeviceStatusText 设备状态展示文本。
func DeviceStatusText(status string) string {
	switch status {
	case constants.DeviceStatusInStorage:
		return "在库"
	case constants.DeviceStatusInUse:
		return "使用中"
	case constants.DeviceStatusUnderMaintenance:
		return "维修中"
	case constants.DeviceStatusDisabled:
		return "已禁用"
	case constants.DeviceStatusScrapped:
		return "已报废"
	default:
		return status
	}
}

// PurchaseStatusText 采购状态展示文本。
func PurchaseStatusText(status string) string {
	switch status {
	case constants.PurchaseStatusPendingDeviceAdmin:
		return "待设备科审核"
	case constants.PurchaseStatusPendingDean:
		return "待院长审批"
	case constants.PurchaseStatusApproved:
		return "审批通过"
	case constants.PurchaseStatusDelivered:
		return "已到货待验收"
	case constants.PurchaseStatusAccepted:
		return "已验收"
	case constants.PurchaseStatusRejected:
		return "已拒绝"
	default:
		return status
	}
}

// MaintenanceTypeText 保养/维修类型展示文本。
func MaintenanceTypeText(mType string) string {
	switch mType {
	case constants.MaintenanceTypeDaily:
		return "日检"
	case constants.MaintenanceTypeWeekly:
		return "周检"
	case constants.MaintenanceTypeMonthly:
		return "月检"
	case constants.MaintenanceTypeYearly:
		return "年检"
	case constants.MaintenanceTypeRepair:
		return "故障维修"
	default:
		return mType
	}
}

// MaintenanceStatusText 保养/维修状态展示文本。
func MaintenanceStatusText(status string) string {
	switch status {
	case constants.MaintenanceStatusPending:
		return "待处理"
	case constants.MaintenanceStatusInProgress:
		return "处理中"
	case constants.MaintenanceStatusCompleted:
		return "已完成"
	case constants.MaintenanceStatusCancelled:
		return "已取消"
	default:
		return status
	}
}

// CalibrationStatusText 计量状态展示文本。
func CalibrationStatusText(status string) string {
	switch status {
	case constants.CalibrationStatusNormal:
		return "合格"
	case constants.CalibrationStatusUnqualified:
		return "不合格"
	case constants.CalibrationStatusDue:
		return "即将到期"
	case constants.CalibrationStatusExpired:
		return "已过期"
	default:
		return status
	}
}

// TransferStatusText 调拨状态展示文本。
func TransferStatusText(status string) string {
	switch status {
	case constants.TransferStatusPending:
		return "待审批"
	case constants.TransferStatusApproved:
		return "已批准"
	case constants.TransferStatusRejected:
		return "已驳回"
	default:
		return status
	}
}

// ScrapStatusText 报废状态展示文本。
func ScrapStatusText(status string) string {
	switch status {
	case constants.ScrapStatusPending:
		return "待审批"
	case constants.ScrapStatusApproved:
		return "已批准"
	case constants.ScrapStatusRejected:
		return "已驳回"
	default:
		return status
	}
}
