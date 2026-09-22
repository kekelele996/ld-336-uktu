package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/handler"
	"github.com/medasset/medasset/internal/middleware"
)

// registerPurchaseRoutes 采购与验收路由。
func registerPurchaseRoutes(g *gin.RouterGroup, h *handler.PurchaseHandler) {
	purchases := g.Group("/purchases")
	{
		purchases.GET("", h.List)
		purchases.GET("/:id", h.Get)
		purchases.POST("", h.Create)
	}
	// 设备科审核。
	admin := g.Group("/purchases", middleware.RequireRoles(constants.RoleDeviceAdmin, constants.RoleSuperAdmin))
	{
		admin.POST("/:id/device-admin-approve", h.DeviceAdminApprove)
		admin.POST("/:id/device-admin-reject", h.DeviceAdminReject)
		admin.POST("/:id/deliver", h.Deliver)
		admin.POST("/:id/accept", h.Accept)
	}
	// 院长审批。
	dean := g.Group("/purchases", middleware.RequireRoles(constants.RoleDean, constants.RoleSuperAdmin))
	{
		dean.POST("/:id/dean-approve", h.DeanApprove)
		dean.POST("/:id/dean-reject", h.DeanReject)
	}
}
