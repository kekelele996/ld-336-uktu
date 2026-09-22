package service

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/util"
	"github.com/medasset/medasset/pkg/pointerx"
	"gorm.io/gorm"
)

// MaintenanceService 维护保养与故障维修服务。
type MaintenanceService struct {
	repo        *repository.MaintenanceRepository
	device      *repository.DeviceRepository
	calibration *repository.CalibrationRepository
	audit       *AuditService
	log         *slog.Logger
}

func NewMaintenanceService(repo *repository.MaintenanceRepository, device *repository.DeviceRepository, calibration *repository.CalibrationRepository, audit *AuditService, log *slog.Logger) *MaintenanceService {
	return &MaintenanceService{repo: repo, device: device, calibration: calibration, audit: audit, log: log}
}

// Create 创建保养/维修工单（报修或计划执行）。
// 故障维修工单创建即做设备级互斥：同一设备已有待处理/处理中的维修工单时整次拒绝，
// 从源头保证每台设备至多一张未闭环维修单（开始时还会再次加锁校验以防并发）。
func (s *MaintenanceService) Create(req *dto.CreateMaintenanceReq, operator string) (*model.MaintenanceRecord, error) {
	m := &model.MaintenanceRecord{
		RecordNo:         util.GenSerial("MT"),
		Type:             req.Type,
		Status:           constants.MaintenanceStatusPending,
		PlannedDate:      req.PlannedDate,
		Content:          req.Content,
		FaultDescription: req.FaultDescription,
		Engineer:         req.Engineer,
		CreatedBy:        operator,
	}
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		d, err := s.device.FindByIDForUpdate(tx, req.DeviceID)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "设备不存在: device_id="+util.Uint64String(req.DeviceID), nil)
		}
		if err != nil {
			return util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
		}
		if d.Status == constants.DeviceStatusScrapped {
			return util.NewAppError(http.StatusConflict, constants.MsgDeviceInScrapped, nil)
		}
		if req.Type == constants.MaintenanceTypeRepair {
			blocker, err := s.repo.ActiveRepairForUpdateTx(tx, req.DeviceID, 0)
			if err != nil && !errors.Is(err, repository.ErrNotFound) {
				return err
			}
			if blocker != nil {
				s.log.Warn(fmt.Sprintf(constants.LogMaintenanceStartBlocked, req.DeviceID, blocker.RecordNo, blocker.Status))
				return util.NewAppError(http.StatusConflict,
					fmt.Sprintf("%s: device_id=%d blocked_by=%s blocker_status=%s",
						constants.MsgRepairBlocked, req.DeviceID, blocker.RecordNo, blocker.Status), nil)
			}
		}
		m.DeviceID = d.ID
		m.DeviceName = d.Name
		return s.repo.CreateTx(tx, m)
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogMaintenanceCreated, m.RecordNo, m.DeviceID, m.Type, m.Status))
	s.audit.Record(0, operator, "CREATE", "maintenance", util.Uint64String(m.ID), "创建保养/维修工单: "+m.RecordNo, operator, "")
	return m, nil
}

