package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/middleware"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// AuthHandler 认证处理器。
type AuthHandler struct {
	svc *service.UserService
}

func NewAuthHandler(svc *service.UserService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register 注册。
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "注册参数不合法", err))
		return
	}
	user, err := h.svc.Register(&req, c.ClientIP())
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, user)
}

// Login 登录。
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "登录参数不合法", err))
		return
	}
	resp, err := h.svc.Login(&req, c.ClientIP())
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, resp)
}

// Me 当前用户信息。
func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.CurrentUser(c)
	if claims == nil {
		util.Fail(c, http.StatusUnauthorized, 40100, "未登录")
		return
	}
	user, err := h.svc.Me(claims.UserID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, user)
}
