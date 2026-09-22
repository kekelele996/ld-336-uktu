package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// AuditHandler 审计日志处理器。
type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

// List 审计日志列表。
func (h *AuditHandler) List(c *gin.Context) {
	page := util.ParsePage(c.Query("page"))
	pageSize := util.ParsePageSize(c.Query("page_size"))
	module := c.Query("module")
	username := c.Query("username")
	result, err := h.svc.List(page, pageSize, module, username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}
