package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	db "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/packages/handlerlog"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/repository"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/respond"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/utils"
)

func Me(queries repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userIDFromContext := c.MustGet("userID").(string)

		userID, err := utils.ConvertToUUID(userIDFromContext)
		if err != nil {
			handlerlog.Error(c, "invalid user_id in context", err, "user_id", userIDFromContext)

			respond.BadRequest(c, "Invalid user ID format", constants.ValidationFailed)
			return
		}

		user, err := queries.GetUserByID(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				handlerlog.Warn(c, "user not found", "user_id", userID)

				respond.NotFound(c, "User not found", constants.UserNotFound)
				return
			}

			handlerlog.Error(c, "failed to fetch user", err, "user_id", userID)

			respond.InternalError(c, "Failed to fetch user")
			return
		}

		handlerlog.Info(c, "fetched current user", "user_id", userID, "role", user.Role)

		respond.OK(c, "Valid session", db.User{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			Role:     user.Role,
			ImageUrl: user.ImageUrl,
		})
	}
}
