package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/util"
)

// Audit 审计日志中间件：记录写操作与耗时。
func Audit(repo *repository.AuditRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		// 只审计写操作。
		switch c.Request.Method {
		case "POST", "PUT", "DELETE", "PATCH":
		default:
			return
		}
		if c.Writer.Status() >= 400 {
			return
		}
		claims := CurrentUser(c)
		username := "anonymous"
		var userID uint
		if claims != nil {
			username = claims.Username
			userID = claims.UserID
		}
		rid, _ := c.Get(RequestIDKey)
		log := &model.AuditLog{
			UserID:    userID,
			Username:  username,
			Action:    c.Request.Method,
			Module:    c.FullPath(),
			EntityID:  c.Param("id"),
			Detail:    fmt.Sprintf("%s %s 耗时 %dms", c.Request.Method, c.Request.URL.Path, time.Since(start).Milliseconds()),
			IP:        c.ClientIP(),
			RequestID: fmt.Sprintf("%v", rid),
		}
		_ = repo.Create(log)
		util.Log.Info("请求处理完成", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "request_id", rid)
	}
}
