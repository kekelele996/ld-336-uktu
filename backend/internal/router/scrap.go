package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/handler"
	"github.com/medasset/medasset/internal/middleware"
)

// registerScrapRoutes 设备报废路由。
func registerScrapRoutes(g *gin.RouterGroup, h *handler.ScrapHandler) {
	sc := g.Group("/scraps")
	{
		sc.GET("", h.List)
		sc.POST("", h.Create)
	}
	approve := g.Group("/scraps", middleware.RequireRoles(constants.RoleDeviceAdmin, constants.RoleSuperAdmin, constants.RoleDean))
	{
		approve.POST("/:id/approve", h.Approve)
		approve.POST("/:id/reject", h.Reject)
	}
}
