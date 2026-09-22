package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// StatsHandler 资产统计处理器。
type StatsHandler struct {
	svc *service.StatsService
}

func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

// Overview 资产总览。
func (h *StatsHandler) Overview(c *gin.Context) {
	data, err := h.svc.Overview()
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, data)
}
