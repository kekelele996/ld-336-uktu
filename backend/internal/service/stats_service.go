package service

import (
	"log/slog"
	"net/http"

	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/util"
)

// StatsService 资产统计与合规报表服务。
type StatsService struct {
	device       *repository.DeviceRepository
	maintenance  *repository.MaintenanceRepository
	calibration  *repository.CalibrationRepository
	purchase     *repository.PurchaseRepository
	audit        *AuditService
	log          *slog.Logger
}

func NewStatsService(device *repository.DeviceRepository, maintenance *repository.MaintenanceRepository,
	calibration *repository.CalibrationRepository, purchase *repository.PurchaseRepository,
	audit *AuditService, log *slog.Logger) *StatsService {
	return &StatsService{device: device, maintenance: maintenance, calibration: calibration,
		purchase: purchase, audit: audit, log: log}
}

// Overview 全院设备资产总览（复用各仓储统计方法）。
func (s *StatsService) Overview() (*dto.OverviewResp, error) {
	total, err := s.device.Count("")
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	totalAmount, err := s.device.SumAmount()
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	inUse, err := s.device.Count(constants.DeviceStatusInUse)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	underMaint, err := s.device.Count(constants.DeviceStatusUnderMaintenance)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	scrapped, err := s.device.Count(constants.DeviceStatusScrapped)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	maintCost, err := s.maintenance.SumCost()
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	departmentDist, err := s.device.GroupCount("department")
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	manufacturerDist, err := s.device.GroupCount("manufacturer")
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	categoryDist, err := s.device.GroupCount("category")
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	calibDue, err := s.calibration.CountByStatus(constants.CalibrationStatusDue)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	calibExpired, err := s.calibration.CountByStatus(constants.CalibrationStatusExpired)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	pendingPurchases, err := s.purchaseCountPending()
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	s.audit.Record(0, "system", "VIEW", "stats", "overview", "查看资产统计总览", "system", "")
	return &dto.OverviewResp{
		TotalDevices:     total,
		TotalAmount:      totalAmount,
		InUseDevices:     inUse,
		UnderMaintenance: underMaint,
		ScrappedDevices:  scrapped,
		MaintenanceCost:  maintCost,
		DepartmentDist:   departmentDist,
		ManufacturerDist: manufacturerDist,
		CategoryDist:     categoryDist,
		CalibrationDue:   calibDue + calibExpired,
		PendingPurchases: pendingPurchases,
	}, nil
}

// purchaseCountPending 统计待处理采购申请数。
func (s *StatsService) purchaseCountPending() (int64, error) {
	_, total, err := s.purchase.List(1, 1, constants.PurchaseStatusPendingDeviceAdmin, "")
	if err != nil {
		return 0, err
	}
	_, total2, err := s.purchase.List(1, 1, constants.PurchaseStatusPendingDean, "")
	if err != nil {
		return 0, err
	}
	return total + total2, nil
}
