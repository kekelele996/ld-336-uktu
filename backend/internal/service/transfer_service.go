package service

import (
	"fmt"
	"errors"
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

// TransferService 设备调拨服务。
type TransferService struct {
	repo   *repository.TransferRepository
	device *repository.DeviceRepository
	audit  *AuditService
	log    *slog.Logger
}

func NewTransferService(repo *repository.TransferRepository, device *repository.DeviceRepository, audit *AuditService, log *slog.Logger) *TransferService {
	return &TransferService{repo: repo, device: device, audit: audit, log: log}
}

// Create 发起调拨申请。
func (s *TransferService) Create(req *dto.CreateTransferReq, applicant string) (*model.TransferRequest, error) {
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
	t := &model.TransferRequest{
		TransferNo:     util.GenSerial("TR"),
		DeviceID:       d.ID,
		DeviceName:     d.Name,
		FromDepartment: d.Department,
		ToDepartment:   req.ToDepartment,
		FromPerson:     d.ResponsiblePerson,
		ToPerson:       req.ToPerson,
		Reason:         req.Reason,
		Status:         constants.TransferStatusPending,
		Applicant:      applicant,
	}
	if err := s.repo.Create(t); err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "创建调拨申请失败: device_name="+d.Name, err)
	}
	s.log.Info(fmt.Sprintf(constants.LogTransferCreated, t.TransferNo, t.DeviceID, t.FromDepartment, t.ToDepartment, t.Status))
	s.audit.Record(0, applicant, "CREATE", "transfer", util.Uint64String(t.ID), "发起调拨: "+t.TransferNo, applicant, "")
	return t, nil
}

// List 分页查询调拨申请。
func (s *TransferService) List(page, pageSize int, status string) (*util.PageResult, error) {
	list, total, err := s.repo.List(page, pageSize, status)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	return &util.PageResult{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// Approve 审批通过：自动更新设备所属科室和责任人。
func (s *TransferService) Approve(id uint, req *dto.TransferApproveReq, operator string) (*model.TransferRequest, error) {
	var updated *model.TransferRequest
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		t, err := s.repo.FindByID(id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "调拨申请不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		if t.Status != constants.TransferStatusPending {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		t.Status = constants.TransferStatusApproved
		t.Approver = operator
		t.ApproveComment = req.Comment
		now := time.Now()
		t.ApproveAt = &now
		if err := s.repo.UpdateTx(tx, t); err != nil {
			return err
		}
		// 更新设备所属科室与责任人。
		if err := tx.Model(&model.Device{}).Where("id = ?", t.DeviceID).
			Updates(map[string]any{"department": t.ToDepartment, "responsible_person": t.ToPerson}).Error; err != nil {
			return err
		}
		updated = t
		return nil
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogTransferApproved, updated.TransferNo, updated.DeviceID, updated.ToDepartment))
	s.audit.Record(0, operator, "APPROVE", "transfer", util.Uint64String(updated.ID), "调拨通过: "+updated.TransferNo, operator, "")
	return updated, nil
}

// Reject 驳回调拨申请。
func (s *TransferService) Reject(id uint, req *dto.TransferApproveReq, operator string) (*model.TransferRequest, error) {
	t, err := s.repo.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, util.NewAppError(http.StatusNotFound, "调拨申请不存在: id="+util.Uint64String(id), nil)
	}
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	if t.Status != constants.TransferStatusPending {
		return nil, util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
	}
	t.Status = constants.TransferStatusRejected
	t.Approver = operator
	t.ApproveComment = req.Comment
	now := time.Now()
	t.ApproveAt = &now
	if err := s.repo.Update(t); err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "驳回调拨失败: transfer_no="+t.TransferNo, err)
	}
	s.log.Info(fmt.Sprintf(constants.LogTransferRejected, t.TransferNo, operator, req.Comment))
	s.audit.Record(0, operator, "REJECT", "transfer", util.Uint64String(t.ID), "调拨驳回: "+t.TransferNo, operator, "")
	return t, nil
}
