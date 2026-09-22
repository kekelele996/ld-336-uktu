package router

import (
	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/handler"
)

// registerPublicAuthRoutes 公开认证路由（注册/登录）。
func registerPublicAuthRoutes(g *gin.RouterGroup, h *handler.AuthHandler) {
	g.POST("/auth/register", h.Register)
	g.POST("/auth/login", h.Login)
}

// registerProtectedAuthRoutes 需登录的认证路由。
func registerProtectedAuthRoutes(g *gin.RouterGroup, h *handler.AuthHandler) {
	g.GET("/auth/me", h.Me)
}
