package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/handler"
)

// registerDeviceRoutes 设备台账路由（登录用户可访问，写操作按角色限制）。
func registerDeviceRoutes(g *gin.RouterGroup, h *handler.DeviceHandler) {
	devices := g.Group("/devices")
	{
		devices.GET("", h.List)
		devices.GET("/:id", h.Get)
		devices.POST("", h.Create)
		devices.PUT("/:id", h.Update)
		devices.POST("/:id/disable", h.Disable)
		devices.POST("/:id/enable", h.Enable)
	}
}
