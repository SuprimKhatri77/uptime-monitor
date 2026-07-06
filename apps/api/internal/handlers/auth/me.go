package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/packages/handlerlog"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/repository"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/types"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
)

func Me(queries repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userIDFromContext := c.MustGet("user_id").(string)

		userID, err := utils.ConvertToUUID(userIDFromContext)
		if err != nil {
			handlerlog.Error(c, "invalid user_id in context", err, "user_id", userIDFromContext)

			c.JSON(http.StatusBadRequest, types.Error("Invalid user ID format", constants.ValidationFailed))
			return
		}

		user, err := queries.GetUserByID(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				handlerlog.Warn(c, "user not found", "user_id", userID)

				c.JSON(http.StatusNotFound, types.Error("User not found", constants.UserNotFound))
				return
			}

			handlerlog.Error(c, "failed to fetch user", err, "user_id", userID)

			c.JSON(http.StatusInternalServerError, types.Error("Failed to fetch user", constants.InternalServerError))
			return
		}

		handlerlog.Info(c, "fetched current user", "user_id", userID, "role", user.Role)

		c.JSON(http.StatusOK, types.Success("Valid session", user))
	}
}
