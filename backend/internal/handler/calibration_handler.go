package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/middleware"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// CalibrationHandler 计量与质控处理器。
type CalibrationHandler struct {
	svc *service.CalibrationService
}

func NewCalibrationHandler(svc *service.CalibrationService) *CalibrationHandler {
	return &CalibrationHandler{svc: svc}
}

// List 计量记录列表。
func (h *CalibrationHandler) List(c *gin.Context) {
	page := util.ParsePage(c.Query("page"))
	pageSize := util.ParsePageSize(c.Query("page_size"))
	deviceID := uint(util.ParsePage(c.Query("device_id")))
	if deviceID == 1 && c.Query("device_id") == "" {
		deviceID = 0
	}
	status := c.Query("status")
	result, err := h.svc.List(page, pageSize, deviceID, status)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Create 建立计量台账。
func (h *CalibrationHandler) Create(c *gin.Context) {
	var req dto.CreateCalibrationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "计量参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	cal, err := h.svc.Create(&req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, cal)
}

// DueList 计量到期预警清单。
func (h *CalibrationHandler) DueList(c *gin.Context) {
	list, err := h.svc.DueList()
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, list)
}

// RecordResult 登记计量结果。
func (h *CalibrationHandler) RecordResult(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "计量记录ID不合法", err))
		return
	}
	var req dto.CalibrationResultReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "计量参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	cal, err := h.svc.RecordResult(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, cal)
}
