package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/handler"
	"github.com/medasset/medasset/internal/middleware"
)

// registerAuditRoutes 审计日志路由（设备科及以上）。
func registerAuditRoutes(g *gin.RouterGroup, h *handler.AuditHandler) {
	g.GET("/audits", middleware.RequireRoles(constants.RoleSuperAdmin, constants.RoleDeviceAdmin), h.List)
}
