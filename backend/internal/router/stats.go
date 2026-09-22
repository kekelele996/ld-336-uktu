package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/handler"
)

// registerStatsRoutes 统计报表路由。
func registerStatsRoutes(g *gin.RouterGroup, h *handler.StatsHandler) {
	g.GET("/stats/overview", h.Overview)
}
