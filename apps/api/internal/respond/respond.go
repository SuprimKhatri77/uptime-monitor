package respond

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/types"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/validator"
)

func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, types.Success(message, data))
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, types.Success(message, data))
}

func OKMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, types.Success(message, nil))
}

func OKList(c *gin.Context, message string, data any, meta types.PaginationMeta) {
	c.JSON(http.StatusOK, types.SuccessWithMeta(message, data, meta))
}

func Validation(c *gin.Context, message string, errors []types.AppError) {
	c.JSON(http.StatusBadRequest, types.ValidationErrorResponse(message, errors))
}

func BindJSON(c *gin.Context, obj any, message string) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		Validation(c, message, validator.Parse(err, obj))
		return false
	}
	return true
}

func BadRequest(c *gin.Context, message, code string) {
	c.JSON(http.StatusBadRequest, types.Error(message, code))
}

func Unauthorized(c *gin.Context, message, code string) {
	c.JSON(http.StatusUnauthorized, types.Error(message, code))
}

func UnauthorizedAbort(c *gin.Context, message, code string) {
	Unauthorized(c, message, code)
	c.Abort()
}

func Forbidden(c *gin.Context, message, code string) {
	c.JSON(http.StatusForbidden, types.Error(message, code))
}

func ForbiddenAbort(c *gin.Context, message, code string) {
	Forbidden(c, message, code)
	c.Abort()
}

func NotFound(c *gin.Context, message, code string) {
	c.JSON(http.StatusNotFound, types.Error(message, code))
}

func Conflict(c *gin.Context, message, code string) {
	c.JSON(http.StatusConflict, types.Error(message, code))
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, types.Error(message, constants.InternalServerError))
}
