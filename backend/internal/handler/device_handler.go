package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/middleware"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// DeviceHandler 设备台账处理器。
type DeviceHandler struct {
	svc *service.DeviceService
}

func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// List 设备列表（支持科室/类型/状态/关键字检索）。
func (h *DeviceHandler) List(c *gin.Context) {
	page := util.ParsePage(c.Query("page"))
	pageSize := util.ParsePageSize(c.Query("page_size"))
	department := c.Query("department")
	category := c.Query("category")
	status := c.Query("status")
	keyword := c.Query("keyword")
	result, err := h.svc.List(page, pageSize, department, category, status, keyword)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Get 设备详情。
func (h *DeviceHandler) Get(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "设备ID不合法", err))
		return
	}
	device, err := h.svc.Get(p.ID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, device)
}

// Create 创建设备。
func (h *DeviceHandler) Create(c *gin.Context) {
	var req dto.CreateDeviceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "设备参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	device, err := h.svc.Create(&req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, device)
}

// Update 更新设备。
func (h *DeviceHandler) Update(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "设备ID不合法", err))
		return
	}
	var req dto.UpdateDeviceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "设备参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	device, err := h.svc.Update(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, device)
}

// Disable 禁用设备。
func (h *DeviceHandler) Disable(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "设备ID不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	device, err := h.svc.Disable(p.ID, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, device)
}

// Enable 启用设备。
func (h *DeviceHandler) Enable(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "设备ID不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	device, err := h.svc.Enable(p.ID, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, device)
}
