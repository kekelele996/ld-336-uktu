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
	"gorm.io/gorm"
)

// ScrapService 设备报废服务。
type ScrapService struct {
	repo   *repository.ScrapRepository
	device *repository.DeviceRepository
	audit  *AuditService
	log    *slog.Logger
}

func NewScrapService(repo *repository.ScrapRepository, device *repository.DeviceRepository, audit *AuditService, log *slog.Logger) *ScrapService {
	return &ScrapService{repo: repo, device: device, audit: audit, log: log}
}

// Create 发起报废申请。
// 设备处于维修期间整次拒绝：不创建报废申请，设备与原维修工单保持原样。
func (s *ScrapService) Create(req *dto.CreateScrapReq, applicant string) (*model.ScrapRequest, error) {
	var sr *model.ScrapRequest
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// 行锁读取设备并完成已报废/维修中校验。
		d, err := lockDeviceAvailable(tx, req.DeviceID, s.log, "scrap")
		if err != nil {
			return err
		}
		sr = &model.ScrapRequest{
			ScrapNo:        util.GenSerial("SC"),
			DeviceID:       d.ID,
			DeviceName:     d.Name,
			Reason:         req.Reason,
			EstimatedValue: req.EstimatedValue,
			Status:         constants.ScrapStatusPending,
			Applicant:      applicant,
		}
		return s.repo.CreateTx(tx, sr)
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogScrapCreated, sr.ScrapNo, sr.DeviceID, sr.Status))
	s.audit.Record(0, applicant, "CREATE", "scrap", util.Uint64String(sr.ID), "发起报废: "+sr.ScrapNo, applicant, "")
	return sr, nil
}

// List 分页查询报废申请。
func (s *ScrapService) List(page, pageSize int, status string) (*util.PageResult, error) {
	list, total, err := s.repo.List(page, pageSize, status)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	return &util.PageResult{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// Approve 审批通过：设备状态变更为已报废并归档。
func (s *ScrapService) Approve(id uint, req *dto.ScrapApproveReq, operator string) (*model.ScrapRequest, error) {
	var updated *model.ScrapRequest
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		sr, err := s.repo.FindByIDForUpdateTx(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "报废申请不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		if sr.Status != constants.ScrapStatusPending {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		// 先锁设备行（与维修开始/完成保持一致的加锁顺序，避免死锁）；
		// 设备进入维修期间待审批报废不得通过（申请与设备保持原样）。
		d, err := lockDeviceAvailable(tx, sr.DeviceID, s.log, "scrap")
		if err != nil {
			return err
		}
		now := time.Now()
		sr.Status = constants.ScrapStatusApproved
		sr.Approver = operator
		sr.ApproveComment = req.Comment
		sr.ApproveAt = &now
		// CAS 更新：并发审批/状态变更时只有一个事务能推进申请。
		ok, err := s.repo.UpdateIfPendingTx(tx, sr)
		if err != nil {
			return err
		}
		if !ok {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		if err := s.device.UpdateStatusTx(tx, d.ID, constants.DeviceStatusScrapped); err != nil {
			return err
		}
		updated = sr
		return nil
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogScrapApproved, updated.ScrapNo, updated.DeviceID, updated.Status))
	s.audit.Record(0, operator, "APPROVE", "scrap", util.Uint64String(updated.ID), "报废通过并归档: "+updated.ScrapNo, operator, "")
	return updated, nil
}

// Reject 驳回报废申请。
func (s *ScrapService) Reject(id uint, req *dto.ScrapApproveReq, operator string) (*model.ScrapRequest, error) {
	sr, err := s.repo.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, util.NewAppError(http.StatusNotFound, "报废申请不存在: id="+util.Uint64String(id), nil)
	}
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	if sr.Status != constants.ScrapStatusPending {
		return nil, util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
	}
	sr.Status = constants.ScrapStatusRejected
	sr.Approver = operator
	sr.ApproveComment = req.Comment
	now := time.Now()
	sr.ApproveAt = &now
	if err := s.repo.Update(sr); err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "驳回报废失败: scrap_no="+sr.ScrapNo, err)
	}
	s.log.Info(fmt.Sprintf(constants.LogScrapRejected, sr.ScrapNo, operator, req.Comment))
	s.audit.Record(0, operator, "REJECT", "scrap", util.Uint64String(sr.ID), "报废驳回: "+sr.ScrapNo, operator, "")
	return sr, nil
}
