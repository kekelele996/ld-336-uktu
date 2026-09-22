package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/util"
)

// ErrorHandler 全局错误恢复与统一错误响应。
func ErrorHandler(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				rid, _ := c.Get(RequestIDKey)
				log.Error(fmt.Sprintf(constants.LogPanicRecovered, c.Request.URL.Path, rid, rec))
				util.Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
				c.Abort()
			}
		}()
		c.Next()
		for _, err := range c.Errors {
			var appErr *util.AppError
			if errors.As(err.Err, &appErr) {
				util.Fail(c, httpStatusOf(appErr.Code), appErr.Code, appErr.Message)
				continue
			}
			var verrs validator.ValidationErrors
			if errors.As(err.Err, &verrs) {
				msg := "参数校验失败: " + verrs[0].Field() + " " + verrs[0].Tag()
				util.Fail(c, http.StatusBadRequest, constants.CodeValidation, msg)
				continue
			}
			util.Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
		}
	}
}

func httpStatusOf(code int) int {
	switch code {
	case constants.CodeUnauthorized, constants.CodeInvalidToken, constants.CodeTokenExpired, constants.CodeWrongPassword:
		return http.StatusUnauthorized
	case constants.CodeForbidden, constants.CodeUserDisabled:
		return http.StatusForbidden
	case constants.CodeNotFound:
		return http.StatusNotFound
	case constants.CodeConflict, constants.CodeInvalidStatus, constants.CodeDeviceNotAllowed:
		return http.StatusConflict
	case constants.CodeValidation, constants.CodeBadRequest:
		return http.StatusBadRequest
	case constants.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