// GeneratePlans 根据设备类型自动生成保养计划（日检/周检/月检/年检，无待处理计划时生成）。
func (s *MaintenanceService) GeneratePlans(operator string) (int, error) {
	devices, _, err := s.device.List(1, 200, "", "", "", "")
	if err != nil {
		return 0, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	created := 0
	types := []string{constants.MaintenanceTypeDaily, constants.MaintenanceTypeWeekly, constants.MaintenanceTypeMonthly, constants.MaintenanceTypeYearly}
	for _, d := range devices {
		if d.Status == constants.DeviceStatusScrapped {
			continue
		}
		_ = s.repo.DB().Transaction(func(tx *gorm.DB) error {
			for _, t := range types {
				exists, err := s.existsPending(tx, d.ID, t)
				if err != nil {
					return err
				}
				if exists {
					continue
				}
				now := time.Now()
				m := &model.MaintenanceRecord{
					RecordNo:    util.GenSerial("MT"),
					DeviceID:    d.ID,
					DeviceName:  d.Name,
					Type:        t,
					Status:      constants.MaintenanceStatusPending,
					PlannedDate: planDate(now, t),
					Content:     "自动生成" + util.MaintenanceTypeText(t) + "保养计划",
					CreatedBy:   operator,
				}
				if err := tx.Create(m).Error; err != nil {
					return err
				}
				created++
			}
			return nil
		})
	}
	s.log.Info("自动生成保养计划完成", "created", created, "operator", operator)
	return created, nil
}

func (s *MaintenanceService) existsPending(tx *gorm.DB, deviceID uint, mType string) (bool, error) {
	var n int64
	err := tx.Model(&model.MaintenanceRecord{}).
		Where("device_id = ? AND type = ? AND status = ?", deviceID, mType, constants.MaintenanceStatusPending).
		Count(&n).Error
	return n > 0, err
}

// List 分页查询保养/维修记录。
func (s *MaintenanceService) List(page, pageSize int, deviceID uint, mType, status string) (*util.PageResult, error) {
	list, total, err := s.repo.List(page, pageSize, deviceID, mType, status)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	return &util.PageResult{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// Start 开始执行工单。
// 故障维修开始前进行设备级互斥：同一设备已存在其它待处理/处理中的维修工单时整次拒绝，
// 重复开始（工单已非待处理）同样拒绝。统一按"设备行→工单行"加锁，
// 与创建/完成/调拨/报废保持一致的加锁顺序，保证并发提交只有一次生效且无死锁。
func (s *MaintenanceService) Start(id uint, req *dto.StartMaintenanceReq, operator string) (*model.MaintenanceRecord, error) {
	var updated *model.MaintenanceRecord
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// 先快照读取工单拿到 device_id（不加锁），再按统一顺序锁设备行。
		pre, err := s.repo.FindByIDTx(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "工单不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		d, err := s.device.FindByIDForUpdate(tx, pre.DeviceID)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "设备不存在: device_id="+util.Uint64String(pre.DeviceID), nil)
		}
		if err != nil {
			return err
		}
		// 设备行锁之后再加锁工单，读取权威状态。
		m, err := s.repo.FindByIDForUpdate(tx, id)
		if err != nil {
			return err
		}
		if d.Status == constants.DeviceStatusScrapped {
			return util.NewAppError(http.StatusConflict, constants.MsgDeviceInScrapped, nil)
		}
		// 重复开始 / 并发重复提交：状态已被前一个请求推进，整次拒绝且不做任何变更。
		if m.Status != constants.MaintenanceStatusPending {
			return util.NewAppError(http.StatusConflict,
				constants.MsgInvalidStatus+": maintenance_record="+m.RecordNo+" current_status="+m.Status, nil)
		}
		// 维修类工单：同一设备只能有一个待处理/处理中的维修工单进入处理中。
		if m.Type == constants.MaintenanceTypeRepair {
			blocker, err := s.repo.ActiveRepairForUpdateTx(tx, m.DeviceID, m.ID)
			if err != nil && !errors.Is(err, repository.ErrNotFound) {
				return err
			}
			if blocker != nil {
				s.log.Warn(fmt.Sprintf(constants.LogMaintenanceStartBlocked, m.DeviceID, blocker.RecordNo, blocker.Status))
				return util.NewAppError(http.StatusConflict,
					fmt.Sprintf("%s: device_id=%d blocked_by=%s blocker_status=%s",
						constants.MsgRepairBlocked, m.DeviceID, blocker.RecordNo, blocker.Status), nil)
			}
		}
		m.Status = constants.MaintenanceStatusInProgress
		m.Engineer = req.Engineer
		if err := s.repo.UpdateTx(tx, m); err != nil {
			return err
		}
		// 维修类工单开始时设备进入维修中状态。
		if m.Type == constants.MaintenanceTypeRepair {
			if err := s.device.UpdateStatusTx(tx, m.DeviceID, constants.DeviceStatusUnderMaintenance); err != nil {
				return err
			}
		}
		updated = m
		return nil
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogMaintenanceStarted, updated.RecordNo, updated.Engineer, updated.Status))
	s.audit.Record(0, operator, "START", "maintenance", util.Uint64String(updated.ID), "开始执行: "+updated.RecordNo, operator, "")
	return updated, nil
}

// Complete 完成工单（更新工时/成本/配件，恢复设备状态）。
// 维修完成时依据最新计量结果决定设备可用性：最新计量不合格→保持禁用并提示；
// 计量合格或从未纳入计量→恢复使用中。
func (s *MaintenanceService) Complete(id uint, req *dto.CompleteMaintenanceReq, operator string) (*dto.CompleteMaintenanceResp, error) {
	var updated *model.MaintenanceRecord
	notice := ""
	deviceRestored := false
	deviceStatus := ""
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// 先快照读取工单拿到 device_id（不加锁），再按统一顺序锁设备行。
		pre, err := s.repo.FindByIDTx(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "工单不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		d, err := s.device.FindByIDForUpdate(tx, pre.DeviceID)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "设备不存在: device_id="+util.Uint64String(pre.DeviceID), nil)
		}
		if err != nil {
			return err
		}
		// 设备行锁之后再加锁工单，读取权威状态。
		m, err := s.repo.FindByIDForUpdate(tx, id)
		if err != nil {
			return err
		}
		if m.Status != constants.MaintenanceStatusInProgress {
			return util.NewAppError(http.StatusConflict,
				constants.MsgInvalidStatus+": maintenance_record="+m.RecordNo+" current_status="+m.Status, nil)
		}
		now := time.Now()
		m.Status = constants.MaintenanceStatusCompleted
		m.ExecutedDate = &now
		m.Content = req.Content
		m.ReplacedParts = req.ReplacedParts
		m.WorkHours = req.WorkHours
		m.Cost = req.Cost
		m.RepairResult = req.RepairResult
		if err := s.repo.UpdateTx(tx, m); err != nil {
			return err
		}
		// 维修完成后按最新计量结果恢复设备状态。
		if m.Type == constants.MaintenanceTypeRepair {
			targetStatus := constants.DeviceStatusInUse
			calib, calibErr := s.calibration.LatestByDeviceTx(tx, m.DeviceID)
			if calibErr != nil && !errors.Is(calibErr, repository.ErrNotFound) {
				return calibErr
			}
			if calib != nil && calib.Result == constants.CalibrationResultUnqualified {
				// 最新计量结果不合格：设备保持禁用。
				targetStatus = constants.DeviceStatusDisabled
				notice = constants.MsgMaintenanceUnqualified
				s.log.Info(fmt.Sprintf(constants.LogMaintenanceDeviceHeld, m.RecordNo, m.DeviceID, targetStatus, calib.InstrumentNo))
			}
			if d.Status != targetStatus {
				if err := s.device.UpdateStatusTx(tx, m.DeviceID, targetStatus); err != nil {
					return err
				}
			}
			deviceStatus = targetStatus
			deviceRestored = targetStatus == constants.DeviceStatusInUse
		}
		if err := tx.Model(&model.Device{}).Where("id = ?", m.DeviceID).Update("last_maintenance_at", now).Error; err != nil {
			return err
		}
		updated = m
		return nil
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogMaintenanceCompleted, updated.RecordNo, updated.DeviceID, updated.Cost, updated.Status))
	s.audit.Record(0, operator, "COMPLETE", "maintenance", util.Uint64String(updated.ID), "完成工单: "+updated.RecordNo, operator, "")
	return &dto.CompleteMaintenanceResp{
		Record:         updated,
		DeviceStatus:   deviceStatus,
		DeviceRestored: deviceRestored,
		Notice:         notice,
	}, nil
}

