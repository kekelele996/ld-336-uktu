package service

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// wrapSvcErr 统一包装 service 层错误，保留 AppError 链。
func wrapSvcErr(err error) error {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
}

// lockDeviceAvailable 行锁读取设备并校验其可用于调拨/报废等操作：
// 已报废或维修中均拒绝（维修期间调拨/报废联动）。返回加锁后的设备，供调用方读取快照字段。
// 命中拒绝时返回带阻塞原因的 409，调用方不得创建/推进任何单据（原工单、设备与申请保持原样）。
func lockDeviceAvailable(tx *gorm.DB, deviceID uint, logger *slog.Logger, action string) (*model.Device, error) {
	var d model.Device
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&d, deviceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.NewAppError(http.StatusNotFound, "设备不存在: device_id="+util.Uint64String(deviceID), nil)
		}
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	if d.Status == constants.DeviceStatusScrapped {
		return nil, util.NewAppError(http.StatusConflict, constants.MsgDeviceInScrapped, nil)
	}
	if d.Status == constants.DeviceStatusUnderMaintenance {
		repairNo := ""
		var m model.MaintenanceRecord
		if err := tx.Where("device_id = ? AND type = ? AND status = ?",
			deviceID, constants.MaintenanceTypeRepair, constants.MaintenanceStatusInProgress).
			Order("id DESC").First(&m).Error; err == nil {
			repairNo = m.RecordNo
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
		}
		tpl := constants.LogTransferBlockedByMaintenance
		if action == "scrap" {
			tpl = constants.LogScrapBlockedByMaintenance
		}
		logger.Warn(fmt.Sprintf(tpl, deviceID, repairNo))
		return nil, util.NewAppError(http.StatusConflict,
			fmt.Sprintf("%s: device_id=%d in_progress_repair=%s", constants.MsgDeviceUnderMaintenance, deviceID, repairNo), nil)
	}
	return &d, nil
}
