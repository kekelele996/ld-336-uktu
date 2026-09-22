package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/handler"
	"github.com/medasset/medasset/internal/middleware"
)

// registerMaintenanceRoutes 维护保养/维修路由。
func registerMaintenanceRoutes(g *gin.RouterGroup, h *handler.MaintenanceHandler) {
	mt := g.Group("/maintenances")
	{
		mt.GET("", h.List)
		mt.POST("", h.Create)
		mt.POST("/plan/generate", middleware.RequireRoles(constants.RoleDeviceAdmin, constants.RoleSuperAdmin, constants.RoleEngineer), h.GeneratePlans)
		mt.POST("/:id/start", h.Start)
		mt.POST("/:id/complete", h.Complete)
		mt.POST("/:id/cancel", h.Cancel)
	}
}
