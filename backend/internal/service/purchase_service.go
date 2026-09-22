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
	"github.com/medasset/medasset/pkg/pointerx"
	"gorm.io/gorm"
)

// PurchaseService 采购与验收流程服务。
type PurchaseService struct {
	repo   *repository.PurchaseRepository
	device *repository.DeviceRepository
	audit  *AuditService
	log    *slog.Logger
}

func NewPurchaseService(repo *repository.PurchaseRepository, device *repository.DeviceRepository, audit *AuditService, log *slog.Logger) *PurchaseService {
	return &PurchaseService{repo: repo, device: device, audit: audit, log: log}
}

// Create 提交采购申请（科室申请）。
func (s *PurchaseService) Create(req *dto.CreatePurchaseReq, applicant string) (*model.PurchaseRequest, error) {
	p := &model.PurchaseRequest{
		RequestNo:     util.GenSerial("PR"),
		Department:    req.Department,
		ApplicantName: req.ApplicantName,
		DeviceName:    req.DeviceName,
		Model:         req.Model,
		Manufacturer:  req.Manufacturer,
		Quantity:      req.Quantity,
		BudgetAmount:  req.BudgetAmount,
		Reason:        req.Reason,
		Status:        constants.PurchaseStatusPendingDeviceAdmin,
		SubmittedBy:   applicant,
	}
	p.SubmittedAt = pointerx.TimePtr(time.Now())
	if err := s.repo.Create(p); err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "创建采购申请失败: device_name="+req.DeviceName, err)
	}
	s.log.Info(fmt.Sprintf(constants.LogPurchaseCreated, p.RequestNo, p.Department, p.DeviceName, p.Status))
	s.audit.Record(0, applicant, "CREATE", "purchase", util.Uint64String(p.ID), "提交采购申请: "+p.DeviceName, applicant, "")
	return p, nil
}

// List 分页查询采购申请。
func (s *PurchaseService) List(page, pageSize int, status, department string) (*util.PageResult, error) {
	list, total, err := s.repo.List(page, pageSize, status, department)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	return &util.PageResult{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// Get 查询采购申请详情。
func (s *PurchaseService) Get(id uint) (*model.PurchaseRequest, error) {
	p, err := s.repo.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, util.NewAppError(http.StatusNotFound, "采购申请不存在: id="+util.Uint64String(id), nil)
	}
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	return p, nil
}

// DeviceAdminApprove 设备科审核通过 → 待院长审批。
func (s *PurchaseService) DeviceAdminApprove(id uint, req *dto.ApproveReq, operator string) (*model.PurchaseRequest, error) {
	var updated *model.PurchaseRequest
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		p, err := s.repo.FindByIDForUpdate(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "采购申请不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		if p.Status != constants.PurchaseStatusPendingDeviceAdmin {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		p.Status = constants.PurchaseStatusPendingDean
		p.DeviceAdmin = operator
		p.DeviceAdminComment = req.Comment
		now := time.Now()
		p.DeviceAdminAt = &now
		if err := s.repo.UpdateTx(tx, p); err != nil {
			return err
		}
		updated = p
		return nil
	})
	if err != nil {
		return nil, s.wrapStatus(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogPurchaseAdminApprove, updated.RequestNo, operator, updated.Status))
	s.audit.Record(0, operator, "APPROVE", "purchase", util.Uint64String(updated.ID), "设备科审核通过: "+updated.RequestNo, operator, "")
	return updated, nil
}

// DeviceAdminReject 设备科驳回。
func (s *PurchaseService) DeviceAdminReject(id uint, req *dto.ApproveReq, operator string) (*model.PurchaseRequest, error) {
	var updated *model.PurchaseRequest
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		p, err := s.repo.FindByIDForUpdate(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "采购申请不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		if p.Status != constants.PurchaseStatusPendingDeviceAdmin {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		p.Status = constants.PurchaseStatusRejected
		p.DeviceAdmin = operator
		p.DeviceAdminComment = req.Comment
		now := time.Now()
		p.DeviceAdminAt = &now
		if err := s.repo.UpdateTx(tx, p); err != nil {
			return err
		}
		updated = p
		return nil
	})
	if err != nil {
		return nil, s.wrapStatus(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogPurchaseAdminReject, updated.RequestNo, operator, req.Comment))
	s.audit.Record(0, operator, "REJECT", "purchase", util.Uint64String(updated.ID), "设备科驳回: "+updated.RequestNo, operator, "")
	return updated, nil
}

// DeanApprove 院长审批通过 → 审批通过。
func (s *PurchaseService) DeanApprove(id uint, req *dto.ApproveReq, operator string) (*model.PurchaseRequest, error) {
	var updated *model.PurchaseRequest
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		p, err := s.repo.FindByIDForUpdate(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "采购申请不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		if p.Status != constants.PurchaseStatusPendingDean {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		p.Status = constants.PurchaseStatusApproved
		p.Dean = operator
		p.DeanComment = req.Comment
		now := time.Now()
		p.DeanAt = &now
		if err := s.repo.UpdateTx(tx, p); err != nil {
			return err
		}
		updated = p
		return nil
	})
	if err != nil {
		return nil, s.wrapStatus(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogPurchaseDeanApprove, updated.RequestNo, operator, updated.Status))
	s.audit.Record(0, operator, "APPROVE", "purchase", util.Uint64String(updated.ID), "院长审批通过: "+updated.RequestNo, operator, "")
	return updated, nil
}

