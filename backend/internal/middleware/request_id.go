package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey 请求 ID 上下文键。
const RequestIDKey = "request_id"

// RequestID 为每个请求注入唯一 request id，并回写响应头。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(RequestIDKey, rid)
		c.Header("X-Request-Id", rid)
		c.Next()
	}
}
