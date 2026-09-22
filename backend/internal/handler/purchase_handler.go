package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/middleware"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// PurchaseHandler 采购与验收处理器。
type PurchaseHandler struct {
	svc *service.PurchaseService
}

func NewPurchaseHandler(svc *service.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{svc: svc}
}

// List 采购申请列表。
func (h *PurchaseHandler) List(c *gin.Context) {
	page := util.ParsePage(c.Query("page"))
	pageSize := util.ParsePageSize(c.Query("page_size"))
	status := c.Query("status")
	department := c.Query("department")
	result, err := h.svc.List(page, pageSize, status, department)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Get 采购申请详情。
func (h *PurchaseHandler) Get(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "采购申请ID不合法", err))
		return
	}
	req, err := h.svc.Get(p.ID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, req)
}

// Create 提交采购申请。
func (h *PurchaseHandler) Create(c *gin.Context) {
	var req dto.CreatePurchaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "采购申请参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	p, err := h.svc.Create(&req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, p)
}

// DeviceAdminApprove 设备科审核通过。
func (h *PurchaseHandler) DeviceAdminApprove(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "采购申请ID不合法", err))
		return
	}
	var req dto.ApproveReq
	_ = c.ShouldBindJSON(&req)
	operator := middleware.CurrentUser(c)
	r, err := h.svc.DeviceAdminApprove(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, r)
}

// DeviceAdminReject 设备科驳回。
func (h *PurchaseHandler) DeviceAdminReject(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "采购申请ID不合法", err))
		return
	}
	var req dto.ApproveReq
	_ = c.ShouldBindJSON(&req)
	operator := middleware.CurrentUser(c)
	r, err := h.svc.DeviceAdminReject(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, r)
}

// DeanApprove 院长审批通过。
func (h *PurchaseHandler) DeanApprove(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "采购申请ID不合法", err))
		return
	}
	var req dto.ApproveReq
	_ = c.ShouldBindJSON(&req)
	operator := middleware.CurrentUser(c)
	r, err := h.svc.DeanApprove(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, r)
}

// DeanReject 院长驳回。
func (h *PurchaseHandler) DeanReject(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "采购申请ID不合法", err))
		return
	}
	var req dto.ApproveReq
	_ = c.ShouldBindJSON(&req)
	operator := middleware.CurrentUser(c)
	r, err := h.svc.DeanReject(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, r)
}

// Deliver 到货登记。
func (h *PurchaseHandler) Deliver(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "采购申请ID不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	r, err := h.svc.Deliver(p.ID, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, r)
}

// Accept 验收登记并入台账。
func (h *PurchaseHandler) Accept(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "采购申请ID不合法", err))
		return
	}
	var req dto.AcceptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "验收参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	device, err := h.svc.Accept(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, device)
}
