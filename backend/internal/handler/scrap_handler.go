package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/middleware"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// ScrapHandler 设备报废处理器。
type ScrapHandler struct {
	svc *service.ScrapService
}

func NewScrapHandler(svc *service.ScrapService) *ScrapHandler {
	return &ScrapHandler{svc: svc}
}

// List 报废申请列表。
func (h *ScrapHandler) List(c *gin.Context) {
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

// Create 发起报废申请。
func (h *ScrapHandler) Create(c *gin.Context) {
	var req dto.CreateScrapReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "报废参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	s, err := h.svc.Create(&req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, s)
}

// Approve 审批通过。
func (h *ScrapHandler) Approve(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "报废申请ID不合法", err))
		return
	}
	var req dto.ScrapApproveReq
	_ = c.ShouldBindJSON(&req)
	operator := middleware.CurrentUser(c)
	s, err := h.svc.Approve(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, s)
}

// Reject 驳回。
func (h *ScrapHandler) Reject(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "报废申请ID不合法", err))
		return
	}
	var req dto.ScrapApproveReq
	_ = c.ShouldBindJSON(&req)
	operator := middleware.CurrentUser(c)
	s, err := h.svc.Reject(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, s)
}
