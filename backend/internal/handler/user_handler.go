package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/middleware"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
)

// UserHandler 用户管理处理器。
type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// List 用户列表。
func (h *UserHandler) List(c *gin.Context) {
	page := util.ParsePage(c.Query("page"))
	pageSize := util.ParsePageSize(c.Query("page_size"))
	role := c.Query("role")
	keyword := c.Query("keyword")
	result, err := h.svc.List(page, pageSize, role, keyword)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Get 用户详情。
func (h *UserHandler) Get(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "用户ID不合法", err))
		return
	}
	user, err := h.svc.Me(p.ID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, user)
}

// Update 更新用户。
func (h *UserHandler) Update(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "用户ID不合法", err))
		return
	}
	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "用户参数不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	user, err := h.svc.Update(p.ID, &req, operator.Username)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, user)
}

// Delete 删除用户。
func (h *UserHandler) Delete(c *gin.Context) {
	var p dto.IDParam
	if err := c.ShouldBindUri(&p); err != nil {
		c.Error(util.NewAppError(http.StatusBadRequest, "用户ID不合法", err))
		return
	}
	operator := middleware.CurrentUser(c)
	if err := h.svc.Delete(p.ID, operator.Username); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"id": p.ID})
}
