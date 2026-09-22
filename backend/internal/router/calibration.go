package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/handler"
)

// registerCalibrationRoutes 计量与质控路由。
func registerCalibrationRoutes(g *gin.RouterGroup, h *handler.CalibrationHandler) {
	cal := g.Group("/calibrations")
	{
		cal.GET("", h.List)
		cal.POST("", h.Create)
		cal.GET("/due", h.DueList)
		cal.POST("/:id/result", h.RecordResult)
	}
}