// DeanReject 院长驳回。
func (s *PurchaseService) DeanReject(id uint, req *dto.ApproveReq, operator string) (*model.PurchaseRequest, error) {
	var updated *model.PurchaseRequest
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		p, err := s.repo.FindByIDForUpdate(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "采购申请不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		if p.Status != constants.PurchaseStatusPendingDean {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		p.Status = constants.PurchaseStatusRejected
		p.Dean = operator
		p.DeanComment = req.Comment
		now := time.Now()
		p.DeanAt = &now
		if err := s.repo.UpdateTx(tx, p); err != nil {
			return err
		}
		updated = p
		return nil
	})
	if err != nil {
		return nil, s.wrapStatus(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogPurchaseDeanReject, updated.RequestNo, operator, req.Comment))
	s.audit.Record(0, operator, "REJECT", "purchase", util.Uint64String(updated.ID), "院长驳回: "+updated.RequestNo, operator, "")
	return updated, nil
}

// Deliver 到货登记 → 待验收。
func (s *PurchaseService) Deliver(id uint, operator string) (*model.PurchaseRequest, error) {
	var updated *model.PurchaseRequest
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		p, err := s.repo.FindByIDForUpdate(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "采购申请不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		if p.Status != constants.PurchaseStatusApproved {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		p.Status = constants.PurchaseStatusDelivered
		now := time.Now()
		p.DeliveredAt = &now
		if err := s.repo.UpdateTx(tx, p); err != nil {
			return err
		}
		updated = p
		return nil
	})
	if err != nil {
		return nil, s.wrapStatus(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogPurchaseDelivered, updated.RequestNo, updated.Status))
	s.audit.Record(0, operator, "DELIVER", "purchase", util.Uint64String(updated.ID), "到货登记: "+updated.RequestNo, operator, "")
	return updated, nil
}

// Accept 验收登记 → 正式入台账并生成分发条码（多步写操作在事务中完成）。
func (s *PurchaseService) Accept(id uint, req *dto.AcceptReq, operator string) (*model.Device, error) {
	var device *model.Device
	requestNo := ""
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		p, err := s.repo.FindByIDForUpdate(tx, id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "采购申请不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		if p.Status != constants.PurchaseStatusDelivered {
			return util.NewAppError(http.StatusConflict, constants.MsgInvalidStatus, nil)
		}
		if _, err := s.device.FindByAssetCode(req.AssetCode); err == nil {
			return util.NewAppError(http.StatusConflict, constants.MsgDuplicateAssetCode, nil)
		}
		requestNo = p.RequestNo
		acceptanceDate := req.AcceptanceDate
		if acceptanceDate == nil {
			now := time.Now()
			acceptanceDate = &now
		}
		p.Status = constants.PurchaseStatusAccepted
		p.AcceptancePerson = req.AcceptancePerson
		p.AcceptanceDate = acceptanceDate
		p.PartsList = req.PartsList
		p.CertificateNo = req.CertificateNo
		p.RegistrationNo = req.RegistrationNo
		if err := s.repo.UpdateTx(tx, p); err != nil {
			return err
		}
		barcode := util.GenBarcode(req.AssetCode)
		device = &model.Device{
			AssetCode:           req.AssetCode,
			Barcode:             barcode,
			Name:                p.DeviceName,
			Model:               p.Model,
			Manufacturer:        p.Manufacturer,
			Category:            req.Category,
			Department:          p.Department,
			ResponsiblePerson:   req.ResponsiblePerson,
			Location:            req.Location,
			Supplier:            p.Manufacturer,
			PurchaseDate:        acceptanceDate,
			PurchaseAmount:      p.BudgetAmount,
			WarrantyMonths:      req.WarrantyMonths,
			RegistrationNo:      req.RegistrationNo,
			CertificateNo:       req.CertificateNo,
			Status:              constants.DeviceStatusInStorage,
			CalibrationRequired: req.CalibrationRequired,
			PurchaseRequestID:   p.ID,
		}
		if device.WarrantyMonths > 0 {
			expiry := acceptanceDate.AddDate(0, device.WarrantyMonths, 0)
			device.WarrantyExpiry = &expiry
		}
		if err := tx.Create(device).Error; err != nil {
			return err
		}
		p.DeviceID = device.ID
		if err := s.repo.UpdateTx(tx, p); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, s.wrapStatus(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogPurchaseAccepted, requestNo, device.ID, device.AssetCode, device.Barcode))
	s.audit.Record(0, operator, "ACCEPT", "purchase", util.Uint64String(device.PurchaseRequestID), "验收通过并生成设备: "+device.Name, operator, "")
	return device, nil
}

// wrapStatus 统一包装业务错误。
func (s *PurchaseService) wrapStatus(err error) error {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
}
