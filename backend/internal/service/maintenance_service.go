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
func (s *MaintenanceService) Create(req *dto.CreateMaintenanceReq, operator string) (*model.MaintenanceRecord, error) {
	d, err := s.device.FindByID(req.DeviceID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, util.NewAppError(http.StatusNotFound, "设备不存在: device_id="+util.Uint64String(req.DeviceID), nil)
	}
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	if d.Status == constants.DeviceStatusScrapped {
		return nil, util.NewAppError(http.StatusConflict, constants.MsgDeviceInScrapped, nil)
	}
	m := &model.MaintenanceRecord{
		RecordNo:         util.GenSerial("MT"),
		DeviceID:         d.ID,
		DeviceName:       d.Name,
		Type:             req.Type,
		Status:           constants.MaintenanceStatusPending,
		PlannedDate:      req.PlannedDate,
		Content:          req.Content,
		FaultDescription: req.FaultDescription,
		Engineer:         req.Engineer,
		CreatedBy:        operator,
	}
	if err := s.repo.Create(m); err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "创建工单失败: device_name="+d.Name, err)
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
// 维修工单开始前做设备级互斥：同一设备存在其他待处理/处理中的维修工单时整次拒绝
// （含本工单之外的待处理重复报修单），设备行锁串行化并发提交，重复/并发开始只生效一次。
func (s *MaintenanceService) Start(id uint, req *dto.StartMaintenanceReq, operator string) (*dto.MaintenanceActionResp, error) {
	var resp *dto.MaintenanceActionResp
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// 先普通读取工单定位设备，随后锁定设备行再锁工单行：
		// 同一设备的并发开始在设备锁处串行，避免“工单行→设备行”反向加锁导致死锁。
		pre, err := s.repo.FindByIDTx(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "工单不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		d, err := s.device.FindByIDForUpdate(tx, pre.DeviceID)
		if err != nil {
			return err
		}
		m, err := s.repo.FindByIDForUpdate(tx, id)
		if err != nil {
			return err
		}
		if m.Status != constants.MaintenanceStatusPending {
			// 重复开始：仅给出阻塞原因，不改变任何数据。
			if m.Type == constants.MaintenanceTypeRepair && m.Status == constants.MaintenanceStatusInProgress {
				return util.NewAppError(http.StatusConflict, constants.MsgRepairBlockedPrefix+constants.MsgRepairAlreadyRunning, nil)
			}
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		if m.Type == constants.MaintenanceTypeRepair {
			blocker, err := s.repo.FindRepairBlockerForUpdate(tx, m.DeviceID, m.ID)
			if err != nil {
				return err
			}
			if blocker != nil {
				reason := fmt.Sprintf("%s（阻塞工单: %s，状态: %s）",
					constants.MsgRepairBlockedByOther, blocker.RecordNo, util.MaintenanceStatusText(blocker.Status))
				s.log.Warn(fmt.Sprintf(constants.LogMaintenanceBlocked, m.RecordNo, m.DeviceID, blocker.RecordNo, blocker.Status, operator))
				return util.NewAppError(http.StatusConflict, constants.MsgRepairBlockedPrefix+reason, nil)
			}
		}
		m.Status = constants.MaintenanceStatusInProgress
		m.Engineer = req.Engineer
		if err := s.repo.UpdateTx(tx, m); err != nil {
			return err
		}
		notice := ""
		// 维修类工单开始时设备进入维修中状态；维修期间调拨/报废申请将被拒绝。
		if m.Type == constants.MaintenanceTypeRepair {
			if err := s.device.UpdateStatusTx(tx, m.DeviceID, constants.DeviceStatusUnderMaintenance); err != nil {
				return err
			}
			d.Status = constants.DeviceStatusUnderMaintenance
			notice = constants.MsgRepairStartNotice
		}
		resp = &dto.MaintenanceActionResp{MaintenanceRecord: *m, DeviceStatus: d.Status, Notice: notice}
		return nil
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogMaintenanceStarted, resp.RecordNo, resp.Engineer, resp.Status))
	s.audit.Record(0, operator, "START", "maintenance", util.Uint64String(resp.ID), "开始执行: "+resp.RecordNo, operator, "")
	return resp, nil
}

// Complete 完成工单（更新工时/成本/配件，恢复设备状态）。
// 维修完成时按最新计量结果联动设备可用性：最新结果不合格则保持禁用并提示未通过计量；
// 仅计量合格或设备从未纳入计量时恢复使用中。
func (s *MaintenanceService) Complete(id uint, req *dto.CompleteMaintenanceReq, operator string) (*dto.MaintenanceActionResp, error) {
	var resp *dto.MaintenanceActionResp
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// 与 Start 保持相同加锁顺序（设备行先于工单行），杜绝跨事务死锁。
		pre, err := s.repo.FindByIDTx(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "工单不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		d, err := s.device.FindByIDForUpdate(tx, pre.DeviceID)
		if err != nil {
			return err
		}
		m, err := s.repo.FindByIDForUpdate(tx, id)
		if err != nil {
			return err
		}
		if m.Status != constants.MaintenanceStatusInProgress {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
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
		notice := ""
		latestResult := ""
		if m.Type == constants.MaintenanceTypeRepair {
			targetStatus := constants.DeviceStatusInUse
			notice = constants.MsgRepairRestoredInUse
			latest, err := s.calibration.LatestByDeviceTx(tx, m.DeviceID)
			if err != nil {
				return err
			}
			if latest != nil {
				latestResult = latest.Result
				if latest.Result == constants.CalibrationResultUnqualified {
					// 最新计量结果不合格：设备保持禁用，提示未通过计量。
					targetStatus = constants.DeviceStatusDisabled
					notice = constants.MsgRepairCalibrationUnqualified
				}
			}
			if err := s.device.UpdateStatusTx(tx, m.DeviceID, targetStatus); err != nil {
				return err
			}
			d.Status = targetStatus
			s.log.Info(fmt.Sprintf(constants.LogMaintenanceCalibration, m.RecordNo, m.DeviceID, latestResult, targetStatus))
		}
		if err := tx.Model(&model.Device{}).Where("id = ?", m.DeviceID).Update("last_maintenance_at", now).Error; err != nil {
			return err
		}
		resp = &dto.MaintenanceActionResp{MaintenanceRecord: *m, DeviceStatus: d.Status, Notice: notice}
		return nil
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogMaintenanceCompleted, resp.RecordNo, resp.DeviceID, resp.Cost, resp.Status))
	auditDetail := "完成工单: " + resp.RecordNo
	if resp.Notice == constants.MsgRepairCalibrationUnqualified {
		auditDetail += "；最新计量不合格，设备保持禁用"
	}
	s.audit.Record(0, operator, "COMPLETE", "maintenance", util.Uint64String(resp.ID), auditDetail, operator, "")
	return resp, nil
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
