package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/middleware"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// TransferHandler 设备调拨处理器。
type TransferHandler struct {
	svc *service.TransferService
}

func NewTransferHandler(svc *service.TransferService) *TransferHandler {
	return &TransferHandler{svc: svc}
}

// List 调拨申请列表。
func (h *TransferHandler) List(c *gin.Context) {
	page := util.ParsePage(c.Query("page"))
	pageSize := util.ParsePageSize(c.Query("page_size"))
	status := c.Query("status")
	result, err := h.svc.List(page, pageSize, status)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Create 发起调拨申请。
func (h *TransferHandler) Create(c *gin.Context) {
	var req dto.CreateTransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "调拨参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	t, err := h.svc.Create(&req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, t)
}

// Approve 审批通过。
func (h *TransferHandler) Approve(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "调拨申请ID不合法", err))
		return
	}
	var req dto.TransferApproveReq
	_ = c.ShouldBindJSON(&req)
	operator := middleware.CurrentUser(c)
	t, err := h.svc.Approve(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, t)
}

// Reject 驳回。
func (h *TransferHandler) Reject(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "调拨申请ID不合法", err))
		return
	}
	var req dto.TransferApproveReq
	_ = c.ShouldBindJSON(&req)
	operator := middleware.CurrentUser(c)
	t, err := h.svc.Reject(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, t)
}