// Cancel 取消工单。
func (s *MaintenanceService) Cancel(id uint, req *dto.CancelMaintenanceReq, operator string) (*model.MaintenanceRecord, error) {
	var updated *model.MaintenanceRecord
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		m, err := s.repo.FindByIDForUpdate(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "工单不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		if m.Status != constants.MaintenanceStatusPending {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		m.Status = constants.MaintenanceStatusCancelled
		if req.Reason != "" {
			m.RepairResult = "取消原因: " + req.Reason
		}
		if err := s.repo.UpdateTx(tx, m); err != nil {
			return err
		}
		updated = m
		return nil
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogMaintenanceCancelled, updated.RecordNo, req.Reason, updated.Status))
	s.audit.Record(0, operator, "CANCEL", "maintenance", util.Uint64String(updated.ID), "取消工单: "+updated.RecordNo, operator, "")
	return updated, nil
}

func planDate(now time.Time, mType string) *time.Time {
	var add time.Duration
	switch mType {
	case constants.MaintenanceTypeDaily:
		add = 24 * time.Hour
	case constants.MaintenanceTypeWeekly:
		add = 7 * 24 * time.Hour
	case constants.MaintenanceTypeMonthly:
		add = 30 * 24 * time.Hour
	case constants.MaintenanceTypeYearly:
		add = 365 * 24 * time.Hour
	default:
		add = 30 * 24 * time.Hour
	}
	t := now.Add(add)
	return pointerx.TimePtr(t)
}
