package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/constants"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/respond"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.MustGet("role").(string)

		slog.Info("role from context", " ", role)

		for _, r := range roles {
			if role == r {
				slog.Info("valid role")
				c.Next()
				return
			}
		}

		slog.Info("invalid role")

		respond.ForbiddenAbort(c, "Insufficient permissions", constants.Forbidden)
	}
}
