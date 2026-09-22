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

// CalibrationService 计量与质控服务。
type CalibrationService struct {
	repo   *repository.CalibrationRepository
	device *repository.DeviceRepository
	audit  *AuditService
	log    *slog.Logger
}

func NewCalibrationService(repo *repository.CalibrationRepository, device *repository.DeviceRepository, audit *AuditService, log *slog.Logger) *CalibrationService {
	return &CalibrationService{repo: repo, device: device, audit: audit, log: log}
}

// Create 建立计量台账。
func (s *CalibrationService) Create(req *dto.CreateCalibrationReq, operator string) (*model.CalibrationRecord, error) {
	taken, err := s.repo.IsInstrumentNoTaken(req.InstrumentNo)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	if taken {
		return nil, util.NewAppError(http.StatusConflict, constants.MsgDuplicateInstrumentNo, nil)
	}
	d, err := s.device.FindByID(req.DeviceID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, util.NewAppError(http.StatusNotFound, "设备不存在: device_id="+util.Uint64String(req.DeviceID), nil)
	}
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	now := time.Now()
	last := req.LastCalibrationDate
	if last == nil {
		last = &now
	}
	next := last.AddDate(0, req.CalibrationCycleMonths, 0)
	c := &model.CalibrationRecord{
		InstrumentNo:            req.InstrumentNo,
		DeviceID:                d.ID,
		DeviceName:              d.Name,
		CalibrationCycleMonths:  req.CalibrationCycleMonths,
		LastCalibrationDate:     last,
		NextCalibrationDate:     &next,
		Status:                  constants.CalibrationStatusNormal,
		Result:                  constants.CalibrationResultQualified,
		CertificateNo:           req.CertificateNo,
		CalibrationOrg:          req.CalibrationOrg,
		Remark:                  req.Remark,
		CreatedBy:               operator,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "建立计量台账失败: instrument_no="+req.InstrumentNo, err)
	}
	s.log.Info(fmt.Sprintf(constants.LogCalibrationCreated, c.InstrumentNo, c.DeviceID, c.CalibrationCycleMonths, c.Status))
	s.audit.Record(0, operator, "CREATE", "calibration", util.Uint64String(c.ID), "建立计量台账: "+c.InstrumentNo, operator, "")
	return c, nil
}

// List 分页查询计量记录。
func (s *CalibrationService) List(page, pageSize int, deviceID uint, status string) (*util.PageResult, error) {
	list, total, err := s.repo.List(page, pageSize, deviceID, status)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	return &util.PageResult{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// DueList 计量到期预警清单（被列表页与统计接口复用）。
func (s *CalibrationService) DueList() ([]model.CalibrationRecord, error) {
	list, err := s.repo.ListDue()
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	return list, nil
}

// RecordResult 登记计量结果；不合格自动标记设备禁用。
func (s *CalibrationService) RecordResult(id uint, req *dto.CalibrationResultReq, operator string) (*model.CalibrationRecord, error) {
	var updated *model.CalibrationRecord
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		c, err := s.repo.FindByID(id)
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(http.StatusNotFound, "计量记录不存在: id="+util.Uint64String(id), nil)
		}
		if err != nil {
			return err
		}
		now := time.Now()
		next := req.NextCalibrationDate
		if next == nil {
			nextDate := now.AddDate(0, c.CalibrationCycleMonths, 0)
			next = &nextDate
		}
		c.NextCalibrationDate = next
		c.LastCalibrationDate = &now
		c.CertificateNo = req.CertificateNo
		c.CalibrationOrg = req.CalibrationOrg
		c.Remark = req.Remark
		if req.Result == constants.CalibrationResultQualified {
			c.Status = constants.CalibrationStatusNormal
			c.Result = constants.CalibrationResultQualified
		} else {
			c.Status = constants.CalibrationStatusUnqualified
			c.Result = constants.CalibrationResultUnqualified
			// 不合格设备自动标记禁用。
			if err := s.device.UpdateStatusTx(tx, c.DeviceID, constants.DeviceStatusDisabled); err != nil {
				return err
			}
		}
		if err := s.repo.UpdateTx(tx, c); err != nil {
			return err
		}
		updated = c
		return nil
	})
	if err != nil {
		return nil, wrapSvcErr(err)
	}
	s.log.Info(fmt.Sprintf(constants.LogCalibrationResult, updated.InstrumentNo, updated.Result, updated.Status, updated.DeviceID))
	s.audit.Record(0, operator, "RESULT", "calibration", util.Uint64String(updated.ID), "登记计量结果: "+updated.InstrumentNo+"="+updated.Result, operator, "")
	return updated, nil
}
