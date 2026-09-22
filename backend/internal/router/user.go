package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/handler"
	"github.com/medasset/medasset/internal/middleware"
)

// registerUserRoutes 用户管理路由（系统管理员）。
func registerUserRoutes(g *gin.RouterGroup, h *handler.UserHandler) {
	users := g.Group("/users", middleware.RequireRoles(constants.RoleSuperAdmin))
	{
		users.GET("", h.List)
		users.GET("/:id", h.Get)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
	}
}
