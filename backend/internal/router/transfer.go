package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/handler"
	"github.com/medasset/medasset/internal/middleware"
)

// registerTransferRoutes 设备调拨路由。
func registerTransferRoutes(g *gin.RouterGroup, h *handler.TransferHandler) {
	tr := g.Group("/transfers")
	{
		tr.GET("", h.List)
		tr.POST("", h.Create)
	}
	approve := g.Group("/transfers", middleware.RequireRoles(constants.RoleDeviceAdmin, constants.RoleSuperAdmin, constants.RoleDean))
	{
		approve.POST("/:id/approve", h.Approve)
		approve.POST("/:id/reject", h.Reject)
	}
}
