package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Resp 统一响应结构：{ "code": 0, "message": "ok", "data": ... }
type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Resp{Code: 0, Message: "ok", Data: data})
}

// Fail 返回业务错误响应。
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, Resp{Code: code, Message: message})
}

// PageResult 分页返回结构。
type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
