package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/util"
)

// RequireRoles RBAC 权限中间件：仅允许指定角色访问。
func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		claims := CurrentUser(c)
		if claims == nil {
			util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		if !allowed[claims.Role] {
			util.Fail(c, http.StatusForbidden, constants.CodeForbidden, "无权访问该资源: role="+claims.Role)
			c.Abort()
			return
		}
		c.Next()
	}
}
