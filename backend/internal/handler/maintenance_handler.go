package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/middleware"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// MaintenanceHandler 维护保养/维修处理器。
type MaintenanceHandler struct {
	svc *service.MaintenanceService
}

func NewMaintenanceHandler(svc *service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{svc: svc}
}

// List 保养/维修记录列表。
func (h *MaintenanceHandler) List(c *gin.Context) {
	page := util.ParsePage(c.Query("page"))
	pageSize := util.ParsePageSize(c.Query("page_size"))
	deviceID := uint(util.ParsePage(c.Query("device_id")))
	if deviceID == 1 && c.Query("device_id") == "" {
		deviceID = 0
	}
	mType := c.Query("type")
	status := c.Query("status")
	result, err := h.svc.List(page, pageSize, deviceID, mType, status)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Create 创建工单。
func (h *MaintenanceHandler) Create(c *gin.Context) {
	var req dto.CreateMaintenanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "工单参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	m, err := h.svc.Create(&req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, m)
}

// GeneratePlans 自动生成保养计划。
func (h *MaintenanceHandler) GeneratePlans(c *gin.Context) {
	operator := middleware.CurrentUser(c)
	n, err := h.svc.GeneratePlans(operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"created": n})
}

// Start 开始执行工单。
func (h *MaintenanceHandler) Start(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "工单ID不合法", err))
		return
	}
	var req dto.StartMaintenanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "工单参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	m, err := h.svc.Start(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, m)
}

// Complete 完成工单。
func (h *MaintenanceHandler) Complete(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "工单ID不合法", err))
		return
	}
	var req dto.CompleteMaintenanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "工单参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	m, err := h.svc.Complete(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, m)
}

// Cancel 取消工单。
func (h *MaintenanceHandler) Cancel(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "工单ID不合法", err))
		return
	}
	var req dto.CancelMaintenanceReq
	_ = c.ShouldBindJSON(&req)
	operator := middleware.CurrentUser(c)
	m, err := h.svc.Cancel(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, m)
}
